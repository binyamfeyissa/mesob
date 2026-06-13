package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/branch/internal/app"
	"github.com/mesob-wallet/branch/internal/infra/config"
	"github.com/mesob-wallet/branch/internal/infra/postgres"
	branchhttp "github.com/mesob-wallet/branch/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("branch")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Warn().Err(err).Msg("branch: could not connect to postgres — running without DB")
	} else if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Warn().Err(pingErr).Msg("branch: postgres ping failed — running without DB")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	var settlementRepo *postgres.SettlementRepo
	var disputeRepo *postgres.DisputeRepo
	if pool != nil {
		settlementRepo = &postgres.SettlementRepo{DB: pool}
		disputeRepo = &postgres.DisputeRepo{DB: pool}
	}

	h := &branchhttp.Handler{
		ApproveAgent: &app.ApproveAgentUseCase{},
		Settle: &app.SettleUseCase{
			Settlements: settlementRepo,
		},
		ResolveDispute: &app.ResolveDisputeUseCase{
			Disputes: disputeRepo,
		},
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
	branchhttp.RegisterRoutes(mux, h)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("branch starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("branch failed")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("branch stopped")
}
