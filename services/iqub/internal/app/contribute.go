package app

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/iqub/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type ContributeInput struct {
	GroupID        uuid.UUID
	UserID         uuid.UUID
	CycleID        uuid.UUID
	IdempotencyKey string
}

type ContributeOutput struct {
	ContributionID string `json:"contribution_id"`
	TransactionID  string `json:"transaction_id"`
	CycleStatus    string `json:"cycle_status"`
}

type ContributeUseCase struct {
	Groups      GroupRepository
	Memberships MembershipRepository
	Ledger      LedgerClient
	Events      EventPublisher
}

func (uc *ContributeUseCase) Execute(ctx context.Context, in ContributeInput) (*ContributeOutput, error) {
	if uc.Groups == nil {
		return nil, &kiterr.DomainError{Code: "GROUP_UNAVAILABLE", Message: "group repository not configured"}
	}
	group, err := uc.Groups.FindByID(ctx, in.GroupID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "GROUP_NOT_FOUND", Message: "group not found"}
	}
	if group.Status != domain.GroupActive && group.Status != domain.GroupForming {
		return nil, &kiterr.DomainError{Code: "GROUP_INACTIVE", Message: "group is not active"}
	}

	if uc.Memberships != nil {
		_, err = uc.Memberships.FindByGroupAndUser(ctx, in.GroupID, in.UserID)
		if err != nil {
			return nil, &kiterr.DomainError{Code: "NOT_MEMBER", Message: "user is not a member of this group"}
		}
	}

	// Post ledger: user_wallet(D) → group pool_account(C)
	poolAccountID := ""
	if group.PoolAccountID != nil {
		poolAccountID = group.PoolAccountID.String()
	}

	var txnID string
	if uc.Ledger != nil && poolAccountID != "" {
		txnID, err = uc.Ledger.PostTransaction(ctx, "IQUB_CONTRIBUTION", in.IdempotencyKey, in.UserID.String(), "APP", []LedgerEntry{
			{AccountID: in.UserID.String() + "-wallet", Direction: "D", AmountMinor: group.CycleMinor},
			{AccountID: poolAccountID, Direction: "C", AmountMinor: group.CycleMinor},
		})
		if err != nil {
			return nil, err
		}
	} else {
		id, _ := uuid.NewV7()
		txnID = id.String()
	}

	// Update membership cycle state
	if uc.Memberships != nil {
		m, memberErr := uc.Memberships.FindByGroupAndUser(ctx, in.GroupID, in.UserID)
		if memberErr == nil {
			m.CycleState = domain.CycleStatePaid
			uc.Memberships.Save(ctx, m)
		}
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "IqubContributionRecorded", txnID, map[string]any{
			"group_id":    in.GroupID.String(),
			"user_id":     in.UserID.String(),
			"cycle_minor": group.CycleMinor,
		})
	}

	contribID := txnID + "-contrib"
	return &ContributeOutput{
		ContributionID: contribID,
		TransactionID:  txnID,
		CycleStatus:    string(domain.CycleStatePaid),
	}, nil
}
