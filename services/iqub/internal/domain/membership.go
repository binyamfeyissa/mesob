package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

type CycleState string
const (
	CycleStatePending  CycleState = "PENDING"
	CycleStatePaid     CycleState = "PAID"
	CycleStateMissed   CycleState = "MISSED"
	CycleStateReceived CycleState = "RECEIVED"
)

type Membership struct {
	ID          uuid.UUID
	GroupID     uuid.UUID
	UserID      uuid.UUID
	PayoutOrder int
	CycleState  CycleState
	JoinedAt    time.Time
}
