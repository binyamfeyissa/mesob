package app

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/ledger/internal/domain"
)

type AccountRepository interface {
	Save(ctx context.Context, a *domain.Account) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetBalance(ctx context.Context, id uuid.UUID) (int64, error)
}

type TransactionRepository interface {
	Save(ctx context.Context, tx *domain.Transaction) error
	FindByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
}

type EntryRepository interface {
	ListByAccount(ctx context.Context, accountID uuid.UUID, limit int, cursor string) ([]domain.Entry, string, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}

// UnitOfWork atomically saves transaction+entries and publishes event via outbox
type UnitOfWork interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}
