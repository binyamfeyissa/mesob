package main

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	"github.com/mesob-wallet/notification/internal/app"
	"github.com/mesob-wallet/notification/internal/infra/config"
	notifyfcm "github.com/mesob-wallet/notification/internal/infra/fcm"
	notifyhttp "github.com/mesob-wallet/notification/internal/infra/http"
	notifypostgres "github.com/mesob-wallet/notification/internal/infra/postgres"
	notifysms "github.com/mesob-wallet/notification/internal/infra/sms"
	notifytelegram "github.com/mesob-wallet/notification/internal/infra/telegram"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// topics we consume; each carries a "user_id" field in its JSON payload.
var subscribedTopics = []string{
	"ledger.transaction-posted",
	"identity.user-activated",
	"identity.kyc-tier-changed",
	"loans.disbursed",
	"loans.decisioned",
	"iqub.contribution-recorded",
	"iqub.contribution-missed",
	"iqub.cycle-closed",
	"iddir.premium-paid",
	"iddir.claim-settled",
	"agent.float-low",
}

func main() {
	cfg := config.Load()
	log := kitlogging.New("notification-worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil || pool.Ping(ctx) != nil {
		log.Warn().Msg("notification-worker: postgres unavailable")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	var templateRepo *notifypostgres.TemplateRepo
	var deliveryRepo *notifypostgres.DeliveryRepo
	if pool != nil {
		templateRepo = &notifypostgres.TemplateRepo{DB: pool}
		deliveryRepo = &notifypostgres.DeliveryRepo{DB: pool}
	}

	identityURL := os.Getenv("MESOB_IDENTITY_URL")
	if identityURL == "" {
		identityURL = "http://identity:8002"
	}

	sendUC := &app.SendUseCase{
		Templates:  templateRepo,
		Deliveries: deliveryRepo,
		SMS:        &notifysms.LoggerClient{},
		FCM:        &notifyfcm.LoggerClient{},
		Telegram:   &notifytelegram.LoggerClient{},
		Users:      notifyhttp.NewIdentityUserResolver(identityURL),
	}

	for _, topic := range subscribedTopics {
		go consumeTopic(ctx, cfg.KafkaBrokers, topic, sendUC, log)
	}

	log.Info().Str("brokers", cfg.KafkaBrokers).Msg("notification worker started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	log.Info().Msg("notification worker stopped")
}

func consumeTopic(ctx context.Context, brokers, topic string, send *app.SendUseCase, log zerolog.Logger) {
	addrs := strings.Split(brokers, ",")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     addrs,
		GroupID:     "notification-service",
		Topic:       topic,
		MinBytes:    1,
		MaxBytes:    1 << 20, // 1 MiB
		MaxWait:     2 * time.Second,
		StartOffset: kafka.LastOffset,
	})
	defer r.Close()

	log.Info().Str("topic", topic).Msg("consumer started")

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if isNetworkErr(err) {
				time.Sleep(5 * time.Second)
				continue
			}
			log.Warn().Err(err).Str("topic", topic).Msg("fetch error")
			time.Sleep(2 * time.Second)
			continue
		}

		if processErr := handleEvent(ctx, topic, msg.Value, send, log); processErr != nil {
			log.Warn().Err(processErr).Str("topic", topic).Msg("event processing failed — committing anyway")
		}

		if commitErr := r.CommitMessages(ctx, msg); commitErr != nil && !errors.Is(commitErr, context.Canceled) {
			log.Warn().Err(commitErr).Str("topic", topic).Msg("commit failed")
		}
	}
}

type domainEvent struct {
	EventID   string            `json:"event_id"`
	UserID    string            `json:"user_id"`
	Params    map[string]string `json:"params"`
	// Loan / iqub events carry amount directly.
	AmountMinor int64  `json:"amount_minor"`
	Decision    string `json:"decision"` // loans.decisioned: APPROVED | DECLINED
}

func handleEvent(ctx context.Context, topic string, payload []byte, send *app.SendUseCase, log zerolog.Logger) error {
	var ev domainEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}
	if ev.UserID == "" {
		return nil // no user to notify
	}

	// Map topic → notification template key.
	templateKey := topicToTemplate(topic, &ev)
	if templateKey == "" {
		return nil
	}

	params := ev.Params
	if params == nil {
		params = map[string]string{}
	}

	_, err := send.Execute(ctx, app.SendInput{
		UserID:      ev.UserID,
		TemplateKey: templateKey,
		Params:      params,
	})
	if err != nil {
		log.Warn().Err(err).Str("template", templateKey).Str("user", ev.UserID).Msg("notification send failed")
	}
	return nil
}

func topicToTemplate(topic string, ev *domainEvent) string {
	switch topic {
	case "ledger.transaction-posted":
		return "ledger.transaction_posted"
	case "identity.user-activated":
		return "identity.welcome"
	case "identity.kyc-tier-changed":
		return "identity.kyc_upgraded"
	case "loans.disbursed":
		return "loan.disbursed"
	case "loans.decisioned":
		if ev.Decision == "DECLINED" {
			return "loan.declined"
		}
		return "" // APPROVED gets separate disbursed event
	case "iqub.contribution-recorded":
		return "iqub.contribution_recorded"
	case "iqub.contribution-missed":
		return "iqub.contribution_missed"
	case "iqub.cycle-closed":
		return "iqub.cycle_closed"
	case "iddir.premium-paid":
		return "iddir.premium_paid"
	case "iddir.claim-settled":
		return "iddir.claim_settled"
	case "agent.float-low":
		return "agent.float_low"
	}
	return ""
}

func isNetworkErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "connection refused") ||
		strings.Contains(s, "no such host") ||
		strings.Contains(s, "dial") ||
		strings.Contains(s, "EOF")
}
