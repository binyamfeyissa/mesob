package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

type Branch struct {
	ID        uuid.UUID
	Name      string
	RegionID  uuid.UUID
	OfficerID *uuid.UUID
	CreatedAt time.Time
}

type Dispute struct {
	ID              uuid.UUID
	TransactionID   uuid.UUID
	RaisedBy        uuid.UUID
	Reason          string
	Resolution      string
	ReversalTxnID   *uuid.UUID
	SecondAuthoriser *uuid.UUID
	ResolvedAt      *time.Time
	CreatedAt       time.Time
}

type Settlement struct {
	ID               uuid.UUID
	AgentID          uuid.UUID
	BranchID         uuid.UUID
	AmountMinor      int64
	TransactionID    *uuid.UUID
	AuthorisedBy     uuid.UUID
	SecondAuthoriser *uuid.UUID
	ConfirmedAt      *time.Time
	CreatedAt        time.Time
}
