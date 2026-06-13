package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/payments/internal/app"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	P2P             *app.P2PUseCase
	MerchantPayment *app.MerchantPaymentUseCase
	BillPayment     *app.BillPaymentUseCase
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /payments/p2p",           h.P2P_)
	mux.HandleFunc("POST /payments/merchant",      h.Merchant_)
	mux.HandleFunc("POST /payments/bill",          h.Bill)
	mux.HandleFunc("GET /payments/merchants/{id}", h.GetMerchant)
}

func (h *Handler) P2P_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	var req struct {
		ToMSISDN    string `json:"to_msisdn"`
		AmountMinor int64  `json:"amount_minor"`
		Note        string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	out, err := h.P2P.Execute(r.Context(), app.P2PInput{
		ToMSISDN: req.ToMSISDN, AmountMinor: req.AmountMinor,
		Note: req.Note, IdempotencyKey: idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) Merchant_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	var req struct {
		MerchantID  string `json:"merchant_id"`
		AmountMinor int64  `json:"amount_minor"`
		Ref         string `json:"ref"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	merchantID, err := uuid.FromString(req.MerchantID)
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant id", meta(r))
		return
	}
	out, err := h.MerchantPayment.Execute(r.Context(), app.MerchantPaymentInput{
		MerchantID: merchantID, AmountMinor: req.AmountMinor,
		Ref: req.Ref, IdempotencyKey: idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) Bill(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key required", meta(r))
		return
	}
	var req struct {
		BillerID    string `json:"biller_id"`
		AccountRef  string `json:"account_ref"`
		AmountMinor int64  `json:"amount_minor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	// Extract userID from JWT Bearer token
	userID := subFromBearer(r)
	out, err := h.BillPayment.Execute(r.Context(), app.BillPaymentInput{
		UserID:      userID.String(),
		BillerID:    req.BillerID,
		AccountRef:  req.AccountRef,
		AmountMinor: req.AmountMinor,
		IdemKey:     idemKey,
	})
	if err != nil {
		mapError(w, r, err)
		return
	}
	status := http.StatusAccepted
	if out.Status == "COMPLETED" {
		status = http.StatusCreated
	}
	kitresp.Success(w, status, map[string]any{
		"transaction_id": out.TransactionID,
		"status":         out.Status,
		"biller_ref":     out.BillerRef,
	}, meta(r))
}

func (h *Handler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid merchant id", meta(r))
		return
	}
	if h.MerchantPayment == nil || h.MerchantPayment.Merchants == nil {
		kitresp.Err(w, http.StatusNotFound, "NOT_FOUND", "merchant not found", meta(r))
		return
	}
	merchant, err := h.MerchantPayment.Merchants.FindByID(r.Context(), id)
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, merchant, meta(r))
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	if de, ok := err.(*kiterr.DomainError); ok {
		status := http.StatusUnprocessableEntity
		switch de.Code {
		case "FRAUD_BLOCKED":
			status = http.StatusForbidden
		case "FRAUD_UNAVAILABLE", "SCORING_DEFERRED":
			status = http.StatusServiceUnavailable
		case "NOT_FOUND", "PAYEE_NOT_FOUND", "MERCHANT_NOT_FOUND", "BILLER_NOT_FOUND":
			status = http.StatusNotFound
		case "IDENTITY_UNAVAILABLE", "LEDGER_UNAVAILABLE", "BILLER_UNAVAILABLE", "MERCHANT_UNAVAILABLE":
			status = http.StatusServiceUnavailable
		}
		kitresp.Err(w, status, de.Code, de.Message, meta(r))
		return
	}
	kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
}

func meta(r *http.Request) kitresp.MetaBlock {
	return kitresp.MetaBlock{RequestID: r.Header.Get("X-Request-ID"), TraceID: r.Header.Get("X-Trace-ID")}
}

// subFromBearer decodes a JWT Bearer token without verification and extracts the sub claim.
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
