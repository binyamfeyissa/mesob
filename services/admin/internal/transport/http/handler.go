package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/admin/internal/app"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	UpdateConfig *app.UpdateConfigUseCase
	UpdateFlag   *app.UpdateFlagUseCase
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /admin/dashboard/summary",  h.DashboardSummary)
	mux.HandleFunc("GET /admin/dashboard/volume",   h.DashboardVolume)
	mux.HandleFunc("GET /admin/audit",              h.GetAudit)
	mux.HandleFunc("GET /admin/flags",              h.GetFlags)
	mux.HandleFunc("GET /admin/fraud/alerts",       h.GetFraudAlerts)
	mux.HandleFunc("PUT /admin/config/{key}",       h.UpdateConfig_)
	mux.HandleFunc("PATCH /admin/flags/{flag}",     h.UpdateFlag_)
	mux.HandleFunc("GET /admin/live",               h.LiveWebSocket)
}

func (h *Handler) DashboardSummary(w http.ResponseWriter, r *http.Request) {
	kitresp.Success(w, http.StatusOK, map[string]any{
		"users_active":       1847,
		"txn_today":          392,
		"volume_today_minor": 4823500,
		"loans_active":       63,
		"open_alerts":        2,
		"float_health":       "OK",
	}, meta(r))
}

func (h *Handler) DashboardVolume(w http.ResponseWriter, r *http.Request) {
	today := time.Now()
	labels := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	volumes := []int64{120000, 185000, 230000, 175000, 210000, 95000, 140000}
	counts := []int{48, 73, 92, 67, 84, 38, 56}

	var data []map[string]any
	for i := 6; i >= 0; i-- {
		day := today.AddDate(0, 0, -i)
		idx := int(day.Weekday()) // 0=Sun
		if idx == 0 {
			idx = 6
		} else {
			idx--
		}
		data = append(data, map[string]any{
			"date":         labels[idx],
			"volume_minor": volumes[idx],
			"count":        counts[idx],
		})
	}
	kitresp.Success(w, http.StatusOK, data, meta(r))
}

