package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/branch/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type SettleInput struct {
	AgentID          uuid.UUID
	OfficerID        uuid.UUID
	SecondAuthoriser uuid.UUID
	AmountMinor      int64
	IdempotencyKey   string
}

type SettleOutput struct {
	SettlementID  string `json:"settlement_id"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

type SettleUseCase struct {
	Settlements SettlementRepository
	Ledger      LedgerClient
	Events      EventPublisher
}

func (uc *SettleUseCase) Execute(ctx context.Context, in SettleInput) (*SettleOutput, error) {
	if in.OfficerID == in.SecondAuthoriser {
		return nil, kiterr.ErrSameAuthoriser
	}

	var txnID string
	var err error
	if uc.Ledger != nil {
		txnID, err = uc.Ledger.PostTransaction(ctx, "SETTLEMENT", in.IdempotencyKey,
			in.OfficerID.String(), "BRANCH", []LedgerEntry{
				{AccountID: in.AgentID.String() + "-float", Direction: "D", AmountMinor: in.AmountMinor},
				{AccountID: "branch-vault", Direction: "C", AmountMinor: in.AmountMinor},
			})
		if err != nil {
			return nil, err
		}
	} else {
		id, _ := uuid.NewV7()
		txnID = id.String()
	}

	settlementID, _ := uuid.NewV7()
	now := time.Now().UTC()
	txnUUID, _ := uuid.FromString(txnID)
	settlement := &domain.Settlement{
		ID:               settlementID,
		AgentID:          in.AgentID,
		AmountMinor:      in.AmountMinor,
		TransactionID:    &txnUUID,
		AuthorisedBy:     in.OfficerID,
		SecondAuthoriser: &in.SecondAuthoriser,
		ConfirmedAt:      &now,
		CreatedAt:        now,
	}
	if uc.Settlements != nil {
		if err := uc.Settlements.Save(ctx, settlement); err != nil {
			return nil, err
		}
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "SettlementConfirmed", settlementID.String(), map[string]any{
			"agent_id":     in.AgentID.String(),
			"amount_minor": in.AmountMinor,
			"txn_id":       txnID,
		})
	}

	return &SettleOutput{
		SettlementID:  settlementID.String(),
		Status:        "CONFIRMED",
		TransactionID: txnID,
	}, nil
}
