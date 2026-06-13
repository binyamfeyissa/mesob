package domain

type Operation struct {
	IdempotencyKey string `json:"idempotency_key"`
	Type           string `json:"type"`
	UserMSISDN     string `json:"user_msisdn"`
	AmountMinor    int64  `json:"amount_minor"`
	AuthCode       string `json:"authorisation_code,omitempty"`
	CapturedAt     string `json:"captured_at"`
}

type SyncResult struct {
	Applied  []OperationResult
	Rejected []OperationResult
}

type OperationResult struct {
	IdempotencyKey string `json:"idempotency_key"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	TransactionID  string `json:"transaction_id,omitempty"`
}