func (h *Handler) GetAudit(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	entries := []map[string]any{
		{
			"id":         "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			"actor_id":   "550e8400-e29b-41d4-a716-446655440001",
			"actor_role": "SUPER_ADMIN",
			"action":     "KYC_APPROVE",
			"target":     "user:550e8400-e29b-41d4-a716-446655440010",
			"channel":    "WEB",
			"ip":         "192.168.1.100",
			"created_at": now.Add(-5 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         "b2c3d4e5-f6a7-8901-bcde-f12345678901",
			"actor_id":   "550e8400-e29b-41d4-a716-446655440002",
			"actor_role": "ADMIN",
			"action":     "AGENT_APPROVE",
			"target":     "agent:550e8400-e29b-41d4-a716-446655440020",
			"channel":    "WEB",
			"ip":         "10.0.0.5",
			"created_at": now.Add(-30 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         "c3d4e5f6-a7b8-9012-cdef-123456789012",
			"actor_id":   "550e8400-e29b-41d4-a716-446655440001",
			"actor_role": "SUPER_ADMIN",
			"action":     "FLAG_TOGGLE",
			"target":     "flag:iqub_enabled",
			"channel":    "WEB",
			"ip":         "192.168.1.100",
			"created_at": now.Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":         "d4e5f6a7-b8c9-0123-def0-234567890123",
			"actor_id":   "550e8400-e29b-41d4-a716-446655440003",
			"actor_role": "BRANCH_MANAGER",
			"action":     "SETTLEMENT_CONFIRM",
			"target":     "settlement:d4e5f6a7-b8c9-0123-def0-234567890123",
			"channel":    "WEB",
			"ip":         "172.16.0.50",
			"created_at": now.Add(-4 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":         "e5f6a7b8-c9d0-1234-ef01-345678901234",
			"actor_id":   "550e8400-e29b-41d4-a716-446655440002",
			"actor_role": "ADMIN",
			"action":     "FRAUD_DISMISS",
			"target":     "alert:e5f6a7b8-c9d0-1234-ef01-345678901234",
			"channel":    "WEB",
			"ip":         "10.0.0.5",
			"created_at": now.Add(-6 * time.Hour).Format(time.RFC3339),
		},
	}
	kitresp.Success(w, http.StatusOK, entries, meta(r))
}

func (h *Handler) GetFlags(w http.ResponseWriter, r *http.Request) {
	flags := []map[string]any{
		{"name": "iqub_enabled", "enabled": true, "rollout_pct": 100},
		{"name": "iddir_enabled", "enabled": true, "rollout_pct": 100},
		{"name": "loans_enabled", "enabled": true, "rollout_pct": 50},
		{"name": "ussd_enabled", "enabled": true, "rollout_pct": 100},
		{"name": "telegram_bot_enabled", "enabled": false, "rollout_pct": 0},
		{"name": "biometric_kyc", "enabled": false, "rollout_pct": 0},
	}
	kitresp.Success(w, http.StatusOK, flags, meta(r))
}

func (h *Handler) GetFraudAlerts(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	alerts := []map[string]any{
		{
			"alert_id":       "f1a2b3c4-d5e6-7890-abcd-ef1234567890",
			"severity":       "HIGH",
			"rule":           "VELOCITY_BURST",
			"transaction_id": "t1a2b3c4-d5e6-7890-abcd-ef1234567890",
			"status":         "OPEN",
			"created_at":     now.Add(-10 * time.Minute).Format(time.RFC3339),
		},
		{
			"alert_id":       "f2b3c4d5-e6f7-8901-bcde-f12345678901",
			"severity":       "MEDIUM",
			"rule":           "LARGE_OUTFLOW",
			"transaction_id": "t2b3c4d5-e6f7-8901-bcde-f12345678901",
			"status":         "OPEN",
			"created_at":     now.Add(-45 * time.Minute).Format(time.RFC3339),
		},
	}
	kitresp.Success(w, http.StatusOK, alerts, meta(r))
}

func (h *Handler) UpdateConfig_(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var req struct {
		Value            json.RawMessage `json:"value"`
		SecondAuthoriser string          `json:"second_authoriser_id"`
		Reason           string          `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	actorID := subFromBearer(r)
	secondID, _ := uuid.FromString(req.SecondAuthoriser)
	out, err := h.UpdateConfig.Execute(r.Context(), app.UpdateConfigInput{
		Key: key, Value: req.Value, Reason: req.Reason,
		ActorID: actorID, SecondAuthoriser: secondID,
	})
	if err != nil {
		if de, ok := err.(*kiterr.DomainError); ok {
			status := http.StatusUnprocessableEntity
			if de.Code == "SAME_AUTHORISER" {
				status = http.StatusForbidden
			}
			kitresp.Err(w, status, de.Code, de.Message, meta(r))
			return
		}
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

// subFromBearer decodes the JWT payload (without re-verifying the signature —
// the gateway has already validated it) and returns the "sub" claim as a UUID.
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
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil
	}
	sub, _ := claims["sub"].(string)
	id, _ := uuid.FromString(sub)
	return id
}

func (h *Handler) UpdateFlag_(w http.ResponseWriter, r *http.Request) {
	flag := r.PathValue("flag")
	var req struct {
		Enabled    bool `json:"enabled"`
		RolloutPct int  `json:"rollout_pct"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	actorID := subFromBearer(r)
	out, err := h.UpdateFlag.Execute(r.Context(), app.UpdateFlagInput{
		Flag: flag, Enabled: req.Enabled, RolloutPct: req.RolloutPct, ActorID: actorID,
	})
	if err != nil {
		kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

// LiveWebSocket implements a Server-Sent Events stream for the admin live dashboard.
// Clients connect with Accept: text/event-stream; no WebSocket upgrade required.
func (h *Handler) LiveWebSocket(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	sendEvent := func(eventType string, data any) {
		payload, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, payload)
		flusher.Flush()
	}

	// Send an initial snapshot immediately.
	sendEvent("summary", map[string]any{
		"users_active":       1847,
		"txn_today":          392,
		"volume_today_minor": 4823500,
		"loans_active":       63,
		"open_alerts":        2,
		"float_health":       "OK",
		"ts":                 time.Now().UTC().Format(time.RFC3339),
	})

	for {
		select {
		case <-r.Context().Done():
			return
		case t := <-ticker.C:
			sendEvent("heartbeat", map[string]any{"ts": t.UTC().Format(time.RFC3339)})
		}
	}
}

func meta(r *http.Request) kitresp.MetaBlock {
	return kitresp.MetaBlock{RequestID: r.Header.Get("X-Request-ID"), TraceID: r.Header.Get("X-Trace-ID")}
}
