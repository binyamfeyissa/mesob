package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
	"github.com/mesob-wallet/notification/internal/app"
	"github.com/mesob-wallet/notification/internal/infra/config"
	notifyfcm "github.com/mesob-wallet/notification/internal/infra/fcm"
	notifyhttp "github.com/mesob-wallet/notification/internal/infra/http"
	notifypostgres "github.com/mesob-wallet/notification/internal/infra/postgres"
	notifysms "github.com/mesob-wallet/notification/internal/infra/sms"
	notifytelegram "github.com/mesob-wallet/notification/internal/infra/telegram"
	httpTransport "github.com/mesob-wallet/notification/internal/transport/http"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("notification")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Warn().Err(err).Msg("notification: could not connect to postgres — running without DB")
	} else if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Warn().Err(pingErr).Msg("notification: postgres ping failed — running without DB")
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
	userResolver := notifyhttp.NewIdentityUserResolver(identityURL)

	sendUC := &app.SendUseCase{
		Templates:  templateRepo,
		Deliveries: deliveryRepo,
		SMS:        &notifysms.LoggerClient{},
		FCM:        &notifyfcm.LoggerClient{},
		Telegram:   &notifytelegram.LoggerClient{},
		Users:      userResolver,
	}

	handler := httpTransport.NewHandler(sendUC)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("notification starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("notification stopped")
}
