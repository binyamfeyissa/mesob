package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	"github.com/mesob-wallet/loans/internal/infra/config"
	"github.com/rs/zerolog"
)

var log zerolog.Logger

func main() {
	cfg := config.Load()
	log = kitlogging.New("loans-worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil || pool.Ping(ctx) != nil {
		log.Warn().Msg("loans-worker: postgres unavailable, running in no-op mode")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	notifyURL := getenv("MESOB_NOTIFICATION_URL", "http://notification:8012")

	go runDueCron(ctx, pool, notifyURL)
	go runOverdueCron(ctx, pool, notifyURL)

	log.Info().Msg("loans worker started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	log.Info().Msg("loans worker stopped")
}

func runDueCron(ctx context.Context, pool *pgxpool.Pool, notifyURL string) {
	// Run once at startup, then daily.
	processDueReminders(ctx, pool, notifyURL)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processDueReminders(ctx, pool, notifyURL)
		}
	}
}

func runOverdueCron(ctx context.Context, pool *pgxpool.Pool, notifyURL string) {
	processOverdueLoans(ctx, pool, notifyURL)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processOverdueLoans(ctx, pool, notifyURL)
		}
	}
}

func processDueReminders(ctx context.Context, pool *pgxpool.Pool, notifyURL string) {
	if pool == nil {
		return
	}
	tomorrow := time.Now().UTC().Add(24 * time.Hour).Truncate(24 * time.Hour)
	rows, err := pool.Query(ctx,
		`SELECT id, user_id, principal_minor FROM loans
		 WHERE status = 'ACTIVE' AND due_date::date = $1::date`,
		tomorrow,
	)
	if err != nil {
		log.Error().Err(err).Msg("due-reminders: query failed")
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var loanID, userID string
		var principalMinor int64
		if err := rows.Scan(&loanID, &userID, &principalMinor); err != nil {
			continue
		}
		sendNotification(ctx, notifyURL, userID, "loan.due_tomorrow", map[string]string{
			"loan_id":         loanID,
			"amount_etb":      fmt.Sprintf("%.2f", float64(principalMinor)/100),
			"due_date":        tomorrow.Format("2006-01-02"),
		})
		count++
	}
	if count > 0 {
		log.Info().Int("count", count).Msg("due reminders sent")
	}
}

func processOverdueLoans(ctx context.Context, pool *pgxpool.Pool, notifyURL string) {
	if pool == nil {
		return
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	rows, err := pool.Query(ctx,
		`SELECT id, user_id, principal_minor FROM loans
		 WHERE status = 'ACTIVE' AND due_date < $1`,
		today,
	)
	if err != nil {
		log.Error().Err(err).Msg("overdue: query failed")
		return
	}
	defer rows.Close()

	var ids []string
	type overdueRow struct {
		id, userID    string
		principalMinor int64
	}
	var overdue []overdueRow
	for rows.Next() {
		var r overdueRow
		if err := rows.Scan(&r.id, &r.userID, &r.principalMinor); err != nil {
			continue
		}
		ids = append(ids, r.id)
		overdue = append(overdue, r)
	}
	rows.Close()

	if len(ids) == 0 {
		return
	}

	_, err = pool.Exec(ctx,
		`UPDATE loans SET status = 'OVERDUE' WHERE id = ANY($1::uuid[]) AND status = 'ACTIVE'`,
		ids,
	)
	if err != nil {
		log.Error().Err(err).Msg("overdue: update failed")
		return
	}

	for _, r := range overdue {
		sendNotification(ctx, notifyURL, r.userID, "loan.overdue", map[string]string{
			"loan_id":    r.id,
			"amount_etb": fmt.Sprintf("%.2f", float64(r.principalMinor)/100),
		})
	}
	log.Info().Int("count", len(ids)).Msg("loans marked overdue")
}

func sendNotification(ctx context.Context, notifyURL, userID, templateKey string, params map[string]string) {
	body, _ := json.Marshal(map[string]any{
		"user_id":  userID,
		"template": templateKey,
		"params":   params,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, notifyURL+"/notify/send", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Str("template", templateKey).Msg("notification send failed")
		return
	}
	resp.Body.Close()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
