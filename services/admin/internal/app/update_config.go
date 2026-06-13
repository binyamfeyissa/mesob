package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/admin/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

type UpdateConfigInput struct {
	Key              string
	Value            json.RawMessage
	ActorID          uuid.UUID
	SecondAuthoriser uuid.UUID
	Reason           string
}

type UpdateConfigOutput struct {
	Key         string `json:"key"`
	Version     int    `json:"version"`
	EffectiveAt string `json:"effective_at"`
}

type UpdateConfigUseCase struct {
	Config ConfigRepository
	Audit  AuditRepository
}

func (uc *UpdateConfigUseCase) Execute(ctx context.Context, in UpdateConfigInput) (*UpdateConfigOutput, error) {
	if in.ActorID == in.SecondAuthoriser {
		return nil, kiterr.ErrSameAuthoriser
	}

	version := 1
	if uc.Config != nil {
		existing, err := uc.Config.FindByKey(ctx, in.Key)
		if err == nil && existing != nil {
			uc.Config.SaveHistory(ctx, in.Key, existing.Value, existing.Version, in.ActorID, in.Reason)
			version = existing.Version + 1
		}

		item := &domain.ConfigItem{
			Key:       in.Key,
			Value:     in.Value,
			Version:   version,
			UpdatedBy: &in.ActorID,
			UpdatedAt: time.Now().UTC(),
		}
		if err := uc.Config.Save(ctx, item); err != nil {
			return nil, err
		}
	}

	if uc.Audit != nil {
		uc.Audit.Append(ctx, &in.ActorID, "ADMIN", "UPDATE_CONFIG", in.Key, "API", "")
	}

	return &UpdateConfigOutput{
		Key:         in.Key,
		Version:     version,
		EffectiveAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
