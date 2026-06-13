package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/loans/internal/app"
	"github.com/mesob-wallet/loans/internal/infra/config"
	httpinfra "github.com/mesob-wallet/loans/internal/infra/http"
	"github.com/mesob-wallet/loans/internal/infra/postgres"
	loanshttp "github.com/mesob-wallet/loans/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("loans")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Warn().Err(err).Msg("loans: could not connect to postgres — running without DB")
	} else if err := pool.Ping(ctx); err != nil {
		log.Warn().Err(err).Msg("loans: postgres ping failed — running without DB")
		pool = nil
	} else {
		log.Info().Msg("postgres connected")
		defer pool.Close()
	}

	var loanRepo *postgres.LoanRepo
	if pool != nil {
		loanRepo = &postgres.LoanRepo{DB: pool}
	}

	ledgerClient := &httpinfra.LedgerHTTPClient{
		BaseURL:    getenv("MESOB_LOANS_LEDGER_URL", "http://ledger:8002"),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	scoringClient := &httpinfra.ScoringHTTPClient{
		BaseURL:    cfg.ScoringURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	h := &loanshttp.Handler{
		CheckEligibility: &app.CheckEligibilityUseCase{
			Scoring: scoringClient,
		},
		ApplyLoan: &app.ApplyLoanUseCase{
			Loans:   loanRepo,
			Scoring: scoringClient,
			Ledger:  ledgerClient,
		},
		RepayLoan: &app.RepayLoanUseCase{
			Loans:  loanRepo,
			Ledger: ledgerClient,
		},
		Loans: loanRepo,
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
	loanshttp.RegisterRoutes(mux, h)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("loans starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("loans failed")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("loans stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
