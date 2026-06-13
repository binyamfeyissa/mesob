package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mesob-wallet/ledger/internal/infra/config"
	"github.com/mesob-wallet/ledger/internal/infra/outbox"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
)

func main() {
	cfg := config.Load()
	log := kitlogging.New("ledger-worker")
	_ = cfg

	ctx, cancel := context.WithCancel(context.Background())
	relay := &outbox.Relay{}

	go relay.Run(ctx)

	log.Info().Msg("ledger worker started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	log.Info().Msg("ledger worker stopped")
}
