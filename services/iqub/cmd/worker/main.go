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
	"github.com/mesob-wallet/iqub/internal/infra/config"
	"github.com/rs/zerolog"
)

var log zerolog.Logger

func main() {
	cfg := config.Load()
	log = kitlogging.New("iqub-worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil || pool.Ping(ctx) != nil {
		log.Warn().Msg("iqub-worker: postgres unavailable, running in no-op mode")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	ledgerURL := getenv("MESOB_IQUB_LEDGER_URL", "http://ledger:8002")
	notifyURL := getenv("MESOB_NOTIFICATION_URL", "http://notification:8012")

	go runCycleScheduler(ctx, pool, ledgerURL, notifyURL)

	log.Info().Msg("iqub worker started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	log.Info().Msg("iqub worker stopped")
}

func runCycleScheduler(ctx context.Context, pool *pgxpool.Pool, ledgerURL, notifyURL string) {
	closeDueCycles(ctx, pool, ledgerURL, notifyURL)
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			closeDueCycles(ctx, pool, ledgerURL, notifyURL)
		}
	}
}

type cycleRow struct {
	cycleID       string
	groupID       string
	poolAccountID string
	recipientID   string // user_id of this cycle's payout recipient
	amountMinor   int64
	cycleNumber   int
}

func closeDueCycles(ctx context.Context, pool *pgxpool.Pool, ledgerURL, notifyURL string) {
	if pool == nil {
		return
	}
	now := time.Now().UTC()

	// Find OPEN cycles past their due date.
	rows, err := pool.Query(ctx, `
		SELECT c.id, c.group_id, g.pool_account_id, c.recipient_user_id,
		       g.contribution_minor * g.member_count AS payout_minor,
		       c.cycle_number
		FROM iqub_cycles c
		JOIN iqub_groups g ON g.id = c.group_id
		WHERE c.status = 'OPEN' AND c.due_date <= $1
	`, now)
	if err != nil {
		log.Error().Err(err).Msg("iqub-worker: cycle query failed")
		return
	}
	defer rows.Close()

	var cycles []cycleRow
	for rows.Next() {
		var cr cycleRow
		if err := rows.Scan(&cr.cycleID, &cr.groupID, &cr.poolAccountID,
			&cr.recipientID, &cr.amountMinor, &cr.cycleNumber); err != nil {
			continue
		}
		cycles = append(cycles, cr)
	}
	rows.Close()

	for _, c := range cycles {
		// Mark non-payers.
		markMissedContributions(ctx, pool, c.cycleID, c.groupID, notifyURL)

		// Post ledger: pool → recipient (idempotent via cycle_id as key).
		if c.poolAccountID != "" && c.recipientID != "" && c.amountMinor > 0 {
			postPayout(ctx, ledgerURL, c)
		}

		// Close the cycle.
		_, err := pool.Exec(ctx,
			`UPDATE iqub_cycles SET status = 'CLOSED', closed_at = $1 WHERE id = $2`,
			now, c.cycleID,
		)
		if err != nil {
			log.Error().Err(err).Str("cycle_id", c.cycleID).Msg("iqub-worker: close cycle failed")
			continue
		}

		// Notify recipient.
		sendNotification(ctx, notifyURL, c.recipientID, "iqub.cycle_closed", map[string]string{
			"amount_etb": fmt.Sprintf("%.2f", float64(c.amountMinor)/100),
		})

		log.Info().Str("cycle_id", c.cycleID).Int64("payout_minor", c.amountMinor).Msg("iqub cycle closed")
	}
}

func markMissedContributions(ctx context.Context, pool *pgxpool.Pool, cycleID, groupID, notifyURL string) {
	// Find members who have not paid this cycle.
	rows, err := pool.Query(ctx, `
		SELECT m.user_id
		FROM iqub_memberships m
		WHERE m.group_id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM iqub_contributions co
		      WHERE co.cycle_id = $2 AND co.user_id = m.user_id
		  )
	`, groupID, cycleID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		sendNotification(ctx, notifyURL, userID, "iqub.contribution_missed", map[string]string{
			"cycle_id": cycleID,
		})
	}
}

func postPayout(ctx context.Context, ledgerURL string, c cycleRow) {
	body, _ := json.Marshal(map[string]any{
		"type":            "IQUB_PAYOUT",
		"idempotency_key": "iqub-payout-" + c.cycleID,
		"initiated_by":    "iqub-worker",
		"channel":         "SYSTEM",
		"entries": []map[string]any{
			{"account_id": c.poolAccountID, "direction": "D", "amount_minor": c.amountMinor},
			{"account_id": c.recipientID,   "direction": "C", "amount_minor": c.amountMinor},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ledgerURL+"/ledger/transactions", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Str("cycle_id", c.cycleID).Msg("iqub payout ledger post failed")
		return
	}
	resp.Body.Close()
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
