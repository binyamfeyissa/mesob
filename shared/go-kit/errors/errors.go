package errors

import "fmt"

type DomainError struct {
	Code      string
	Message   string
	Retryable bool
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) IsRetryable() bool {
	return e.Retryable
}

var (
	ErrInsufficientBalance  = &DomainError{Code: "INSUFFICIENT_BALANCE", Message: "insufficient balance", Retryable: false}
	ErrUnbalanced           = &DomainError{Code: "UNBALANCED", Message: "transaction entries do not balance", Retryable: false}
	ErrLimitExceeded        = &DomainError{Code: "LIMIT_EXCEEDED", Message: "kyc limit exceeded", Retryable: false}
	ErrFloatCeiling         = &DomainError{Code: "FLOAT_CEILING", Message: "agent float ceiling exceeded", Retryable: false}
	ErrIdempotencyMismatch  = &DomainError{Code: "IDEMPOTENCY_MISMATCH", Message: "idempotency key reused with different payload", Retryable: false}
	ErrNotFound             = &DomainError{Code: "NOT_FOUND", Message: "resource not found", Retryable: false}
	ErrFraudBlocked         = &DomainError{Code: "FRAUD_BLOCKED", Message: "transaction blocked by fraud screen", Retryable: false}
	ErrFraudUnavailable     = &DomainError{Code: "FRAUD_UNAVAILABLE", Message: "fraud service unavailable", Retryable: false}
	ErrScoringDeferred      = &DomainError{Code: "SCORING_DEFERRED", Message: "scoring service unavailable", Retryable: true}
	ErrInsufficientHistory  = &DomainError{Code: "INSUFFICIENT_HISTORY", Message: "insufficient transaction history for scoring", Retryable: false}
	ErrOutOfRegion          = &DomainError{Code: "OUT_OF_REGION", Message: "operation outside authorised region", Retryable: false}
	ErrSameAuthoriser       = &DomainError{Code: "SAME_AUTHORISER", Message: "four-eyes: second authoriser must differ", Retryable: false}
	ErrProviderUnavailable  = &DomainError{Code: "PROVIDER_UNAVAILABLE", Message: "external provider unavailable", Retryable: true}
	ErrAccountLocked        = &DomainError{Code: "ACCOUNT_LOCKED", Message: "account locked due to failed attempts", Retryable: false}
)

func New(code, message string, retryable bool) *DomainError {
	return &DomainError{Code: code, Message: message, Retryable: retryable}
}

func Is(err error, target *DomainError) bool {
	if de, ok := err.(*DomainError); ok {
		return de.Code == target.Code
	}
	return false
}
