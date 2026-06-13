package app

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/ledger/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type EntryInput struct {
	AccountID   string `json:"account_id"`
	Direction   string `json:"direction"`
	AmountMinor int64  `json:"amount_minor"`
}

type PostTransactionInput struct {
	IdempotencyKey string
	Type           string
	InitiatedBy    string
	Channel        string
	Entries        []EntryInput
}

type BalanceResult struct {
	AccountID    string `json:"account_id"`
	BalanceMinor int64  `json:"balance_minor"`
	Currency     string `json:"currency"`
}

type PostTransactionOutput struct {
	TransactionID string          `json:"transaction_id"`
	Status        string          `json:"status"`
	PostedAt      string          `json:"posted_at"`
	Balances      []BalanceResult `json:"balances"`
}

type PostTransactionUseCase struct {
	Accounts     AccountRepository
	Transactions TransactionRepository
	Events       EventPublisher
}

func (uc *PostTransactionUseCase) Execute(ctx context.Context, in PostTransactionInput) (*PostTransactionOutput, error) {
	// 1. Idempotency check — return cached response if key already exists.
	existing, err := uc.Transactions.FindByIdempotencyKey(ctx, in.IdempotencyKey)
	if err != nil && !errors.Is(err, kiterr.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return buildPostOutput(existing, nil), nil
	}

	// 2. Convert EntryInputs to domain.Entry, validate account IDs.
	entries := make([]domain.Entry, len(in.Entries))
	for i, e := range in.Entries {
		aid, parseErr := uuid.FromString(e.AccountID)
		if parseErr != nil {
			return nil, parseErr
		}
		entries[i] = domain.Entry{
			AccountID:   aid,
			Direction:   domain.Direction(e.Direction),
			AmountMinor: e.AmountMinor,
		}
	}

	// 3. Balance check for debit entries.
	for _, e := range entries {
		if e.Direction != domain.Debit {
			continue
		}
		acc, accErr := uc.Accounts.FindByID(ctx, e.AccountID)
		if accErr != nil {
			return nil, accErr
		}
		if acc.AllowNegative {
			continue
		}
		bal, balErr := uc.Accounts.GetBalance(ctx, e.AccountID)
		if balErr != nil {
			return nil, balErr
		}
		if bal < e.AmountMinor {
			return nil, kiterr.ErrInsufficientBalance
		}
	}

	// 4. Build balanced transaction domain object.
	txn, err := domain.NewBalancedTransaction(in.Type, in.IdempotencyKey, in.InitiatedBy, in.Channel, entries)
	if err != nil {
		return nil, err
	}

	// 5. Persist transaction, entries, balances and idempotency key atomically.
	if err := uc.Transactions.Save(ctx, txn); err != nil {
		return nil, err
	}

	// 6. Publish domain event (non-fatal — outbox relay will retry).
	_ = uc.Events.Publish(ctx, "TransactionPosted", txn.ID.String(), map[string]any{
		"transaction_id":  txn.ID.String(),
		"type":            txn.Type,
		"status":          txn.Status,
		"idempotency_key": txn.IdempotencyKey,
		"initiated_by":    txn.InitiatedBy.String(),
		"channel":         txn.Channel,
	})

	// 7. Fetch updated balances for all accounts touched by this transaction.
	seen := make(map[uuid.UUID]bool)
	var balances []BalanceResult
	for _, e := range entries {
		if seen[e.AccountID] {
			continue
		}
		seen[e.AccountID] = true
		bal, balErr := uc.Accounts.GetBalance(ctx, e.AccountID)
		if balErr != nil {
			continue // balance fetch is best-effort in the response
		}
		acc, accErr := uc.Accounts.FindByID(ctx, e.AccountID)
		currency := ""
		if accErr == nil {
			currency = acc.Currency
		}
		balances = append(balances, BalanceResult{
			AccountID:    e.AccountID.String(),
			BalanceMinor: bal,
			Currency:     currency,
		})
	}

	return buildPostOutput(txn, balances), nil
}

func buildPostOutput(txn *domain.Transaction, balances []BalanceResult) *PostTransactionOutput {
	return &PostTransactionOutput{
		TransactionID: txn.ID.String(),
		Status:        txn.Status,
		PostedAt:      txn.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Balances:      balances,
	}
}
