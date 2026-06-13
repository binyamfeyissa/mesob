package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/loans/internal/app"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	CheckEligibility *app.CheckEligibilityUseCase
	ApplyLoan        *app.ApplyLoanUseCase
	RepayLoan        *app.RepayLoanUseCase
	Loans            app.LoanRepository
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /loans/eligibility", h.GetEligibility)
	mux.HandleFunc("POST /loans/apply", h.Apply_)
	mux.HandleFunc("POST /loans/{id}/repay", h.Repay_)
	mux.HandleFunc("GET /loans", h.ListLoans)
}

func (h *Handler) GetEligibility(w http.ResponseWriter, r *http.Request) {
	userID := subFromBearer(r)
	out, err := h.CheckEligibility.Execute(r.Context(), userID)
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) Apply_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	var req struct {
		AmountMinor int64  `json:"amount_minor"`
		TermDays    int    `json:"term_days"`
		Purpose     string `json:"purpose"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	userID := subFromBearer(r)
	out, err := h.ApplyLoan.Execute(r.Context(), app.ApplyLoanInput{
		UserID:         userID,
		AmountMinor:    req.AmountMinor,
		TermDays:       req.TermDays,
		Purpose:        req.Purpose,
		IdempotencyKey: idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	status := http.StatusCreated
	if out != nil && out.Decision == "DECLINED" {
		status = http.StatusOK
	}
	kitresp.Success(w, status, out, meta(r))
}

func (h *Handler) Repay_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	loanID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid loan id", meta(r))
		return
	}
	var req struct {
		AmountMinor int64 `json:"amount_minor"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	userID := subFromBearer(r)
	out, err := h.RepayLoan.Execute(r.Context(), app.RepayLoanInput{
		LoanID:         loanID,
		UserID:         userID,
		AmountMinor:    req.AmountMinor,
		IdempotencyKey: idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

func (h *Handler) ListLoans(w http.ResponseWriter, r *http.Request) {
	userID := subFromBearer(r)
	if h.Loans == nil {
		kitresp.Success(w, http.StatusOK, []any{}, meta(r))
		return
	}
	loans, err := h.Loans.ListByUser(r.Context(), userID)
	if err != nil {
		mapError(w, r, err)
		return
	}
	if loans == nil {
		kitresp.Success(w, http.StatusOK, []any{}, meta(r))
		return
	}
	kitresp.Success(w, http.StatusOK, loans, meta(r))
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	if de, ok := err.(*kiterr.DomainError); ok {
		status := http.StatusUnprocessableEntity
		switch de.Code {
		case kiterr.ErrNotFound.Code:
			status = http.StatusNotFound
		case kiterr.ErrInsufficientHistory.Code:
			status = http.StatusUnprocessableEntity
		case kiterr.ErrScoringDeferred.Code:
			status = http.StatusServiceUnavailable
		case kiterr.ErrProviderUnavailable.Code:
			status = http.StatusServiceUnavailable
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
