package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type PayPremiumInput struct {
	GroupID        uuid.UUID
	MemberID       uuid.UUID
	Period         string
	IdempotencyKey string
}

type PayPremiumOutput struct {
	PremiumID     string `json:"premium_id"`
	TransactionID string `json:"transaction_id"`
	Coverage      string `json:"coverage"`
}

type PayPremiumUseCase struct {
	Groups GroupRepository
	Ledger LedgerClient
	Events EventPublisher
}

func (uc *PayPremiumUseCase) Execute(ctx context.Context, in PayPremiumInput) (*PayPremiumOutput, error) {
	if uc.Groups == nil {
		return nil, &kiterr.DomainError{Code: "GROUP_UNAVAILABLE", Message: "group repository not configured"}
	}
	group, err := uc.Groups.FindByID(ctx, in.GroupID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "GROUP_NOT_FOUND", Message: "group not found"}
	}
	if group.Status != "ACTIVE" {
		return nil, &kiterr.DomainError{Code: "GROUP_INACTIVE", Message: "group is not active"}
	}

	poolAccountID := ""
	if group.PoolAccountID != nil {
		poolAccountID = group.PoolAccountID.String()
	}

	var txnID string
	if uc.Ledger != nil && poolAccountID != "" {
		txnID, err = uc.Ledger.PostTransaction(ctx, "IDDIR_PREMIUM", in.IdempotencyKey,
			in.MemberID.String(), "APP", []LedgerEntry{
				{AccountID: in.MemberID.String() + "-wallet", Direction: "D", AmountMinor: group.PremiumMinor},
				{AccountID: poolAccountID, Direction: "C", AmountMinor: group.PremiumMinor},
			})
		if err != nil {
			return nil, err
		}
	} else {
		id, _ := uuid.NewV7()
		txnID = id.String()
	}

	premiumID := in.MemberID.String() + "-" + in.Period + "-" + time.Now().Format("20060102")

	if uc.Events != nil {
		uc.Events.Publish(ctx, "IddirPremiumPaid", txnID, map[string]any{
			"group_id":      in.GroupID.String(),
			"member_id":     in.MemberID.String(),
			"period":        in.Period,
			"premium_minor": group.PremiumMinor,
		})
	}

	return &PayPremiumOutput{
		PremiumID:     premiumID,
		TransactionID: txnID,
		Coverage:      in.Period,
	}, nil
}
