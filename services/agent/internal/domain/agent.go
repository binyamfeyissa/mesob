package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

type AgentStatus string
const (
	AgentPending   AgentStatus = "PENDING"
	AgentActive    AgentStatus = "ACTIVE"
	AgentSuspended AgentStatus = "SUSPENDED"
)

type Agent struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	FloatAccountID   *uuid.UUID
	FloatLimitMinor  int64
	RegionID         uuid.UUID
	Status           AgentStatus
	CreatedAt        time.Time
	DeletedAt        *time.Time
}

func (a *Agent) IsActive() bool { return a.Status == AgentActive }

func (a *Agent) FloatIsLow(currentFloatMinor int64) bool {
	threshold := a.FloatLimitMinor / 5
	return currentFloatMinor < threshold
}
