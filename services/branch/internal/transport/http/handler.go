package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/branch/internal/app"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	ApproveAgent   *app.ApproveAgentUseCase
	Settle         *app.SettleUseCase
	ResolveDispute *app.ResolveDisputeUseCase
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /branch/agents/{id}/approve",  h.ApproveAgent_)
	mux.HandleFunc("POST /branch/settlements",          h.Settle_)
	mux.HandleFunc("GET /branch/settlements",           h.GetSettlements)
	mux.HandleFunc("POST /branch/kyc/{userId}/review",  h.ReviewKYC)
	mux.HandleFunc("GET /branch/kyc/queue",             h.GetKYCQueue)
	mux.HandleFunc("POST /branch/disputes/{id}/resolve", h.ResolveDispute_)
	mux.HandleFunc("GET /branch/disputes",              h.GetDisputes)
	mux.HandleFunc("GET /branch/reconciliation",        h.GetReconciliation)
}

func (h *Handler) ApproveAgent_(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid agent id", meta(r))
		return
	}
	var req struct {
		FloatCeilingMinor int64 `json:"float_ceiling_minor"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	out, err := h.ApproveAgent.Execute(r.Context(), app.ApproveAgentInput{
		AgentID: agentID, FloatCeilingMinor: req.FloatCeilingMinor,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) Settle_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	var req struct {
		AgentID          string `json:"agent_id"`
		AmountMinor      int64  `json:"amount_minor"`
		SecondAuthoriser string `json:"second_authoriser_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	agentID, _ := uuid.FromString(req.AgentID)
	secondID, _ := uuid.FromString(req.SecondAuthoriser)
	out, err := h.Settle.Execute(r.Context(), app.SettleInput{
		AgentID: agentID, SecondAuthoriser: secondID,
		AmountMinor: req.AmountMinor, IdempotencyKey: idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) GetSettlements(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	settlements := []map[string]any{
		{
			"id":           "s1e2f3a4-b5c6-7890-abcd-ef1234567890",
			"agent_id":     "550e8400-e29b-41d4-a716-446655440020",
			"amount_minor": 250000,
			"status":       "CONFIRMED",
			"confirmed_at": now.Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":           "s2e2f3a4-b5c6-7890-abcd-ef1234567890",
			"agent_id":     "550e8400-e29b-41d4-a716-446655440021",
			"amount_minor": 180000,
			"status":       "PENDING",
			"confirmed_at": now.Add(-3 * time.Hour).Format(time.RFC3339),
		},
	}
	kitresp.Success(w, http.StatusOK, settlements, meta(r))
}

func (h *Handler) ReviewKYC(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	var req struct {
		Decision   string `json:"decision"`
		TargetTier int8   `json:"target_tier"`
		Note       string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	kitresp.Success(w, http.StatusOK, map[string]any{
		"user_id":     userID,
		"decision":    req.Decision,
		"target_tier": req.TargetTier,
		"reviewed_at": time.Now().UTC().Format(time.RFC3339),
	}, meta(r))
}

func (h *Handler) GetKYCQueue(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	queue := []map[string]any{
		{
			"user_id":        "550e8400-e29b-41d4-a716-446655440010",
			"msisdn":         "+251911000010",
			"current_tier":   1,
			"requested_tier": 2,
			"submitted_at":   now.Add(-2 * time.Hour).Format(time.RFC3339),
			"status":         "PENDING",
		},
		{
			"user_id":        "550e8400-e29b-41d4-a716-446655440011",
			"msisdn":         "+251911000011",
			"current_tier":   0,
			"requested_tier": 1,
			"submitted_at":   now.Add(-5 * time.Hour).Format(time.RFC3339),
			"status":         "PENDING",
		},
		{
			"user_id":        "550e8400-e29b-41d4-a716-446655440012",
			"msisdn":         "+251922000012",
			"current_tier":   1,
			"requested_tier": 3,
			"submitted_at":   now.Add(-30 * time.Minute).Format(time.RFC3339),
			"status":         "PENDING",
		},
	}
	kitresp.Success(w, http.StatusOK, queue, meta(r))
}

func (h *Handler) ResolveDispute_(w http.ResponseWriter, r *http.Request) {
	disputeID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid dispute id", meta(r))
		return
	}
	var req struct {
		Resolution       string `json:"resolution"`
		SecondAuthoriser string `json:"second_authoriser_id"`
		Reason           string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	secondID, _ := uuid.FromString(req.SecondAuthoriser)
	out, err := h.ResolveDispute.Execute(r.Context(), app.ResolveDisputeInput{
		DisputeID: disputeID, SecondAuthoriser: secondID,
		Resolution: req.Resolution, Reason: req.Reason,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) GetDisputes(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	disputes := []map[string]any{
		{
			"id":             "d1e2f3a4-b5c6-7890-abcd-ef1234567890",
			"transaction_id": "t1e2f3a4-b5c6-7890-abcd-ef1234567890",
			"raised_by":      "+251911000013",
			"reason":         "DOUBLE_CHARGE",
			"resolution":     nil,
			"created_at":     now.Add(-1 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":             "d2e2f3a4-b5c6-7890-abcd-ef1234567890",
			"transaction_id": "t2e2f3a4-b5c6-7890-abcd-ef1234567890",
			"raised_by":      "+251922000014",
			"reason":         "NOT_RECEIVED",
			"resolution":     nil,
			"created_at":     now.Add(-3 * time.Hour).Format(time.RFC3339),
		},
	}
	kitresp.Success(w, http.StatusOK, disputes, meta(r))
}

func (h *Handler) GetReconciliation(w http.ResponseWriter, r *http.Request) {
	kitresp.Success(w, http.StatusOK, map[string]any{
		"region_id":      "550e8400-e29b-41d4-a716-446655440100",
		"date":           time.Now().Format("2006-01-02"),
		"ledger_minor":   4523500,
		"counted_minor":  4523500,
		"variance_minor": 0,
		"status":         "BALANCED",
	}, meta(r))
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	if de, ok := err.(*kiterr.DomainError); ok {
		status := http.StatusUnprocessableEntity
		switch de.Code {
		case "OUT_OF_REGION", "FRAUD_BLOCKED", "SAME_AUTHORISER":
			status = http.StatusForbidden
		case "NOT_FOUND":
			status = http.StatusNotFound
		}
		kitresp.Err(w, status, de.Code, de.Message, meta(r))
		return
	}
	kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
}

func meta(r *http.Request) kitresp.MetaBlock {
	return kitresp.MetaBlock{RequestID: r.Header.Get("X-Request-ID"), TraceID: r.Header.Get("X-Trace-ID")}
}
