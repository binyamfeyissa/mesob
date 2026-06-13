package http

import (
	"encoding/json"
	"net/http"

	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
	kitresp "github.com/mesob-wallet/go-kit/response"
	"github.com/mesob-wallet/telegram/internal/app"
	"github.com/mesob-wallet/telegram/internal/domain"
	"github.com/rs/zerolog"
)

type Handler struct {
	webhook *app.WebhookUseCase
	log     zerolog.Logger
}

func NewHandler(webhook *app.WebhookUseCase) *Handler {
	return &Handler{
		webhook: webhook,
		log:     kitlogging.New("telegram"),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.Handle("POST /telegram/webhook", kitmw.Chain(
		http.HandlerFunc(h.handleWebhook),
		kitmw.RequestID,
		kitmw.Recovery,
		kitmw.Logging,
	))
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var update domain.WebhookUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	out, err := h.webhook.Execute(r.Context(), app.WebhookInput{Update: update})
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), kitresp.MetaBlock{})
		return
	}
	kitresp.Success(w, http.StatusOK, out, kitresp.MetaBlock{})
}
