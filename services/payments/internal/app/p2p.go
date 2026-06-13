package app

import (
	"context"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type P2PInput struct {
	PayerID        uuid.UUID
	PayerAccountID string
	ToMSISDN       string
	AmountMinor    int64
	Note           string
	IdempotencyKey string
}

type P2POutput struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

type P2PUseCase struct {
	Fraud    FraudClient
	Identity IdentityClient
	Ledger   LedgerClient
	Events   EventPublisher
}

func (uc *P2PUseCase) Execute(ctx context.Context, in P2PInput) (*P2POutput, error) {
	// 1. Find payee by MSISDN via IdentityClient
	if uc.Identity == nil {
		return nil, &kiterr.DomainError{Code: "IDENTITY_UNAVAILABLE", Message: "identity service not configured"}
	}
	_, payeeAccountID, err := uc.Identity.FindUserByMSISDN(ctx, in.ToMSISDN)
	if err != nil {
		if kiterr.Is(err, kiterr.ErrNotFound) {
			return nil, &kiterr.DomainError{Code: "PAYEE_NOT_FOUND", Message: "payee not found"}
		}
		return nil, &kiterr.DomainError{Code: "IDENTITY_UNAVAILABLE", Message: err.Error()}
	}

	// 2. Fraud screen (fail-closed)
	if uc.Fraud != nil {
		decision, _, _, err := uc.Fraud.Screen(ctx, in.PayerID.String(), "P2P", in.AmountMinor, in.ToMSISDN, "APP")
		if err != nil {
			return nil, kiterr.ErrFraudUnavailable
		}
		if decision == "BLOCK" {
			return nil, kiterr.ErrFraudBlocked
		}
	}

	// 3. Post ledger: payer_wallet(D) → payee_wallet(C)
	if uc.Ledger == nil {
		return nil, &kiterr.DomainError{Code: "LEDGER_UNAVAILABLE", Message: "ledger not configured"}
	}
	txnID, err := uc.Ledger.PostTransaction(ctx, "P2P", in.IdempotencyKey, in.PayerID.String(), "APP", []LedgerEntry{
		{AccountID: in.PayerAccountID, Direction: "D", AmountMinor: in.AmountMinor},
		{AccountID: payeeAccountID, Direction: "C", AmountMinor: in.AmountMinor},
	})
	if err != nil {
		return nil, err
	}

	// 4. Publish PaymentCompleted
	if uc.Events != nil {
		uc.Events.Publish(ctx, "PaymentCompleted", txnID, map[string]any{
			"type":         "P2P",
			"payer":        in.PayerID.String(),
			"payee_msisdn": in.ToMSISDN,
			"amount_minor": in.AmountMinor,
		})
	}

	return &P2POutput{TransactionID: txnID, Status: "COMPLETED"}, nil
}
