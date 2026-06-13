package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mesob-wallet/gateway/internal/config"
	"github.com/mesob-wallet/gateway/internal/proxy"
	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
)

func corsMiddleware(origins string) func(http.Handler) http.Handler {
	allowed := strings.Split(origins, ",")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, o := range allowed {
				if strings.TrimSpace(o) == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Refresh-Token")
					w.Header().Set("Access-Control-Max-Age", "86400")
					break
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	cfg := config.Load()
	log := kitlogging.New("gateway")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ready"}`))
	})

	routes := map[string]string{
		"/v1/identity/":  cfg.IdentityURL,
		"/v1/ledger/":    cfg.LedgerURL,
		"/v1/payments/":  cfg.PaymentsURL,
		"/v1/loans/":     cfg.LoansURL,
		"/v1/iqub/":      cfg.IqubURL,
		"/v1/iddir/":     cfg.IddirURL,
		"/v1/agent/":     cfg.AgentURL,
		"/v1/branch/":    cfg.BranchURL,
		"/v1/admin/":     cfg.AdminURL,
		"/v1/ussd/":      cfg.UssdURL,
	}

	for prefix, upstream := range routes {
		p := proxy.Handler(upstream)
		mux.Handle(prefix, http.StripPrefix("/v1", p))
	}

	handler := corsMiddleware(cfg.CORSOrigin)(kitmw.Recovery(kitmw.Logging(kitmw.RequestID(mux))))

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.HTTPPort).Msg("gateway starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("gateway failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Info().Msg("gateway stopped")
}
