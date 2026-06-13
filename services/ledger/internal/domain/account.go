package domain

import (
	"time"

	"github.com/gofrs/uuid"
)

type AccountStatus string

const (
	AccountActive AccountStatus = "ACTIVE"
	AccountFrozen AccountStatus = "FROZEN"
	AccountClosed AccountStatus = "CLOSED"
)

type Account struct {
	ID            uuid.UUID
	OwnerType     string
	OwnerID       uuid.UUID
	Type          string
	Currency      string
	Status        AccountStatus
	AllowNegative bool
	Version       int
	CreatedAt     time.Time
}

func NewAccount(ownerType, ownerID, acctType, currency string) (*Account, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	ownerUUID, err := uuid.FromString(ownerID)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:        id,
		OwnerType: ownerType,
		OwnerID:   ownerUUID,
		Type:      acctType,
		Currency:  currency,
		Status:    AccountActive,
		Version:   1,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (a *Account) IsActive() bool { return a.Status == AccountActive }
