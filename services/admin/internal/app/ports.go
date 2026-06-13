package app

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/mesob-wallet/admin/internal/domain"
)

type ConfigRepository interface {
	FindByKey(ctx context.Context, key string) (*domain.ConfigItem, error)
	Save(ctx context.Context, item *domain.ConfigItem) error
	SaveHistory(ctx context.Context, key string, value json.RawMessage, version int, changedBy uuid.UUID, reason string) error
}

type FeatureFlagRepository interface {
	FindByName(ctx context.Context, name string) (*domain.FeatureFlag, error)
	Update(ctx context.Context, flag *domain.FeatureFlag) error
}

type AuditRepository interface {
	Append(ctx context.Context, actorID *uuid.UUID, actorRole, action, target, channel, ip string) error
	List(ctx context.Context, filter AuditFilter) ([]AuditEntry, string, error)
}

type AuditFilter struct {
	ActorID string
	Action  string
	From    string
	To      string
	Limit   int
	Cursor  string
}

type AuditEntry struct {
	ID        string `json:"id"`
	ActorID   string `json:"actor_id"`
	ActorRole string `json:"actor_role"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	IP        string `json:"ip"`
	CreatedAt string `json:"created_at"`
}
