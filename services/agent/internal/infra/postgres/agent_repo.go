package postgres

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/agent/internal/domain"
)

type AgentRepo struct {
	DB *pgxpool.Pool
}

func (r *AgentRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.Agent, error) {
	var a domain.Agent
	err := r.DB.QueryRow(ctx,
		`SELECT id, user_id, float_account_id, float_limit_minor, region_id, status, created_at, deleted_at
		 FROM agents WHERE user_id=$1 AND deleted_at IS NULL`,
		userID,
	).Scan(
		&a.ID, &a.UserID, &a.FloatAccountID, &a.FloatLimitMinor,
		&a.RegionID, &a.Status, &a.CreatedAt, &a.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AgentRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	var a domain.Agent
	err := r.DB.QueryRow(ctx,
		`SELECT id, user_id, float_account_id, float_limit_minor, region_id, status, created_at, deleted_at
		 FROM agents WHERE id=$1 AND deleted_at IS NULL`,
		id,
	).Scan(
		&a.ID, &a.UserID, &a.FloatAccountID, &a.FloatLimitMinor,
		&a.RegionID, &a.Status, &a.CreatedAt, &a.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AgentRepo) Save(ctx context.Context, a *domain.Agent) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO agents (id, user_id, float_account_id, float_limit_minor, region_id, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO UPDATE SET
		   status = EXCLUDED.status,
		   float_account_id = EXCLUDED.float_account_id,
		   float_limit_minor = EXCLUDED.float_limit_minor`,
		a.ID, a.UserID, a.FloatAccountID, a.FloatLimitMinor,
		a.RegionID, a.Status, a.CreatedAt,
	)
	return err
}

func (r *AgentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AgentStatus) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE agents SET status=$1 WHERE id=$2`,
		status, id,
	)
	return err
}
