package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/agent/internal/app"
	"github.com/mesob-wallet/agent/internal/infra/config"
	httpinfra "github.com/mesob-wallet/agent/internal/infra/http"
	"github.com/mesob-wallet/agent/internal/infra/postgres"
	agenthttp "github.com/mesob-wallet/agent/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("agent")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Warn().Err(err).Msg("agent: could not connect to postgres — running without DB")
	} else if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Warn().Err(pingErr).Msg("agent: postgres ping failed — running without DB")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	var agentRepo *postgres.AgentRepo
	if pool != nil {
		agentRepo = &postgres.AgentRepo{DB: pool}
	}

	ledgerClient := &httpinfra.LedgerHTTPClient{
		BaseURL:    getenv("MESOB_AGENT_LEDGER_URL", "http://ledger:8002"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	identityClient := &httpinfra.IdentityHTTPClient{
		BaseURL:    getenv("MESOB_AGENT_IDENTITY_URL", "http://identity:8001"),
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	cashIn := &app.CashInUseCase{
		Agents:   agentRepo,
		Ledger:   ledgerClient,
		Identity: identityClient,
	}
	cashOut := &app.CashOutUseCase{
		Agents:   agentRepo,
		Ledger:   ledgerClient,
		Identity: identityClient,
	}

	h := &agenthttp.Handler{
		CashIn:  cashIn,
		CashOut: cashOut,
		Sync: &app.SyncUseCase{
			Agents:  agentRepo,
			CashIn:  cashIn,
			CashOut: cashOut,
		},
		Agents:   agentRepo,
		Identity: identityClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ready"}`))
	})
	agenthttp.RegisterRoutes(mux, h)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("agent starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("agent failed")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("agent stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
