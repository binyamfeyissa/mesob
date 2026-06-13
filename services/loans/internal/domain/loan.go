package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

type LoanStatus string
const (
	LoanActive    LoanStatus = "ACTIVE"
	LoanOverdue   LoanStatus = "OVERDUE"
	LoanRepaid    LoanStatus = "REPAID"
	LoanDeclined  LoanStatus = "DECLINED"
	LoanDefaulted LoanStatus = "DEFAULTED"
)

type Loan struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	PrincipalMinor   int64
	FeeMinor         int64
	OutstandingMinor int64
	ScoreID          *uuid.UUID
	Status           LoanStatus
	Mode             string
	DueDate          time.Time
	MFIFacilityID    string
	CreatedAt        time.Time
}

func (l *Loan) IsActive() bool { return l.Status == LoanActive }

func (l *Loan) ApplyRepayment(amountMinor int64) {
	l.OutstandingMinor -= amountMinor
	if l.OutstandingMinor <= 0 {
		l.OutstandingMinor = 0
		l.Status = LoanRepaid
	}
}
