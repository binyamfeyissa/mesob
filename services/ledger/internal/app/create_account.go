package app

import (
	"context"

	"github.com/mesob-wallet/ledger/internal/domain"
)

type CreateAccountInput struct {
	OwnerType string `json:"owner_type"`
	OwnerID   string `json:"owner_id"`
	Type      string `json:"type"`
	Currency  string `json:"currency"`
}

type CreateAccountOutput struct {
	AccountID string `json:"account_id"`
	Status    string `json:"status"`
}

type CreateAccountUseCase struct {
	Accounts AccountRepository
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, in CreateAccountInput) (*CreateAccountOutput, error) {
	acc, err := domain.NewAccount(in.OwnerType, in.OwnerID, in.Type, in.Currency)
	if err != nil {
		return nil, err
	}
	if err := uc.Accounts.Save(ctx, acc); err != nil {
		return nil, err
	}
	return &CreateAccountOutput{
		AccountID: acc.ID.String(),
		Status:    string(acc.Status),
	}, nil
}
