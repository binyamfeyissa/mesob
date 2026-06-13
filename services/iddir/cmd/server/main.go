package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/iddir/internal/app"
	"github.com/mesob-wallet/iddir/internal/infra/config"
	httpinfra "github.com/mesob-wallet/iddir/internal/infra/http"
	iddiroutbox "github.com/mesob-wallet/iddir/internal/infra/outbox"
	"github.com/mesob-wallet/iddir/internal/infra/postgres"
	iddirhttp "github.com/mesob-wallet/iddir/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("iddir")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Warn().Err(err).Msg("iddir: could not connect to postgres — running without DB")
	} else if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Warn().Err(pingErr).Msg("iddir: postgres ping failed — running without DB")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	var groupRepo *postgres.GroupRepo
	var claimRepo *postgres.ClaimRepo
	var membershipRepo *postgres.MembershipRepo
	if pool != nil {
		groupRepo = &postgres.GroupRepo{DB: pool}
		claimRepo = &postgres.ClaimRepo{DB: pool}
		membershipRepo = &postgres.MembershipRepo{DB: pool}
	}

	ledgerClient := &httpinfra.LedgerHTTPClient{
		BaseURL:    getenv("MESOB_IDDIR_LEDGER_URL", "http://ledger:8002"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	publisher := &iddiroutbox.Publisher{DB: pool}
	relay := &iddiroutbox.Relay{DB: pool, KafkaBrokers: cfg.KafkaBrokers}
	go relay.Run(ctx)

	h := &iddirhttp.Handler{
		Memberships: membershipRepo,
		PayPremium: &app.PayPremiumUseCase{
			Groups: groupRepo,
			Ledger: ledgerClient,
			Events: publisher,
		},
		FileClaim: &app.FileClaimUseCase{
			Groups: groupRepo,
			Claims: claimRepo,
			Events: publisher,
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
	iddirhttp.RegisterRoutes(mux, h)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("iddir starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("iddir failed")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("iddir stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
