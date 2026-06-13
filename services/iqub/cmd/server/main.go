package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/iqub/internal/app"
	"github.com/mesob-wallet/iqub/internal/infra/config"
	httpinfra "github.com/mesob-wallet/iqub/internal/infra/http"
	iquboutbox "github.com/mesob-wallet/iqub/internal/infra/outbox"
	"github.com/mesob-wallet/iqub/internal/infra/postgres"
	iqubhttp "github.com/mesob-wallet/iqub/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("iqub")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Warn().Err(err).Msg("iqub: could not connect to postgres — running without DB")
	} else if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Warn().Err(pingErr).Msg("iqub: postgres ping failed — running without DB")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	var groupRepo *postgres.GroupRepo
	var membershipRepo *postgres.MembershipRepo
	var cycleRepo *postgres.CycleRepo
	if pool != nil {
		groupRepo = &postgres.GroupRepo{DB: pool}
		membershipRepo = &postgres.MembershipRepo{DB: pool}
		cycleRepo = &postgres.CycleRepo{DB: pool}
	}

	ledgerClient := &httpinfra.LedgerHTTPClient{
		BaseURL:    getenv("MESOB_IQUB_LEDGER_URL", "http://ledger:8002"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	publisher := &iquboutbox.Publisher{DB: pool}
	relay := &iquboutbox.Relay{DB: pool, KafkaBrokers: cfg.KafkaBrokers}
	go relay.Run(ctx)

	h := &iqubhttp.Handler{
		Cycles: cycleRepo,
		CreateGroup: &app.CreateGroupUseCase{
			Groups:      groupRepo,
			Memberships: membershipRepo,
			Ledger:      ledgerClient,
			Events:      publisher,
		},
		Contribute: &app.ContributeUseCase{
			Groups:      groupRepo,
			Memberships: membershipRepo,
			Ledger:      ledgerClient,
			Events:      publisher,
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
	iqubhttp.RegisterRoutes(mux, h)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("iqub starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("iqub failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("iqub stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
