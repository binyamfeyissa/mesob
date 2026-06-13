package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	"github.com/mesob-wallet/branch/internal/infra/config"
	"github.com/rs/zerolog"
)

var log zerolog.Logger

func main() {
	cfg := config.Load()
	log = kitlogging.New("branch-worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil || pool.Ping(ctx) != nil {
		log.Warn().Msg("branch-worker: postgres unavailable, running in no-op mode")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	notifyURL := getenv("MESOB_NOTIFICATION_URL", "http://notification:8012")

	go runSettlementSweep(ctx, pool, notifyURL)

	log.Info().Msg("branch worker started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	log.Info().Msg("branch worker stopped")
}

func runSettlementSweep(ctx context.Context, pool *pgxpool.Pool, notifyURL string) {
	sweepSettlements(ctx, pool, notifyURL)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sweepSettlements(ctx, pool, notifyURL)
		}
	}
}

type regionSummary struct {
	regionID         string
	totalFloatMinor  int64
	agentCount       int
}

func sweepSettlements(ctx context.Context, pool *pgxpool.Pool, notifyURL string) {
	if pool == nil {
		return
	}

	// Aggregate float balances per region from agent records.
	rows, err := pool.Query(ctx, `
		SELECT region_id,
		       COUNT(*) AS agent_count,
		       COALESCE(SUM(float_limit_minor), 0) AS total_float_minor
		FROM agents
		WHERE status = 'ACTIVE'
		GROUP BY region_id
	`)
	if err != nil {
		log.Error().Err(err).Msg("branch settlement sweep: query failed")
		return
	}
	defer rows.Close()

	var summaries []regionSummary
	for rows.Next() {
		var s regionSummary
		if err := rows.Scan(&s.regionID, &s.agentCount, &s.totalFloatMinor); err != nil {
			continue
		}
		summaries = append(summaries, s)
	}
	rows.Close()

	for _, s := range summaries {
		// Record the settlement summary.
		_, err := pool.Exec(ctx, `
			INSERT INTO branch_settlement_sweeps
			    (region_id, agent_count, total_float_minor, swept_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`, s.regionID, s.agentCount, s.totalFloatMinor, time.Now().UTC())
		if err != nil {
			log.Warn().Err(err).Str("region_id", s.regionID).Msg("settlement record failed")
		}
	}

	if len(summaries) > 0 {
		log.Info().Int("regions", len(summaries)).Msg("settlement sweep complete")
	}

	// Notify any branch managers of the sweep result.
	sendNotification(ctx, notifyURL, "system", "branch.settlement_sweep_complete", map[string]string{
		"regions": string(mustMarshal(len(summaries))),
	})
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

func mustMarshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
