package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/iqub/internal/domain"
)

type CreateGroupInput struct {
	Name        string
	CycleMinor  int64
	Frequency   string
	MemberLimit int
	PayoutOrder string
	LeaderID    uuid.UUID
}

type CreateGroupOutput struct {
	GroupID  string `json:"group_id"`
	Status   string `json:"status"`
	JoinCode string `json:"join_code"`
}

type CreateGroupUseCase struct {
	Groups      GroupRepository
	Memberships MembershipRepository
	Ledger      LedgerClient
	Events      EventPublisher
}

func (uc *CreateGroupUseCase) Execute(ctx context.Context, in CreateGroupInput) (*CreateGroupOutput, error) {
	g, err := domain.NewGroup(in.Name, in.CycleMinor, in.Frequency, in.MemberLimit, in.PayoutOrder, in.LeaderID)
	if err != nil {
		return nil, err
	}

	// Create pool account in ledger (non-fatal — proceed without)
	if uc.Ledger != nil {
		accountID, ledgerErr := uc.Ledger.CreateAccount(ctx, "IQUB_GROUP", g.ID.String(), "POOL", "ETB")
		if ledgerErr == nil {
			poolID, parseErr := uuid.FromString(accountID)
			if parseErr == nil {
				g.PoolAccountID = &poolID
			}
		}
	}

	if uc.Groups != nil {
		if err := uc.Groups.Save(ctx, g); err != nil {
			return nil, err
		}
	}

	// Auto-enroll the leader as the first member
	if uc.Memberships != nil {
		memberID, _ := uuid.NewV7()
		leader := &domain.Membership{
			ID:         memberID,
			GroupID:    g.ID,
			UserID:     in.LeaderID,
			CycleState: domain.CycleStatePending,
			JoinedAt:   time.Now().UTC(),
		}
		_ = uc.Memberships.Save(ctx, leader) // non-fatal
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "IqubGroupCreated", g.ID.String(), map[string]any{
			"name":      g.Name,
			"leader_id": g.LeaderID.String(),
		})
	}

	return &CreateGroupOutput{
		GroupID:  g.ID.String(),
		Status:   string(g.Status),
		JoinCode: g.JoinCode,
	}, nil
}
