package domain

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type Direction string

const (
	Debit  Direction = "D"
	Credit Direction = "C"
)

type Entry struct {
	AccountID   uuid.UUID
	Direction   Direction
	AmountMinor int64
}

type Transaction struct {
	ID             uuid.UUID
	Type           string
	Status         string
	IdempotencyKey string
	InitiatedBy    uuid.UUID
	Channel        string
	RevertsTxnID   *uuid.UUID
	Entries        []Entry
	CreatedAt      time.Time
}

func NewBalancedTransaction(txnType, idemKey, initiatedBy, channel string, entries []Entry) (*Transaction, error) {
	if len(entries) < 2 {
		return nil, fmt.Errorf("%w: need at least 2 entries", kiterr.ErrUnbalanced)
	}

	var debitSum, creditSum int64
	for _, e := range entries {
		if e.AmountMinor <= 0 {
			return nil, fmt.Errorf("amount_minor must be positive")
		}
		if e.Direction == Debit {
			debitSum += e.AmountMinor
		} else {
			creditSum += e.AmountMinor
		}
	}
	if debitSum != creditSum {
		return nil, fmt.Errorf("%w: debits=%d credits=%d", kiterr.ErrUnbalanced, debitSum, creditSum)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	initiatedByUUID, _ := uuid.FromString(initiatedBy)

	return &Transaction{
		ID:             id,
		Type:           txnType,
		Status:         "POSTED",
		IdempotencyKey: idemKey,
		InitiatedBy:    initiatedByUUID,
		Channel:        channel,
		Entries:        entries,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
