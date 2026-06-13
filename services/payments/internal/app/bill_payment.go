package app

import (
	"context"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type BillPaymentUseCase struct {
	Payments      PaymentRepository
	Billers       BillerRepository
	LedgerClient  LedgerClient
	BillerAdapter BillerAdapterClient
	Events        EventPublisher
}

type BillPaymentInput struct {
	UserID      string
	BillerID    string
	AccountRef  string
	AmountMinor int64
	IdemKey     string
}

type BillPaymentOutput struct {
	TransactionID string
	Status        string // PENDING | COMPLETED
	BillerRef     string
}

// BillerRepository looks up biller configuration.
type BillerRepository interface {
	FindByID(ctx context.Context, id string) (*Biller, error)
}

// BillerAdapterClient routes the payment to the external biller.
// Returns a biller reference and initial status.
// Status stays PENDING until the biller sends a webhook confirmation.
// This is never optimistically completed.
type BillerAdapterClient interface {
	Submit(ctx context.Context, billerKey, accountRef string, amountMinor int64) (billerRef string, err error)
}

type Biller struct {
	ID         string
	Name       string
	AdapterKey string
	Status     string
}

func (uc *BillPaymentUseCase) Execute(ctx context.Context, in BillPaymentInput) (*BillPaymentOutput, error) {
	if in.AmountMinor <= 0 {
		return nil, &kiterr.DomainError{Code: "INVALID_AMOUNT", Message: "amount_minor must be positive"}
	}

	// 1. Look up biller
	if uc.Billers == nil {
		return nil, &kiterr.DomainError{Code: "BILLER_UNAVAILABLE", Message: "biller repo not configured"}
	}
	biller, err := uc.Billers.FindByID(ctx, in.BillerID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "BILLER_NOT_FOUND", Message: "biller not found"}
	}
	if biller.Status != "ACTIVE" {
		return nil, &kiterr.DomainError{Code: "BILLER_INACTIVE", Message: "biller not active"}
	}

	// 2. Submit to biller adapter
	var billerRef string
	if uc.BillerAdapter != nil {
		billerRef, err = uc.BillerAdapter.Submit(ctx, biller.AdapterKey, in.AccountRef, in.AmountMinor)
		if err != nil {
			return nil, kiterr.ErrProviderUnavailable
		}
	} else {
		billerRef = "demo-" + in.IdemKey
	}

	// 3. Post ledger (if available)
	var txnID string
	if uc.LedgerClient != nil {
		txnID, err = uc.LedgerClient.PostTransaction(ctx, "BILL_PAYMENT", in.IdemKey, in.UserID, "APP", []LedgerEntry{
			{AccountID: "TODO-user-wallet", Direction: "D", AmountMinor: in.AmountMinor},
		})
		if err != nil {
			return nil, err
		}
	} else {
		id, _ := uuid.NewV7()
		txnID = id.String()
	}

	// 4. Save payment ref
	if uc.Payments != nil {
		uc.Payments.SaveRef(ctx, "BILL", txnID, billerRef, "PENDING")
	}

	return &BillPaymentOutput{TransactionID: txnID, Status: "PENDING", BillerRef: billerRef}, nil
}
