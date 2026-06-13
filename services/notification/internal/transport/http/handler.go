package http

import (
	"encoding/json"
	"net/http"

	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
	kitresp "github.com/mesob-wallet/go-kit/response"
	"github.com/mesob-wallet/notification/internal/app"
	"github.com/mesob-wallet/notification/internal/domain"
	"github.com/rs/zerolog"
)

type Handler struct {
	send *app.SendUseCase
	log  zerolog.Logger
}

func NewHandler(send *app.SendUseCase) *Handler {
	return &Handler{send: send, log: kitlogging.New("notification")}
}


func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)

	chain := func(f http.HandlerFunc) http.Handler {
		return kitmw.Chain(f, kitmw.RequestID, kitmw.Recovery, kitmw.Logging)
	}

	mux.Handle("POST /notify/send", chain(h.sendNotification))
	mux.Handle("GET /notify/deliveries/{id}", chain(h.getDelivery))
	mux.Handle("PUT /notify/templates/{key}", chain(h.upsertTemplate))
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ready"}`))
}

func (h *Handler) sendNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string            `json:"user_id"`
		TemplateKey string            `json:"template"`
		Params      map[string]string `json:"params"`
		ChannelHint string            `json:"channel_hint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	out, err := h.send.Execute(r.Context(), app.SendInput{
		UserID:      req.UserID,
		TemplateKey: req.TemplateKey,
		Params:      req.Params,
		ChannelHint: domain.Channel(req.ChannelHint),
	})
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), kitresp.MetaBlock{})
		return
	}
	kitresp.Success(w, http.StatusAccepted, map[string]any{"delivery_id": out.DeliveryID, "status": out.Status}, kitresp.MetaBlock{})
}

func (h *Handler) getDelivery(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if h.send.Deliveries != nil {
		d, err := h.send.Deliveries.FindByID(r.Context(), id)
		if err != nil {
			kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "delivery not found", kitresp.MetaBlock{})
			return
		}
		kitresp.Success(w, http.StatusOK, map[string]any{
			"delivery_id": d.ID,
			"user_id":     d.UserID,
			"template":    d.TemplateKey,
			"channel":     string(d.Channel),
			"status":      string(d.Status),
			"attempts":    d.Attempts,
			"created_at":  d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}, kitresp.MetaBlock{})
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]any{"delivery_id": id, "status": "QUEUED"}, kitresp.MetaBlock{})
}

func (h *Handler) upsertTemplate(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var req struct {
		Lang    string `json:"lang"`
		Channel string `json:"channel"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.send.UpsertTemplate(r.Context(), app.UpsertTemplateInput{
		Key:     key,
		Lang:    domain.Lang(req.Lang),
		Channel: domain.Channel(req.Channel),
		Body:    req.Body,
	}); err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), kitresp.MetaBlock{})
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]string{"key": key}, kitresp.MetaBlock{})
}
