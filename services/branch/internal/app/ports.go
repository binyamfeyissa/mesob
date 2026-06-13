package app

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/branch/internal/domain"
)

type BranchRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	FindByOfficerID(ctx context.Context, officerID uuid.UUID) (*domain.Branch, error)
}

type DisputeRepository interface {
	Save(ctx context.Context, d *domain.Dispute) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Dispute, error)
	Update(ctx context.Context, d *domain.Dispute) error
}

type SettlementRepository interface {
	Save(ctx context.Context, s *domain.Settlement) error
}

type LedgerClient interface {
	PostTransaction(ctx context.Context, txnType, idemKey, initiatedBy, channel string, entries []LedgerEntry) (string, error)
	ReverseTransaction(ctx context.Context, txnID, idemKey, reason, authorisedBy string) (string, error)
}

type LedgerEntry struct {
	AccountID   string
	Direction   string
	AmountMinor int64
}

type IdentityClient interface {
	GetUserKYCTier(ctx context.Context, userID string) (int8, error)
	UpgradeTier(ctx context.Context, userID string, tier int8) error
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}
