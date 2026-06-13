package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/iddir/internal/app"
	"github.com/mesob-wallet/iddir/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	PayPremium  *app.PayPremiumUseCase
	FileClaim   *app.FileClaimUseCase
	Memberships app.MembershipRepository
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /iddir/groups", h.ListGroups)
	mux.HandleFunc("POST /iddir/groups", h.CreateGroup)
	mux.HandleFunc("GET /iddir/groups/{id}", h.GetGroup)
	mux.HandleFunc("POST /iddir/groups/{id}/premium", h.PayPremium_)
	mux.HandleFunc("POST /iddir/groups/{id}/claims", h.FileClaim_)
	mux.HandleFunc("GET /iddir/groups/{id}/claims", h.ListClaims)
	mux.HandleFunc("GET /iddir/groups/{id}/claims/{claimId}", h.GetClaim)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	userID := subFromBearer(r)
	if h.PayPremium == nil || h.PayPremium.Groups == nil {
		kitresp.Success(w, http.StatusOK, []any{}, meta(r))
		return
	}
	groups, err := h.PayPremium.Groups.ListByUserID(r.Context(), userID)
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	if groups == nil {
		groups = []app.GroupWithCoverage{}
	}
	kitresp.Success(w, http.StatusOK, groups, meta(r))
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string `json:"name"`
		PremiumMinor int64  `json:"premium_minor"`
		Frequency    string `json:"frequency"`
		BenefitMinor int64  `json:"benefit_minor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	leaderID := subFromBearer(r)
	id, _ := uuid.NewV7()
	group := &domain.IddirGroup{
		ID:           id,
		Name:         req.Name,
		PremiumMinor: req.PremiumMinor,
		Frequency:    req.Frequency,
		BenefitMinor: req.BenefitMinor,
		Status:       "FORMING",
		LeaderID:     leaderID,
		CreatedAt:    time.Now().UTC(),
	}
	if h.PayPremium != nil && h.PayPremium.Groups != nil {
		if err := h.PayPremium.Groups.Save(r.Context(), group); err != nil {
			kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
			return
		}
	}
	// Auto-enroll the leader as the first member
	if h.Memberships != nil {
		_ = h.Memberships.Save(r.Context(), id, leaderID) // non-fatal
	}
	kitresp.Success(w, http.StatusCreated, map[string]any{
		"group_id": id.String(),
		"status":   "FORMING",
	}, meta(r))
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	if h.PayPremium == nil || h.PayPremium.Groups == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "GROUP_UNAVAILABLE", "group repository not configured", meta(r))
		return
	}
	g, err := h.PayPremium.Groups.FindByID(r.Context(), groupID)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "group not found", meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]any{
		"group_id":       g.ID.String(),
		"name":           g.Name,
		"premium_minor":  g.PremiumMinor,
		"benefit_minor":  g.BenefitMinor,
		"frequency":      g.Frequency,
		"status":         g.Status,
		"leader_id":      g.LeaderID.String(),
	}, meta(r))
}

func (h *Handler) ListClaims(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	if h.FileClaim == nil || h.FileClaim.Claims == nil {
		kitresp.Success(w, http.StatusOK, []any{}, meta(r))
		return
	}
	memberID := subFromBearer(r)
	claims, err := h.FileClaim.Claims.ListByGroupAndMember(r.Context(), groupID, memberID)
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	if claims == nil {
		claims = []domain.Claim{}
	}
	kitresp.Success(w, http.StatusOK, claims, meta(r))
}

func (h *Handler) PayPremium_(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	var req struct {
		Period         string `json:"period"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	if req.IdempotencyKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "idempotency_key required", meta(r))
		return
	}
	memberID := subFromBearer(r)
	out, err := h.PayPremium.Execute(r.Context(), app.PayPremiumInput{
		GroupID:        groupID,
		MemberID:       memberID,
		Period:         req.Period,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) FileClaim_(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	var req struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		EvidenceRef string `json:"evidence_ref"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	memberID := subFromBearer(r)
	out, err := h.FileClaim.Execute(r.Context(), app.FileClaimInput{
		GroupID:     groupID,
		MemberID:    memberID,
		Type:        req.Type,
		Description: req.Description,
		EvidenceRef: req.EvidenceRef,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	claimID, err := uuid.FromString(r.PathValue("claimId"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid claim id", meta(r))
		return
	}
	if h.FileClaim == nil || h.FileClaim.Claims == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "CLAIM_UNAVAILABLE", "claim repository not configured", meta(r))
		return
	}
	claim, err := h.FileClaim.Claims.FindByID(r.Context(), claimID)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "claim not found", meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, claim, meta(r))
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	if de, ok := err.(*kiterr.DomainError); ok {
		status := http.StatusUnprocessableEntity
		switch de.Code {
		case "GROUP_NOT_FOUND", "NOT_FOUND":
			status = http.StatusNotFound
		case "GROUP_UNAVAILABLE", "CLAIM_UNAVAILABLE":
			status = http.StatusServiceUnavailable
		case "GROUP_INACTIVE":
			status = http.StatusConflict
		}
		kitresp.Err(w, status, de.Code, de.Message, meta(r))
		return
	}
	kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
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
