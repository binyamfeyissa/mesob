package app

import (
	"context"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type CashOutInput struct {
	AgentUserID    uuid.UUID
	UserMSISDN     string
	AmountMinor    int64
	AuthCode       string
	IdempotencyKey string
}

type CashOutOutput struct {
	TransactionID   string `json:"transaction_id"`
	AgentFloatMinor int64  `json:"agent_float_minor"`
}

type CashOutUseCase struct {
	Agents   AgentRepository
	Ledger   LedgerClient
	Identity IdentityClient
	Events   EventPublisher
}

func (uc *CashOutUseCase) Execute(ctx context.Context, in CashOutInput) (*CashOutOutput, error) {
	// 1. Find agent, verify ACTIVE
	agent, err := uc.Agents.FindByUserID(ctx, in.AgentUserID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "AGENT_NOT_FOUND", Message: "agent not found"}
	}
	if !agent.IsActive() {
		return nil, &kiterr.DomainError{Code: "AGENT_SUSPENDED", Message: "agent not active"}
	}
	if in.AmountMinor <= 0 {
		return nil, &kiterr.DomainError{Code: "INVALID_AMOUNT", Message: "amount must be positive"}
	}
	// Validate auth code (must be non-empty)
	if in.AuthCode == "" {
		return nil, &kiterr.DomainError{Code: "MISSING_AUTH_CODE", Message: "authorisation code required"}
	}

	// 2. Find user by MSISDN
	if uc.Identity == nil {
		return nil, &kiterr.DomainError{Code: "IDENTITY_UNAVAILABLE", Message: "identity service not configured"}
	}
	_, userAccountID, err := uc.Identity.FindUserByMSISDN(ctx, in.UserMSISDN)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "USER_NOT_FOUND", Message: "user not found"}
	}

	// 3. Post ledger: user_wallet(D) → agent_float_account(C)
	floatAccountID := ""
	if agent.FloatAccountID != nil {
		floatAccountID = agent.FloatAccountID.String()
	}
	txnID, err := uc.Ledger.PostTransaction(ctx, "CASH_OUT", in.IdempotencyKey, in.AgentUserID.String(), "AGENT", []LedgerEntry{
		{AccountID: userAccountID, Direction: "D", AmountMinor: in.AmountMinor},
		{AccountID: floatAccountID, Direction: "C", AmountMinor: in.AmountMinor},
	})
	if err != nil {
		return nil, err
	}

	// 4. Get updated float balance
	var floatMinor int64
	if uc.Ledger != nil && floatAccountID != "" {
		floatMinor, _ = uc.Ledger.GetBalance(ctx, floatAccountID)
	}

	// 5. Publish CashOutRecorded
	if uc.Events != nil {
		uc.Events.Publish(ctx, "CashOutRecorded", txnID, map[string]any{
			"agent_id":     agent.ID.String(),
			"user_msisdn":  in.UserMSISDN,
			"amount_minor": in.AmountMinor,
		})
	}

	return &CashOutOutput{
		TransactionID:   txnID,
		AgentFloatMinor: floatMinor,
	}, nil
}
