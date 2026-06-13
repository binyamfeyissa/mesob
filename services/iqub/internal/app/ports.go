package app

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/iqub/internal/domain"
)

type GroupRepository interface {
	Save(ctx context.Context, g *domain.Group) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	FindByJoinCode(ctx context.Context, code string) (*domain.Group, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]GroupWithCycleInfo, error)
}

type CycleInfo struct {
	ID               string `json:"id"`
	Number           int    `json:"number"`
	Paid             int    `json:"paid"`
	Total            int    `json:"total"`
	NextPayoutMember string `json:"next_payout_member"`
	DueDate          string `json:"due_date"`
}

type GroupWithCycleInfo struct {
	GroupID     string     `json:"group_id"`
	Name        string     `json:"name"`
	CycleMinor  int64      `json:"cycle_minor"`
	Frequency   string     `json:"frequency"`
	MemberLimit int        `json:"member_limit"`
	Cycle       *CycleInfo `json:"cycle,omitempty"`
}

type MembershipRepository interface {
	Save(ctx context.Context, m *domain.Membership) error
	FindByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*domain.Membership, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error)
	ListByGroup(ctx context.Context, groupID uuid.UUID) ([]domain.Membership, error)
}

type MemberListRow struct {
	MembershipID string `json:"membership_id"`
	UserID       string `json:"user_id"`
	PayoutOrder  int    `json:"payout_order"`
	CycleState   string `json:"cycle_state"`
}

type LedgerClient interface {
	PostTransaction(ctx context.Context, txnType, idemKey, initiatedBy, channel string, entries []LedgerEntry) (string, error)
	CreateAccount(ctx context.Context, ownerType, ownerID, acctType, currency string) (string, error)
}

type LedgerEntry struct {
	AccountID   string
	Direction   string
	AmountMinor int64
}

type CycleRepository interface {
	Close(ctx context.Context, id uuid.UUID) error
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}
