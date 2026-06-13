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
	"github.com/mesob-wallet/adapter-hub/internal/app"
	"github.com/mesob-wallet/adapter-hub/internal/domain"
	"github.com/mesob-wallet/adapter-hub/internal/infra/config"
	webhookinfra "github.com/mesob-wallet/adapter-hub/internal/infra/webhook"
	httpTransport "github.com/mesob-wallet/adapter-hub/internal/transport/http"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("adapter-hub")

	mode := domain.AdapterMode(cfg.Mode)

	nidUC := &app.NIDVerifyUseCase{Mode: mode}
	mfiUC := &app.MFIOriginateUseCase{Mode: mode}
	whUC := &app.PartnerWebhookUseCase{Processor: &webhookinfra.LoggingProcessor{}}

	handler := httpTransport.NewHandler(nidUC, mfiUC, whUC, mode)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      kitmw.CORS(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux)))),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("adapter-hub starting")
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
	log.Info().Msg("adapter-hub stopped")
}
