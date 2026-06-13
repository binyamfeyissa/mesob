package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/ledger/internal/app"
	"github.com/mesob-wallet/ledger/internal/infra/config"
	"github.com/mesob-wallet/ledger/internal/infra/outbox"
	"github.com/mesob-wallet/ledger/internal/infra/postgres"
	ledgerhttp "github.com/mesob-wallet/ledger/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("ledger")

	// Database pool.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := pgxpool.New(ctx, cfg.DBURL)
	cancel()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create pgx pool")
	}
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}
	pingCancel()
	defer pool.Close()

	// Repositories and event publisher.
	accountRepo := &postgres.AccountRepo{DB: pool}
	txnRepo := &postgres.TransactionRepo{DB: pool}
	entryRepo := &postgres.EntryRepo{DB: pool}
	publisher := &outbox.Publisher{DB: pool}
	relay := &outbox.Relay{DB: pool, KafkaBrokers: cfg.KafkaBrokers}

	// HTTP handler with wired use-cases.
	handler := &ledgerhttp.Handler{
		CreateAccount: &app.CreateAccountUseCase{
			Accounts: accountRepo,
		},
		PostTransaction: &app.PostTransactionUseCase{
			Accounts:     accountRepo,
			Transactions: txnRepo,
			Events:       publisher,
		},
		GetBalance: &app.GetBalanceUseCase{
			Accounts: accountRepo,
		},
		Entries:      entryRepo,
		Transactions: txnRepo,
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

	ledgerhttp.RegisterRoutes(mux, handler)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// Start outbox relay goroutine.
	relayCtx, relayCancel := context.WithCancel(context.Background())
	defer relayCancel()
	go relay.Run(relayCtx)

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("ledger starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("ledger failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("ledger stopped")
}
