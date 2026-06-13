package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/payments/internal/app"
	"github.com/mesob-wallet/payments/internal/infra/config"
	httpinfra "github.com/mesob-wallet/payments/internal/infra/http"
	"github.com/mesob-wallet/payments/internal/infra/postgres"
	paymentshttp "github.com/mesob-wallet/payments/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("payments")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Warn().Err(err).Msg("payments: could not connect to postgres — running without DB")
	}

	var merchantRepo *postgres.MerchantRepo
	var paymentRepo *postgres.PaymentRepo
	if pool != nil {
		merchantRepo = &postgres.MerchantRepo{DB: pool}
		paymentRepo = &postgres.PaymentRepo{DB: pool}
	}

	ledgerClient := &httpinfra.LedgerHTTPClient{
		BaseURL:    cfg.LedgerURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
	fraudClient := &httpinfra.FraudHTTPClient{
		BaseURL:    cfg.FraudURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
	identityClient := &httpinfra.IdentityHTTPClient{
		BaseURL:    cfg.IdentityURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	h := &paymentshttp.Handler{
		P2P: &app.P2PUseCase{
			Fraud:    fraudClient,
			Identity: identityClient,
			Ledger:   ledgerClient,
		},
		MerchantPayment: &app.MerchantPaymentUseCase{
			Merchants: merchantRepo,
			Ledger:    ledgerClient,
		},
		BillPayment: &app.BillPaymentUseCase{
			Billers:      paymentRepo,
			LedgerClient: ledgerClient,
			Payments:     paymentRepo,
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
	paymentshttp.RegisterRoutes(mux, h)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
	}
	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("payments starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("payments failed")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
}
