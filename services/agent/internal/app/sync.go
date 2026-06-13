package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
	"github.com/mesob-wallet/agent/internal/domain"
)

type SyncInput struct {
	AgentUserID uuid.UUID
	Operations  []domain.Operation
	SinceCursor string
}

type SyncOutput struct {
	Applied  []domain.OperationResult `json:"applied"`
	Rejected []domain.OperationResult `json:"rejected"`
	State    struct {
		FloatMinor int64  `json:"float_minor"`
		Cursor     string `json:"cursor"`
	} `json:"state"`
}

type SyncUseCase struct {
	Agents  AgentRepository
	CashIn  *CashInUseCase
	CashOut *CashOutUseCase
}

func (uc *SyncUseCase) Execute(ctx context.Context, in SyncInput) (*SyncOutput, error) {
	// Find agent
	agent, err := uc.Agents.FindByUserID(ctx, in.AgentUserID)
	if err != nil {
		return nil, &kiterr.DomainError{Code: "AGENT_NOT_FOUND", Message: "agent not found"}
	}
	_ = agent

	out := &SyncOutput{}
	out.Applied = []domain.OperationResult{}
	out.Rejected = []domain.OperationResult{}

	// Process each operation individually (not all-or-nothing)
	for _, op := range in.Operations {
		var opErr error
		var txnID string

		switch op.Type {
		case "CASH_IN":
			result, err := uc.CashIn.Execute(ctx, CashInInput{
				AgentUserID:    in.AgentUserID,
				UserMSISDN:     op.UserMSISDN,
				AmountMinor:    op.AmountMinor,
				CapturedAt:     op.CapturedAt,
				IdempotencyKey: op.IdempotencyKey,
			})
			if err != nil {
				opErr = err
			} else {
				txnID = result.TransactionID
			}
		case "CASH_OUT":
			result, err := uc.CashOut.Execute(ctx, CashOutInput{
				AgentUserID:    in.AgentUserID,
				UserMSISDN:     op.UserMSISDN,
				AmountMinor:    op.AmountMinor,
				AuthCode:       op.AuthCode,
				IdempotencyKey: op.IdempotencyKey,
			})
			if err != nil {
				opErr = err
			} else {
				txnID = result.TransactionID
			}
		default:
			opErr = fmt.Errorf("unknown operation type: %s", op.Type)
		}

		if opErr != nil {
			out.Rejected = append(out.Rejected, domain.OperationResult{
				IdempotencyKey: op.IdempotencyKey,
				Status:         "REJECTED",
				Error:          opErr.Error(),
			})
		} else {
			out.Applied = append(out.Applied, domain.OperationResult{
				IdempotencyKey: op.IdempotencyKey,
				Status:         "APPLIED",
				TransactionID:  txnID,
			})
		}
	}

	// Get current float state
	var floatMinor int64
	agentForFloat, _ := uc.Agents.FindByUserID(ctx, in.AgentUserID)
	if agentForFloat != nil && agentForFloat.FloatAccountID != nil && uc.CashIn.Ledger != nil {
		floatMinor, _ = uc.CashIn.Ledger.GetBalance(ctx, agentForFloat.FloatAccountID.String())
	}
	out.State.FloatMinor = floatMinor
	out.State.Cursor = time.Now().UTC().Format(time.RFC3339)

	return out, nil
}
