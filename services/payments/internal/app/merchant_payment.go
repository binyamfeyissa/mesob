package app

import (
	"context"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type MerchantPaymentInput struct {
	PayerID        uuid.UUID
	PayerAccountID string
	MerchantID     uuid.UUID
	AmountMinor    int64
	Ref            string
	IdempotencyKey string
}

type MerchantPaymentOutput struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	ReceiptRef    string `json:"receipt_ref"`
}

type MerchantPaymentUseCase struct {
	Merchants MerchantRepository
	Ledger    LedgerClient
	Events    EventPublisher
}

func (uc *MerchantPaymentUseCase) Execute(ctx context.Context, in MerchantPaymentInput) (*MerchantPaymentOutput, error) {
	// 1. Find merchant
	if uc.Merchants == nil {
		return nil, &kiterr.DomainError{Code: "MERCHANT_UNAVAILABLE", Message: "merchant repo not configured"}
	}
	merchant, err := uc.Merchants.FindByID(ctx, in.MerchantID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "MERCHANT_NOT_FOUND", Message: "merchant not found"}
	}

	// 2. Calculate commission
	commissionMinor := int64(float64(in.AmountMinor) * merchant.CommissionPct / 100.0)
	merchantNetMinor := in.AmountMinor - commissionMinor

	// 3. Post ledger
	receiptRef := in.IdempotencyKey
	if uc.Ledger != nil {
		entries := []LedgerEntry{
			{AccountID: in.PayerAccountID, Direction: "D", AmountMinor: in.AmountMinor},
			{AccountID: merchant.AccountID, Direction: "C", AmountMinor: merchantNetMinor},
		}
		if commissionMinor > 0 {
			entries = append(entries, LedgerEntry{AccountID: "mesob-fee-pool", Direction: "C", AmountMinor: commissionMinor})
		}
		txnID, err := uc.Ledger.PostTransaction(ctx, "MERCHANT_PAYMENT", in.IdempotencyKey, in.PayerID.String(), "APP", entries)
		if err != nil {
			return nil, err
		}
		receiptRef = txnID
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "PaymentCompleted", receiptRef, map[string]any{
			"type":         "MERCHANT",
			"payer":        in.PayerID.String(),
			"merchant":     merchant.ID.String(),
			"amount_minor": in.AmountMinor,
		})
	}

	return &MerchantPaymentOutput{TransactionID: receiptRef, Status: "COMPLETED", ReceiptRef: receiptRef}, nil
}
