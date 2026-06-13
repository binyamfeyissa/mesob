package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

type IddirGroup struct {
	ID             uuid.UUID
	Name           string
	PremiumMinor   int64
	Frequency      string
	BenefitMinor   int64
	Status         string
	LeaderID       uuid.UUID
	PoolAccountID  *uuid.UUID
	CreatedAt      time.Time
	DeletedAt      *time.Time
}

type Claim struct {
	ID            uuid.UUID
	GroupID       uuid.UUID
	MemberID      uuid.UUID
	Type          string
	Description   string
	EvidenceRef   string
	Status        string
	SettledMinor  int64
	TransactionID *uuid.UUID
	CreatedAt     time.Time
}
