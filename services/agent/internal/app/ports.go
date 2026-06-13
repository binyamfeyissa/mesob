package app

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/agent/internal/domain"
)

type AgentRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.Agent, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error)
	Save(ctx context.Context, a *domain.Agent) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AgentStatus) error
}

type LedgerClient interface {
	PostTransaction(ctx context.Context, txnType, idemKey, initiatedBy, channel string, entries []LedgerEntry) (string, error)
	GetBalance(ctx context.Context, accountID string) (int64, error)
}

type LedgerEntry struct {
	AccountID   string
	Direction   string
	AmountMinor int64
}

type IdentityClient interface {
	FindUserByMSISDN(ctx context.Context, msisdn string) (userID string, accountID string, err error)
}

type CustomerRegistrar interface {
	RegisterCustomer(ctx context.Context, msisdn, lang string) error
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}
