package app

import (
	"context"

	"github.com/gofrs/uuid"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type ApproveAgentInput struct {
	AgentID           uuid.UUID
	OfficerID         uuid.UUID
	OfficerRegionID   uuid.UUID
	AgentRegionID     uuid.UUID
	FloatCeilingMinor int64
}

type ApproveAgentOutput struct {
	AgentID      string `json:"agent_id"`
	Status       string `json:"status"`
	CeilingMinor int64  `json:"ceiling_minor"`
}

type ApproveAgentUseCase struct {
	Branches BranchRepository
	Events   EventPublisher
}

func (uc *ApproveAgentUseCase) Execute(ctx context.Context, in ApproveAgentInput) (*ApproveAgentOutput, error) {
	if in.OfficerRegionID != (uuid.UUID{}) && in.AgentRegionID != (uuid.UUID{}) {
		if in.OfficerRegionID != in.AgentRegionID {
			return nil, kiterr.ErrOutOfRegion
		}
	}

	if uc.Events != nil {
		uc.Events.Publish(ctx, "AgentApproved", in.AgentID.String(), map[string]any{
			"agent_id":            in.AgentID.String(),
			"officer_id":          in.OfficerID.String(),
			"float_ceiling_minor": in.FloatCeilingMinor,
		})
	}

	return &ApproveAgentOutput{
		AgentID:      in.AgentID.String(),
		Status:       "ACTIVE",
		CeilingMinor: in.FloatCeilingMinor,
	}, nil
}
