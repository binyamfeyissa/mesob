package http

import (
	"encoding/json"
	"net/http"

	kitlogging "github.com/mesob-wallet/go-kit/logging"
	kitmw "github.com/mesob-wallet/go-kit/middleware"
	kitresp "github.com/mesob-wallet/go-kit/response"
	"github.com/mesob-wallet/adapter-hub/internal/app"
	"github.com/mesob-wallet/adapter-hub/internal/domain"
	"github.com/rs/zerolog"
)

type Handler struct {
	nid     *app.NIDVerifyUseCase
	mfi     *app.MFIOriginateUseCase
	webhook *app.PartnerWebhookUseCase
	mode    domain.AdapterMode
	log     zerolog.Logger
}

func NewHandler(nid *app.NIDVerifyUseCase, mfi *app.MFIOriginateUseCase, wh *app.PartnerWebhookUseCase, mode domain.AdapterMode) *Handler {
	return &Handler{nid: nid, mfi: mfi, webhook: wh, mode: mode, log: kitlogging.New("adapter-hub")}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)

	chain := func(f http.HandlerFunc) http.Handler {
		return kitmw.Chain(f, kitmw.RequestID, kitmw.Recovery, kitmw.Logging)
	}

	mux.Handle("POST /adapters/nid/verify", chain(h.verifyNID))
	mux.Handle("POST /adapters/mfi/originate", chain(h.originateMFI))
	mux.Handle("POST /adapters/webhooks/{partner}", chain(h.partnerWebhook))
	mux.Handle("GET /adapters/status", chain(h.status))
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ready"}`))
}

func (h *Handler) verifyNID(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FAN   string `json:"fan"`
		Claim struct {
			Name string `json:"name"`
			DOB  string `json:"dob"`
		} `json:"claim"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	out, err := h.nid.Execute(r.Context(), app.NIDVerifyInput{FAN: req.FAN, Name: req.Claim.Name, DOB: req.Claim.DOB})
	if err != nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "PROVIDER_UNAVAILABLE", err.Error(), kitresp.MetaBlock{})
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]any{
		"verified":    out.Verified,
		"match_score": out.MatchScore,
		"mode":        out.Mode,
	}, kitresp.MetaBlock{})
}

func (h *Handler) originateMFI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserRef        string `json:"user_ref"`
		PrincipalMinor int64  `json:"principal_minor"`
		TermDays       int    `json:"term_days"`
		ScoreRef       string `json:"score_ref"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	out, err := h.mfi.Execute(r.Context(), app.MFIOriginateInput{
		UserRef:        req.UserRef,
		PrincipalMinor: req.PrincipalMinor,
		TermDays:       req.TermDays,
		ScoreRef:       req.ScoreRef,
	})
	if err != nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "PROVIDER_UNAVAILABLE", err.Error(), kitresp.MetaBlock{})
		return
	}
	kitresp.Success(w, http.StatusCreated, map[string]any{
		"facility_id": out.FacilityID,
		"status":      out.Status,
		"mode":        out.Mode,
	}, kitresp.MetaBlock{})
}

func (h *Handler) partnerWebhook(w http.ResponseWriter, r *http.Request) {
	partner := r.PathValue("partner")
	sig := r.Header.Get("X-Signature")
	ts := r.Header.Get("X-Timestamp")
	out, err := h.webhook.Execute(r.Context(), app.PartnerWebhookInput{
		Partner:   partner,
		Body:      r.Body,
		Signature: sig,
		Timestamp: ts,
	})
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_SIGNATURE", err.Error(), kitresp.MetaBlock{})
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]bool{"received": out.Received}, kitresp.MetaBlock{})
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	statuses := []map[string]any{
		{"adapter": "nid", "mode": h.mode, "breaker": "CLOSED", "healthy": true},
		{"adapter": "mfi", "mode": h.mode, "breaker": "CLOSED", "healthy": true},
	}
	kitresp.Success(w, http.StatusOK, statuses, kitresp.MetaBlock{})
}
