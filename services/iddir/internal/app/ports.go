package app

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/iddir/internal/domain"
)

type GroupRepository interface {
	Save(ctx context.Context, g *domain.IddirGroup) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.IddirGroup, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]GroupWithCoverage, error)
}

type MembershipRepository interface {
	Save(ctx context.Context, groupID, userID uuid.UUID) error
}

type GroupWithCoverage struct {
	GroupID        string `json:"group_id"`
	Name           string `json:"name"`
	PremiumMinor   int64  `json:"premium_minor"`
	BenefitMinor   int64  `json:"benefit_minor"`
	Frequency      string `json:"frequency"`
	CoverageStatus string `json:"coverage_status"`
}

type ClaimRepository interface {
	Save(ctx context.Context, c *domain.Claim) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Claim, error)
	ListByGroupAndMember(ctx context.Context, groupID, memberID uuid.UUID) ([]domain.Claim, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, settledMinor int64, txnID *uuid.UUID) error
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

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}
