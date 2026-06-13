package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/branch/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type ResolveDisputeInput struct {
	DisputeID        uuid.UUID
	OfficerID        uuid.UUID
	SecondAuthoriser uuid.UUID
	Resolution       string
	Reason           string
}

type ResolveDisputeOutput struct {
	DisputeID     string  `json:"dispute_id"`
	Status        string  `json:"status"`
	ReversalTxnID *string `json:"reversal_txn_id,omitempty"`
}

type ResolveDisputeUseCase struct {
	Disputes DisputeRepository
	Ledger   LedgerClient
	Events   EventPublisher
}

func (uc *ResolveDisputeUseCase) Execute(ctx context.Context, in ResolveDisputeInput) (*ResolveDisputeOutput, error) {
	if in.OfficerID == in.SecondAuthoriser {
		return nil, kiterr.ErrSameAuthoriser
	}

	var dispute *domain.Dispute
	var err error
	if uc.Disputes != nil {
		dispute, err = uc.Disputes.FindByID(ctx, in.DisputeID)
		if err != nil {
			return nil, &kiterr.DomainError{Code: "DISPUTE_NOT_FOUND", Message: "dispute not found"}
		}
	} else {
		dispute = &domain.Dispute{
			ID:        in.DisputeID,
			CreatedAt: time.Now().UTC(),
		}
	}

	var reversalTxnID *string
	if in.Resolution == "REFUND" && uc.Ledger != nil {
		idemKey := "reversal-" + in.DisputeID.String()
		reversalID, reverseErr := uc.Ledger.ReverseTransaction(ctx, dispute.TransactionID.String(), idemKey, in.Reason, in.OfficerID.String())
		if reverseErr == nil {
			reversalTxnID = &reversalID
			revUUID, parseErr := uuid.FromString(reversalID)
			if parseErr == nil {
				dispute.ReversalTxnID = &revUUID
			}
		}
	}

	now := time.Now().UTC()
	dispute.Resolution = in.Resolution
	dispute.ResolvedAt = &now
	dispute.SecondAuthoriser = &in.SecondAuthoriser

	if uc.Disputes != nil {
		uc.Disputes.Update(ctx, dispute)
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "DisputeResolved", in.DisputeID.String(), map[string]any{
			"resolution": in.Resolution,
			"officer_id": in.OfficerID.String(),
		})
	}

	return &ResolveDisputeOutput{
		DisputeID:     in.DisputeID.String(),
		Status:        "RESOLVED",
		ReversalTxnID: reversalTxnID,
	}, nil
}
