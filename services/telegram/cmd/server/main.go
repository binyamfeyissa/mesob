package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
	"github.com/mesob-wallet/telegram/internal/app"
	"github.com/mesob-wallet/telegram/internal/infra/bot"
	"github.com/mesob-wallet/telegram/internal/infra/config"
	"github.com/mesob-wallet/telegram/internal/infra/gateway"
	httpTransport "github.com/mesob-wallet/telegram/internal/transport/http"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("telegram")

	gatewayURL := getenv("MESOB_USSD_URL", "http://ussd:8010")

	webhook := app.NewWebhookUseCase(
		&bot.LoggerClient{},
		gateway.New(gatewayURL),
	)

	handler := httpTransport.NewHandler(webhook)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("telegram starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Info().Msg("telegram stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
