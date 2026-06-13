package app

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
)

type RepayLoanInput struct {
	LoanID         uuid.UUID
	UserID         uuid.UUID
	AmountMinor    int64
	IdempotencyKey string
}

type RepayLoanOutput struct {
	LoanID           string `json:"loan_id"`
	OutstandingMinor int64  `json:"outstanding_minor"`
	Status           string `json:"status"`
}

type RepayLoanUseCase struct {
	Loans  LoanRepository
	Ledger LedgerClient
	Events EventPublisher
}

func (uc *RepayLoanUseCase) Execute(ctx context.Context, in RepayLoanInput) (*RepayLoanOutput, error) {
	loan, err := uc.Loans.FindByID(ctx, in.LoanID)
	if err != nil {
		return nil, err
	}
	if loan.UserID != in.UserID && in.UserID != (uuid.UUID{}) {
		return nil, &notFoundError{in.LoanID.String()}
	}
	if loan.Status != "ACTIVE" && loan.Status != "OVERDUE" {
		return nil, &invalidStateError{string(loan.Status)}
	}

	if uc.Ledger != nil {
		_, err = uc.Ledger.PostTransaction(ctx, "LOAN_REPAYMENT", in.IdempotencyKey,
			in.UserID.String(), "APP", []LedgerEntry{
				{AccountID: in.UserID.String() + "-wallet", Direction: "D", AmountMinor: in.AmountMinor},
				{AccountID: "mfi-clearing", Direction: "C", AmountMinor: in.AmountMinor},
			})
		if err != nil {
			return nil, err
		}
	}

	loan.ApplyRepayment(in.AmountMinor)
	if err := uc.Loans.Update(ctx, loan); err != nil {
		return nil, err
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "LoanRepaid", loan.ID.String(), map[string]any{
			"amount_minor":     in.AmountMinor,
			"outstanding_minor": loan.OutstandingMinor,
			"status":           string(loan.Status),
		})
	}

	return &RepayLoanOutput{
		LoanID:           loan.ID.String(),
		OutstandingMinor: loan.OutstandingMinor,
		Status:           string(loan.Status),
	}, nil
}

type notFoundError struct{ id string }
func (e *notFoundError) Error() string { return "not found: " + e.id }

type invalidStateError struct{ status string }
func (e *invalidStateError) Error() string { return fmt.Sprintf("loan cannot be repaid in state %s", e.status) }
