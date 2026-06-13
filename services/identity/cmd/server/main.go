package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
	"github.com/mesob-wallet/identity/internal/app"
	"github.com/mesob-wallet/identity/internal/infra/config"
	httpinfra "github.com/mesob-wallet/identity/internal/infra/http"
	identityjwt "github.com/mesob-wallet/identity/internal/infra/jwt"
	identityoutbox "github.com/mesob-wallet/identity/internal/infra/outbox"
	"github.com/mesob-wallet/identity/internal/infra/otp"
	"github.com/mesob-wallet/identity/internal/infra/postgres"
	identityhttp "github.com/mesob-wallet/identity/internal/transport/http"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("identity")
	ctx := context.Background()

	// Postgres
	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres connect failed")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("postgres ping failed")
	}
	log.Info().Msg("postgres connected")

	// Redis
	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.RedisURL})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("redis ping failed — sessions will not persist across restarts")
	} else {
		log.Info().Msg("redis connected")
	}

	jwtSecret := []byte(getenv("MESOB_JWT_SECRET", "mesob-dev-secret-change-in-production"))

	userRepo := &postgres.UserRepo{DB: pool}
	credRepo := &postgres.CredentialRepo{DB: pool}
	sessions := &identityjwt.SessionStore{Redis: rdb, Secret: jwtSecret}

	otpSvc := &otp.RedisOTP{Redis: rdb}

	ledgerURL := getenv("MESOB_IDENTITY_LEDGER_URL", "http://ledger:8002")
	ledgerClient := &httpinfra.LedgerHTTPClient{
		BaseURL:    ledgerURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	publisher := &identityoutbox.Publisher{DB: pool}
	relay := &identityoutbox.Relay{DB: pool, KafkaBrokers: cfg.KafkaBrokers}

	handler := &identityhttp.Handler{
		Register: &app.RegisterUseCase{
			Users: userRepo,
			OTP:   otpSvc,
		},
		VerifyOTP: &app.VerifyOTPUseCase{
			OTP: otpSvc,
		},
		SetPIN: &app.SetPINUseCase{
			Users:    userRepo,
			Creds:    credRepo,
			Ledger:   ledgerClient,
			Sessions: sessions,
			Events:   publisher,
		},
		Login: &app.LoginUseCase{
			Users:    userRepo,
			Creds:    credRepo,
			Sessions: sessions,
		},
		KYCUpgrade: &app.KYCUpgradeUseCase{
			Users:   userRepo,
			Events:  publisher,
		},
		Users:     userRepo,
		JWTSecret: jwtSecret,
	}

	go relay.Run(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ready"}`))
	})

	identityhttp.RegisterRoutes(mux, handler)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler: kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("identity service starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("identity service failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("identity service stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
