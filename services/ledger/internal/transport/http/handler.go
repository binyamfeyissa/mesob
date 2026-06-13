package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/ledger/internal/app"
	"github.com/mesob-wallet/ledger/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	kitresp "github.com/mesob-wallet/go-kit/response"
)

type Handler struct {
	CreateAccount   *app.CreateAccountUseCase
	PostTransaction *app.PostTransactionUseCase
	GetBalance      *app.GetBalanceUseCase
	Entries         app.EntryRepository
	Transactions    app.TransactionRepository
}

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /ledger/accounts", h.CreateAccount_)
	mux.HandleFunc("POST /ledger/transactions", h.PostTransaction_)
	mux.HandleFunc("POST /ledger/transactions/{id}/reverse", h.ReverseTransaction)
	mux.HandleFunc("GET /ledger/accounts/{id}/balance", h.GetBalance_)
	mux.HandleFunc("GET /ledger/accounts/{id}/entries", h.GetEntries)
}

func (h *Handler) CreateAccount_(w http.ResponseWriter, r *http.Request) {
	var req app.CreateAccountInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	out, err := h.CreateAccount.Execute(r.Context(), req)
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) PostTransaction_(w http.ResponseWriter, r *http.Request) {
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header required", meta(r))
		return
	}
	var req app.PostTransactionInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), meta(r))
		return
	}
	req.IdempotencyKey = idemKey
	out, err := h.PostTransaction.Execute(r.Context(), req)
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

// ReverseTransaction creates a reversal transaction that swaps the direction of
// every entry in the original transaction.
func (h *Handler) ReverseTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	txnID, err := uuid.FromString(idStr)
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid transaction id", meta(r))
		return
	}

	// Load original transaction.
	orig, err := h.Transactions.FindByID(r.Context(), txnID)
	if err != nil {
		mapError(w, r, err)
		return
	}

	// Require an idempotency key for the reversal.
	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		kitresp.Err(w, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header required", meta(r))
		return
	}

	// Build reversed entries (swap D <-> C).
	reversedEntries := make([]app.EntryInput, len(orig.Entries))
	for i, e := range orig.Entries {
		dir := string(domain.Credit)
		if e.Direction == domain.Credit {
			dir = string(domain.Debit)
		}
		reversedEntries[i] = app.EntryInput{
			AccountID:   e.AccountID.String(),
			Direction:   dir,
			AmountMinor: e.AmountMinor,
		}
	}

	in := app.PostTransactionInput{
		IdempotencyKey: idemKey,
		Type:           "REVERSAL",
		InitiatedBy:    orig.InitiatedBy.String(),
		Channel:        orig.Channel,
		Entries:        reversedEntries,
	}

	out, err := h.PostTransaction.Execute(r.Context(), in)
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusCreated, out, meta(r))
}

func (h *Handler) GetBalance_(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid account id", meta(r))
		return
	}
	out, err := h.GetBalance.Execute(r.Context(), id)
	if err != nil {
		mapError(w, r, err)
		return
	}
	kitresp.Success(w, http.StatusOK, out, meta(r))
}

// GetEntries returns a paginated list of ledger entries for an account.
// Query params: limit (default 20, max 100), cursor (opaque string).
func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	accountID, err := uuid.FromString(idStr)
	if err != nil {
		kitresp.Err(w, http.StatusBadRequest, "INVALID_ID", "invalid account id", meta(r))
		return
	}

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, parseErr := strconv.Atoi(lStr); parseErr == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	cursor := r.URL.Query().Get("cursor")

	entries, nextCursor, err := h.Entries.ListByAccount(r.Context(), accountID, limit, cursor)
	if err != nil {
		mapError(w, r, err)
		return
	}

	m := meta(r)
	m.NextCursor = nextCursor
	kitresp.Success(w, http.StatusOK, entries, m)
}

func mapError(w http.ResponseWriter, r *http.Request, err error) {
	var de *kiterr.DomainError
	if errors.As(err, &de) {
		switch de.Code {
		case kiterr.ErrNotFound.Code:
			kitresp.Err(w, http.StatusNotFound, de.Code, de.Message, meta(r))
		case kiterr.ErrInsufficientBalance.Code:
			kitresp.Err(w, http.StatusUnprocessableEntity, de.Code, de.Message, meta(r))
		case kiterr.ErrUnbalanced.Code:
			kitresp.Err(w, http.StatusBadRequest, de.Code, de.Message, meta(r))
		default:
			if de.Retryable {
				kitresp.Err(w, http.StatusServiceUnavailable, de.Code, de.Message, meta(r))
			} else {
				kitresp.Err(w, http.StatusUnprocessableEntity, de.Code, de.Message, meta(r))
			}
		}
		return
	}
	kitresp.Err(w, http.StatusInternalServerError, "INTERNAL", err.Error(), meta(r))
}

func meta(r *http.Request) kitresp.MetaBlock {
	return kitresp.MetaBlock{
		RequestID: r.Header.Get("X-Request-ID"),
		TraceID:   r.Header.Get("X-Trace-ID"),
	}
}
