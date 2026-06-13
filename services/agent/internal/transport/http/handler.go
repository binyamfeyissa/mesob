package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/agent/internal/app"
	"github.com/mesob-wallet/agent/internal/domain"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	CashIn   *app.CashInUseCase
	CashOut  *app.CashOutUseCase
	Sync     *app.SyncUseCase
	Agents   app.AgentRepository
	Identity app.CustomerRegistrar
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /agent/cash-in", h.CashIn_)
	mux.HandleFunc("POST /agent/cash-out", h.CashOut_)
	mux.HandleFunc("POST /agent/sync", h.Sync_)
	mux.HandleFunc("POST /agent/onboard", h.Onboard)
	mux.HandleFunc("GET /agent/float", h.GetFloat)
}

func (h *Handler) CashIn_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	var req struct {
		UserMSISDN  string `json:"user_msisdn"`
		AmountMinor int64  `json:"amount_minor"`
		CapturedAt  string `json:"captured_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	agentUserID := subFromBearer(r)
	out, err := h.CashIn.Execute(r.Context(), app.CashInInput{
		AgentUserID:    agentUserID,
		UserMSISDN:     req.UserMSISDN,
		AmountMinor:    req.AmountMinor,
		CapturedAt:     req.CapturedAt,
		IdempotencyKey: idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) CashOut_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	var req struct {
		UserMSISDN  string `json:"user_msisdn"`
		AmountMinor int64  `json:"amount_minor"`
		AuthCode    string `json:"authorisation_code"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	agentUserID := subFromBearer(r)
	out, err := h.CashOut.Execute(r.Context(), app.CashOutInput{
		AgentUserID:    agentUserID,
		UserMSISDN:     req.UserMSISDN,
		AmountMinor:    req.AmountMinor,
		AuthCode:       req.AuthCode,
		IdempotencyKey: idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) Sync_(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Operations  []domain.Operation `json:"operations"`
		SinceCursor string             `json:"since_cursor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	agentUserID := subFromBearer(r)
	out, err := h.Sync.Execute(r.Context(), app.SyncInput{
		AgentUserID: agentUserID,
		Operations:  req.Operations,
		SinceCursor: req.SinceCursor,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) GetFloat(w http.ResponseWriter, r *http.Request) {
	agentUserID := subFromBearer(r)
	if h.Agents == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "AGENT_UNAVAILABLE", "agent repo not wired", meta(r))
		return
	}
	agent, err := h.Agents.FindByUserID(r.Context(), agentUserID)
	if err != nil {
		// No agent record yet (new user or unregistered) — return a default profile so
		// the mobile app can boot without a 404 error.
		kitresp.Success(w, http.StatusOK, map[string]any{
			"agent_id":          uuid.Nil.String(),
			"region_id":         uuid.Nil.String(),
			"float_minor":       int64(0),
			"float_limit_minor": int64(0),
			"status":            "UNREGISTERED",
		}, meta(r))
		return
	}
	var floatMinor int64
	if h.CashIn != nil && h.CashIn.Ledger != nil && agent.FloatAccountID != nil {
		floatMinor, _ = h.CashIn.Ledger.GetBalance(r.Context(), agent.FloatAccountID.String())
	}
	kitresp.Success(w, http.StatusOK, map[string]any{
		"agent_id":          agent.ID.String(),
		"region_id":         agent.RegionID.String(),
		"float_minor":       floatMinor,
		"float_limit_minor": agent.FloatLimitMinor,
		"status":            string(agent.Status),
	}, meta(r))
}

func (h *Handler) Onboard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MSISDN string `json:"msisdn"`
		Lang   string `json:"preferred_lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	if req.MSISDN == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_MSISDN", "msisdn is required", meta(r))
		return
	}
	if req.Lang == "" {
		req.Lang = "am"
	}
	if h.Identity == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "IDENTITY_UNAVAILABLE", "identity service not wired", meta(r))
		return
	}
	if err := h.Identity.RegisterCustomer(r.Context(), req.MSISDN, req.Lang); err != nil {
		kitresp.Err(w, http.StatusBadGateway, "REGISTRATION_FAILED", err.Error(), meta(r))
		return
	}
	kitresp.Success(w, http.StatusAccepted, map[string]any{
		"msisdn":  req.MSISDN,
		"message": "OTP sent to customer",
	}, meta(r))
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	type domainErr interface {
		Error() string
	}
	kitresp.Err(w, http.StatusUnprocessableEntity, "UNPROCESSABLE", err.Error(), meta(r))
}

func subFromBearer(r *http.Request) uuid.UUID {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return uuid.Nil
	}
	parts := strings.Split(strings.TrimPrefix(auth, "Bearer "), ".")
	if len(parts) != 3 {
		return uuid.Nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil
	}
	sub, _ := claims["sub"].(string)
	id, _ := uuid.FromString(sub)
	return id
}

func meta(r *http.Request) kitresp.MetaBlock {
	return kitresp.MetaBlock{RequestID: r.Header.Get("X-Request-ID"), TraceID: r.Header.Get("X-Trace-ID")}
}
