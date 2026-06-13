package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/mesob-wallet/ussd/internal/app"
	"github.com/mesob-wallet/ussd/internal/infra/config"
	redisinfra "github.com/mesob-wallet/ussd/internal/infra/redis"
	ussdhttp "github.com/mesob-wallet/ussd/internal/transport/http"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("ussd")
	ctx := context.Background()

	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.RedisURL})
	var sessions *redisinfra.SessionStore
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("ussd: redis ping failed — sessions will be in-memory only")
	} else {
		log.Info().Msg("redis connected")
		sessions = &redisinfra.SessionStore{Redis: rdb}
	}

	h := &ussdhttp.Handler{
		Callback: &app.CallbackUseCase{Sessions: sessions},
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
	ussdhttp.RegisterRoutes(mux, h)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("ussd starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("ussd failed")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info().Msg("ussd stopped")
}
