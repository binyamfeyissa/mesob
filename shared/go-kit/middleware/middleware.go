package middleware

import (
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORS reads allowed origins from the MESOB_CORS_ORIGIN env var (comma-separated).
// Defaults include the Swagger UI port so "Try it out" works out of the box.
func CORS(next http.Handler) http.Handler {
	raw := os.Getenv("MESOB_CORS_ORIGIN")
	if raw == "" {
		raw = "http://localhost:3000,http://localhost:3001,http://localhost:8080"
	}
	allowed := strings.Split(raw, ",")
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

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			u, _ := uuid.NewV7()
			id = u.String()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Dur("duration", time.Since(start)).
			Msg("request")
	})
}

// Chain composes middlewares right-to-left around a handler.
// Usage: Chain(handler, mw1, mw2) → mw1(mw2(handler))
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().
					Interface("panic", rec).
					Bytes("stack", debug.Stack()).
					Msg("recovered from panic")
				http.Error(w, `{"error":{"code":"INTERNAL","message":"internal server error"}}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
