package domain

import (
	"time"
	"github.com/gofrs/uuid"
)

type PaymentType string
const (
	PaymentP2P      PaymentType = "P2P"
	PaymentMerchant PaymentType = "MERCHANT"
	PaymentBill     PaymentType = "BILL"
)

type FraudDecision string
const (
	FraudAllow  FraudDecision = "ALLOW"
	FraudReview FraudDecision = "REVIEW"
	FraudBlock  FraudDecision = "BLOCK"
)

type Payment struct {
	ID             uuid.UUID
	Type           PaymentType
	PayerID        uuid.UUID
	PayeeID        *uuid.UUID
	AmountMinor    int64
	Note           string
	TransactionID  *uuid.UUID
	Status         string
	CreatedAt      time.Time
}
