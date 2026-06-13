package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/iddir/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type FileClaimInput struct {
	GroupID     uuid.UUID
	MemberID    uuid.UUID
	Type        string
	Description string
	EvidenceRef string
}

type FileClaimOutput struct {
	ClaimID string `json:"claim_id"`
	Status  string `json:"status"`
}

type FileClaimUseCase struct {
	Groups GroupRepository
	Claims ClaimRepository
	Events EventPublisher
}

func (uc *FileClaimUseCase) Execute(ctx context.Context, in FileClaimInput) (*FileClaimOutput, error) {
	if uc.Groups != nil {
		_, err := uc.Groups.FindByID(ctx, in.GroupID)
		if err != nil {
			return nil, &kiterr.DomainError{Code: "GROUP_NOT_FOUND", Message: "group not found"}
		}
	}

	id, _ := uuid.NewV7()
	claim := &domain.Claim{
		ID:          id,
		GroupID:     in.GroupID,
		MemberID:    in.MemberID,
		Type:        in.Type,
		Description: in.Description,
		EvidenceRef: in.EvidenceRef,
		Status:      "UNDER_REVIEW",
		CreatedAt:   time.Now().UTC(),
	}

	if uc.Claims != nil {
		if err := uc.Claims.Save(ctx, claim); err != nil {
			return nil, err
		}
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "IddirClaimFiled", id.String(), map[string]any{
			"group_id":  in.GroupID.String(),
			"member_id": in.MemberID.String(),
			"type":      in.Type,
		})
	}

	return &FileClaimOutput{
		ClaimID: id.String(),
		Status:  "UNDER_REVIEW",
	}, nil
}
