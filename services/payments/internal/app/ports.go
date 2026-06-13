package app

import (
	"context"
	"github.com/gofrs/uuid"
)

type FraudClient interface {
	Screen(ctx context.Context, userID, txnType string, amountMinor int64, counterparty, channel string) (decision string, riskScore float64, rulesHit []string, err error)
}

type LedgerClient interface {
	PostTransaction(ctx context.Context, txnType, idemKey, initiatedBy, channel string, entries []LedgerEntry) (string, error)
}

type LedgerEntry struct {
	AccountID   string
	Direction   string
	AmountMinor int64
}

type IdentityClient interface {
	FindUserByMSISDN(ctx context.Context, msisdn string) (userID string, accountID string, err error)
}

type MerchantRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Merchant, error)
}

type Merchant struct {
	ID            uuid.UUID
	Name          string
	AccountID     string
	CommissionPct float64
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, aggregateID string, payload any) error
}

type PaymentRepository interface {
	SaveRef(ctx context.Context, refType, txnID, billerRef, status string) error
	UpdateRefStatus(ctx context.Context, txnID, status string) error
}
