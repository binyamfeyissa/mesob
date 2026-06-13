package app

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/admin/internal/domain"
)

type UpdateFlagInput struct {
	Flag       string
	Enabled    bool
	RolloutPct int
	ActorID    uuid.UUID
}

type UpdateFlagOutput struct {
	Flag       string `json:"flag"`
	Enabled    bool   `json:"enabled"`
	RolloutPct int    `json:"rollout_pct"`
}

type UpdateFlagUseCase struct {
	Flags FeatureFlagRepository
	Audit AuditRepository
}

func (uc *UpdateFlagUseCase) Execute(ctx context.Context, in UpdateFlagInput) (*UpdateFlagOutput, error) {
	if uc.Flags != nil {
		flag, err := uc.Flags.FindByName(ctx, in.Flag)
		if err != nil {
			// Flag doesn't exist yet — create it
			id, _ := uuid.NewV7()
			flag = &domain.FeatureFlag{
				ID:         id,
				Name:       in.Flag,
				Enabled:    in.Enabled,
				RolloutPct: in.RolloutPct,
				UpdatedBy:  &in.ActorID,
				UpdatedAt:  time.Now().UTC(),
			}
		} else {
			flag.Enabled = in.Enabled
			flag.RolloutPct = in.RolloutPct
			flag.UpdatedBy = &in.ActorID
			flag.UpdatedAt = time.Now().UTC()
		}
		if err := uc.Flags.Update(ctx, flag); err != nil {
			return nil, err
		}
	}

	if uc.Audit != nil {
		uc.Audit.Append(ctx, &in.ActorID, "ADMIN", "UPDATE_FLAG", in.Flag, "API", "")
	}

	return &UpdateFlagOutput{
		Flag:       in.Flag,
		Enabled:    in.Enabled,
		RolloutPct: in.RolloutPct,
	}, nil
}
