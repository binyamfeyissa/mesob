package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type GetBalanceOutput struct {
	AccountID    string `json:"account_id"`
	BalanceMinor int64  `json:"balance_minor"`
	Currency     string `json:"currency"`
	AsOf         string `json:"as_of"`
}

type GetBalanceUseCase struct {
	Accounts AccountRepository
}

func (uc *GetBalanceUseCase) Execute(ctx context.Context, accountID uuid.UUID) (*GetBalanceOutput, error) {
	bal, err := uc.Accounts.GetBalance(ctx, accountID)
	if err != nil {
		return nil, err
	}

	acc, err := uc.Accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &GetBalanceOutput{
		AccountID:    accountID.String(),
		BalanceMinor: bal,
		Currency:     acc.Currency,
		AsOf:         time.Now().UTC().Format(time.RFC3339),
	}, nil
}
