package http

import (
	"encoding/json"
	"net/http"

	"github.com/mesob-wallet/ussd/internal/app"
	"github.com/mesob-wallet/ussd/internal/domain"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	Callback *app.CallbackUseCase
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /ussd/callback", h.Callback_)
	mux.HandleFunc("GET /ussd/menus", h.GetMenus)
}

func (h *Handler) Callback_(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID   string `json:"session_id"`
		MSISDN      string `json:"msisdn"`
		Input       string `json:"input"`
		ServiceCode string `json:"service_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	out, err := h.Callback.Execute(r.Context(), app.CallbackInput{
		SessionID:   req.SessionID,
		MSISDN:      req.MSISDN,
		Input:       req.Input,
		ServiceCode: req.ServiceCode,
	})
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) GetMenus(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	menus, ok := domain.Menus[lang]
	if !ok {
		menus = domain.Menus["en"]
	}
	kitresp.Success(w, http.StatusOK, menus, meta(r))
}

func meta(r *http.Request) kitresp.MetaBlock {
	return kitresp.MetaBlock{RequestID: r.Header.Get("X-Request-ID"), TraceID: r.Header.Get("X-Trace-ID")}
}
