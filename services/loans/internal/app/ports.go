package app

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/loans/internal/domain"
)

type LoanRepository interface {
	Save(ctx context.Context, l *domain.Loan) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Loan, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Loan, error)
	Update(ctx context.Context, l *domain.Loan) error
}

type ScoringClient interface {
	Score(ctx context.Context, userID string, forceRecompute bool) (*domain.CreditScore, error)
}

type LedgerClient interface {
	PostTransaction(ctx context.Context, txnType, idemKey, initiatedBy, channel string, entries []LedgerEntry) (string, error)
}

type LedgerEntry struct {
	AccountID   string
	Direction   string
	AmountMinor int64
}

type MFIAdapterClient interface {
	Originate(ctx context.Context, userRef string, principalMinor int64, termDays int, scoreRef string, idemKey string) (facilityID string, err error)
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}
