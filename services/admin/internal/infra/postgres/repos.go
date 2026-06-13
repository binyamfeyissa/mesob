package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/admin/internal/app"
	"github.com/mesob-wallet/admin/internal/domain"
)

// ConfigRepo implements app.ConfigRepository.
type ConfigRepo struct {
	DB *pgxpool.Pool
}

func (r *ConfigRepo) FindByKey(ctx context.Context, key string) (*domain.ConfigItem, error) {
	item := &domain.ConfigItem{}
	var updatedAt time.Time
	err := r.DB.QueryRow(ctx,
		`SELECT key, value, version, updated_by, updated_at FROM admin_config_items WHERE key=$1`,
		key,
	).Scan(&item.Key, &item.Value, &item.Version, &item.UpdatedBy, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &notFoundErr{key}
		}
		return nil, err
	}
	item.UpdatedAt = updatedAt
	return item, nil
}

func (r *ConfigRepo) Save(ctx context.Context, item *domain.ConfigItem) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO admin_config_items (key, value, version, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			version = EXCLUDED.version,
			updated_by = EXCLUDED.updated_by,
			updated_at = EXCLUDED.updated_at
	`, item.Key, item.Value, item.Version, item.UpdatedBy, time.Now().UTC())
	return err
}

func (r *ConfigRepo) SaveHistory(ctx context.Context, key string, value json.RawMessage, version int, changedBy uuid.UUID, reason string) error {
	id, _ := uuid.NewV7()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO admin_config_history (id, key, value, version, changed_by, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, key, value, version, changedBy, reason, time.Now().UTC())
	return err
}

// FlagRepo implements app.FeatureFlagRepository.
type FlagRepo struct {
	DB *pgxpool.Pool
}

func (r *FlagRepo) FindByName(ctx context.Context, name string) (*domain.FeatureFlag, error) {
	f := &domain.FeatureFlag{}
	var updatedAt time.Time
	err := r.DB.QueryRow(ctx,
		`SELECT id, name, enabled, rollout_pct, updated_by, updated_at FROM admin_feature_flags WHERE name=$1`,
		name,
	).Scan(&f.ID, &f.Name, &f.Enabled, &f.RolloutPct, &f.UpdatedBy, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &notFoundErr{name}
		}
		return nil, err
	}
	f.UpdatedAt = updatedAt
	return f, nil
}

func (r *FlagRepo) Update(ctx context.Context, flag *domain.FeatureFlag) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO admin_feature_flags (id, name, enabled, rollout_pct, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (name) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			rollout_pct = EXCLUDED.rollout_pct,
			updated_by = EXCLUDED.updated_by,
			updated_at = EXCLUDED.updated_at
	`, flag.ID, flag.Name, flag.Enabled, flag.RolloutPct, flag.UpdatedBy, time.Now().UTC())
	return err
}

// AuditRepo implements app.AuditRepository.
type AuditRepo struct {
	DB *pgxpool.Pool
}

func (r *AuditRepo) Append(ctx context.Context, actorID *uuid.UUID, actorRole, action, target, channel, ip string) error {
	id, _ := uuid.NewV7()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO admin_audit_log (id, actor_id, actor_role, action, target, channel, ip, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, actorID, actorRole, action, target, channel, ip, time.Now().UTC())
	return err
}

func (r *AuditRepo) List(ctx context.Context, filter app.AuditFilter) ([]app.AuditEntry, string, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, actor_id, actor_role, action, target, ip, created_at
		FROM admin_audit_log
		ORDER BY created_at DESC
		LIMIT $1
	`, filter.Limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var entries []app.AuditEntry
	for rows.Next() {
		e := app.AuditEntry{}
		var createdAt time.Time
		if err := rows.Scan(&e.ID, &e.ActorID, &e.ActorRole, &e.Action, &e.Target, &e.IP, &createdAt); err != nil {
			return nil, "", err
		}
		e.CreatedAt = createdAt.Format(time.RFC3339)
		entries = append(entries, e)
	}
	return entries, "", rows.Err()
}

type notFoundErr struct{ key string }

func (e *notFoundErr) Error() string { return "not found: " + e.key }
