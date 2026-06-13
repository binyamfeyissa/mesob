package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/loans/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type ApplyLoanInput struct {
	UserID         uuid.UUID
	AmountMinor    int64
	TermDays       int
	Purpose        string
	IdempotencyKey string
}

type ApplyLoanOutput struct {
	LoanID            string   `json:"loan_id"`
	Decision          string   `json:"decision"`
	ScoreID           string   `json:"score_id,omitempty"`
	PrincipalMinor    int64    `json:"principal_minor,omitempty"`
	FeeMinor          int64    `json:"fee_minor,omitempty"`
	DueDate           string   `json:"due_date,omitempty"`
	DisbursementTxnID string   `json:"disbursement_txn_id,omitempty"`
	Mode              string   `json:"mode,omitempty"`
	Reasons           []string `json:"reasons,omitempty"`
	CeilingMinor      int64    `json:"ceiling_minor,omitempty"`
}

type ApplyLoanUseCase struct {
	Loans   LoanRepository
	Scoring ScoringClient
	Ledger  LedgerClient
	MFI     MFIAdapterClient
	Events  EventPublisher
}

func (uc *ApplyLoanUseCase) Execute(ctx context.Context, in ApplyLoanInput) (*ApplyLoanOutput, error) {
	// 1. Get credit score
	cs, err := uc.Scoring.Score(ctx, in.UserID.String(), false)
	if err != nil {
		return nil, err
	}

	// 2. Check amount <= ceiling
	if in.AmountMinor > cs.CeilingMinor {
		return &ApplyLoanOutput{
			LoanID:       "",
			Decision:     "DECLINED",
			ScoreID:      cs.ScoreID,
			Reasons:      []string{"requested amount exceeds credit ceiling"},
			CeilingMinor: cs.CeilingMinor,
		}, nil
	}

	// 3. Check no active loans (if repo available)
	if uc.Loans != nil {
		existing, _ := uc.Loans.ListByUser(ctx, in.UserID)
		for _, l := range existing {
			if l.IsActive() {
				return &ApplyLoanOutput{
					Decision: "DECLINED",
					Reasons:  []string{"existing active loan"},
				}, nil
			}
		}
	}

	// 4. Originate with MFI
	var facilityID string
	if uc.MFI != nil {
		facilityID, err = uc.MFI.Originate(ctx, in.UserID.String(), in.AmountMinor, in.TermDays, cs.ScoreID, in.IdempotencyKey)
		if err != nil {
			return nil, kiterr.ErrProviderUnavailable
		}
	}

	// 5. Calculate fee (5% of principal)
	feeMinor := in.AmountMinor / 20
	dueDate := time.Now().AddDate(0, 0, in.TermDays)

	// 6. Post ledger: MFI_clearing(C) → user_wallet(D)
	var disbTxnID string
	if uc.Ledger != nil {
		disbTxnID, _ = uc.Ledger.PostTransaction(ctx, "LOAN_DISBURSEMENT", in.IdempotencyKey,
			in.UserID.String(), "SYSTEM", []LedgerEntry{
				{AccountID: "mfi-clearing", Direction: "C", AmountMinor: in.AmountMinor + feeMinor},
				{AccountID: in.UserID.String() + "-wallet", Direction: "D", AmountMinor: in.AmountMinor},
				{AccountID: "mesob-fee-pool", Direction: "D", AmountMinor: feeMinor},
			})
	}

	// 7. Save loan
	loanID, _ := uuid.NewV7()
	scoreUUID, _ := uuid.FromString(cs.ScoreID)
	loan := &domain.Loan{
		ID:               loanID,
		UserID:           in.UserID,
		PrincipalMinor:   in.AmountMinor,
		FeeMinor:         feeMinor,
		OutstandingMinor: in.AmountMinor + feeMinor,
		ScoreID:          &scoreUUID,
		Status:           domain.LoanActive,
		Mode:             "DIGITAL",
		DueDate:          dueDate,
		MFIFacilityID:    facilityID,
		CreatedAt:        time.Now().UTC(),
	}
	if uc.Loans != nil {
		uc.Loans.Save(ctx, loan)
	}

	// 8. Publish
	if uc.Events != nil {
		uc.Events.Publish(ctx, "LoanDisbursed", loanID.String(), map[string]any{
			"amount_minor": in.AmountMinor,
		})
	}

	return &ApplyLoanOutput{
		LoanID:            loanID.String(),
		Decision:          "APPROVED",
		ScoreID:           cs.ScoreID,
		PrincipalMinor:    in.AmountMinor,
		FeeMinor:          feeMinor,
		DueDate:           dueDate.Format("2006-01-02"),
		DisbursementTxnID: disbTxnID,
		Mode:              "DIGITAL",
	}, nil
}
