package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

type GroupStatus string
const (
	GroupForming  GroupStatus = "FORMING"
	GroupActive   GroupStatus = "ACTIVE"
	GroupComplete GroupStatus = "COMPLETE"
	GroupDisbanded GroupStatus = "DISBANDED"
)

type Group struct {
	ID             uuid.UUID
	Name           string
	CycleMinor     int64
	Frequency      string
	MemberLimit    int
	PayoutOrder    string
	Status         GroupStatus
	LeaderID       uuid.UUID
	AgentID        *uuid.UUID
	PoolAccountID  *uuid.UUID
	JoinCode       string
	CreatedAt      time.Time
	DeletedAt      *time.Time
}

func NewGroup(name string, cycleMinor int64, frequency string, memberLimit int, payoutOrder string, leaderID uuid.UUID) (*Group, error) {
	id, _ := uuid.NewV7()
	code, _ := uuid.NewV4()
	return &Group{
		ID: id, Name: name, CycleMinor: cycleMinor,
		Frequency: frequency, MemberLimit: memberLimit,
		PayoutOrder: payoutOrder, Status: GroupForming,
		LeaderID: leaderID, JoinCode: code.String()[:8],
		CreatedAt: time.Now().UTC(),
	}, nil
}
