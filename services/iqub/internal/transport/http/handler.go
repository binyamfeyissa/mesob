package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/iqub/internal/app"
	"github.com/mesob-wallet/iqub/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	CreateGroup *app.CreateGroupUseCase
	Contribute  *app.ContributeUseCase
	Cycles      app.CycleRepository
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /iqub/groups", h.CreateGroup_)
	mux.HandleFunc("GET /iqub/groups/{id}/members", h.ListMembers)
	mux.HandleFunc("POST /iqub/groups/{id}/members", h.JoinGroup)
	mux.HandleFunc("POST /iqub/groups/{id}/contribute", h.Contribute_)
	mux.HandleFunc("GET /iqub/groups/{id}", h.GetGroup)
	mux.HandleFunc("GET /iqub/groups", h.ListGroups)
	mux.HandleFunc("POST /iqub/cycles/{id}/close", h.CloseCycle)
}

func (h *Handler) CreateGroup_(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		CycleMinor  int64  `json:"cycle_minor"`
		Frequency   string `json:"frequency"`
		MemberLimit int    `json:"member_limit"`
		PayoutOrder string `json:"payout_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	leaderID := subFromBearer(r)
	out, err := h.CreateGroup.Execute(r.Context(), app.CreateGroupInput{
		Name:        req.Name,
		CycleMinor:  req.CycleMinor,
		Frequency:   req.Frequency,
		MemberLimit: req.MemberLimit,
		PayoutOrder: req.PayoutOrder,
		LeaderID:    leaderID,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	if h.Contribute == nil || h.Contribute.Memberships == nil {
		kitresp.Success(w, http.StatusOK, []any{}, meta(r))
		return
	}
	memberships, err := h.Contribute.Memberships.ListByGroup(r.Context(), groupID)
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	rows := make([]app.MemberListRow, 0, len(memberships))
	for _, m := range memberships {
		rows = append(rows, app.MemberListRow{
			MembershipID: m.ID.String(),
			UserID:       m.UserID.String(),
			PayoutOrder:  m.PayoutOrder,
			CycleState:   string(m.CycleState),
		})
	}
	kitresp.Success(w, http.StatusOK, rows, meta(r))
}

func (h *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("id")
	groupID, err := uuid.FromString(groupIDStr)
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	var req struct {
		JoinCode string `json:"join_code"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	userID := subFromBearer(r)

	if h.CreateGroup == nil || h.CreateGroup.Groups == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "GROUP_UNAVAILABLE", "group repository not configured", meta(r))
		return
	}
	group, err := h.CreateGroup.Groups.FindByID(r.Context(), groupID)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "GROUP_NOT_FOUND", "group not found", meta(r))
		return
	}
	if req.JoinCode != "" && group.JoinCode != req.JoinCode {
		kitresp.Err(w, http.StatusForbidden, "INVALID_JOIN_CODE", "join code does not match", meta(r))
		return
	}

	memberID, _ := uuid.NewV7()
	membership := &domain.Membership{
		ID:         memberID,
		GroupID:    groupID,
		UserID:     userID,
		CycleState: domain.CycleStatePending,
		JoinedAt:   time.Now().UTC(),
	}
	if h.Contribute != nil && h.Contribute.Memberships != nil {
		if err := h.Contribute.Memberships.Save(r.Context(), membership); err != nil {
			kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
			return
		}
	}

	kitresp.Success(w, http.StatusCreated, map[string]any{
		"membership_id": memberID.String(),
		"group_id":      groupID.String(),
		"status":        string(domain.CycleStatePending),
	}, meta(r))
}

func (h *Handler) Contribute_(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	var req struct {
		CycleID        string `json:"cycle_id"`
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
	cycleID, _ := uuid.FromString(req.CycleID)
	userID := subFromBearer(r)
	out, err := h.Contribute.Execute(r.Context(), app.ContributeInput{
		GroupID:        groupID,
		UserID:         userID,
		CycleID:        cycleID,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid group id", meta(r))
		return
	}
	if h.CreateGroup == nil || h.CreateGroup.Groups == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "GROUP_UNAVAILABLE", "group repository not configured", meta(r))
		return
	}
	g, err := h.CreateGroup.Groups.FindByID(r.Context(), groupID)
	if err != nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "group not found", meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]any{
		"group_id":        g.ID.String(),
		"name":            g.Name,
		"cycle_minor":     g.CycleMinor,
		"frequency":       g.Frequency,
		"member_limit":    g.MemberLimit,
		"payout_order":    g.PayoutOrder,
		"status":          string(g.Status),
		"leader_id":       g.LeaderID.String(),
		"join_code":       g.JoinCode,
	}, meta(r))
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	userID := subFromBearer(r)
	if h.CreateGroup == nil || h.CreateGroup.Groups == nil {
		kitresp.Success(w, http.StatusOK, []any{}, meta(r))
		return
	}
	groups, err := h.CreateGroup.Groups.ListByUserID(r.Context(), userID)
	if err != nil {
		mapError(w, r, err)
		return
	}
	if groups == nil {
		groups = []app.GroupWithCycleInfo{}
	}
	kitresp.Success(w, http.StatusOK, groups, meta(r))
}

func (h *Handler) CloseCycle(w http.ResponseWriter, r *http.Request) {
	cycleID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid cycle id", meta(r))
		return
	}
	if h.Cycles == nil {
		kitresp.Err(w, http.StatusServiceUnavailable, "CYCLE_UNAVAILABLE", "cycle repository not configured", meta(r))
		return
	}
	if err := h.Cycles.Close(r.Context(), cycleID); err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, map[string]string{"status": "CLOSED"}, meta(r))
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	if de, ok := err.(*kiterr.DomainError); ok {
		status := http.StatusUnprocessableEntity
		switch de.Code {
		case "NOT_FOUND", "GROUP_NOT_FOUND":
			status = http.StatusNotFound
		case "GROUP_UNAVAILABLE":
			status = http.StatusServiceUnavailable
		case "NOT_MEMBER":
			status = http.StatusForbidden
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
