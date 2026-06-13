package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/iddir/internal/app"
	"github.com/mesob-wallet/iddir/internal/domain"
)

type GroupRepo struct {
	DB *pgxpool.Pool
}

func (r *GroupRepo) Save(ctx context.Context, g *domain.IddirGroup) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO iddir_groups (id, name, premium_minor, frequency, benefit_minor, status, leader_id, pool_account_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			pool_account_id = EXCLUDED.pool_account_id
	`, g.ID, g.Name, g.PremiumMinor, g.Frequency, g.BenefitMinor, g.Status, g.LeaderID, g.PoolAccountID, g.CreatedAt)
	return err
}

func (r *GroupRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]app.GroupWithCoverage, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT g.id, g.name, g.premium_minor, g.benefit_minor, g.frequency, m.coverage_status
		FROM iddir_groups g
		JOIN iddir_memberships m ON g.id = m.group_id
		WHERE m.user_id = $1 AND g.deleted_at IS NULL
		ORDER BY g.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []app.GroupWithCoverage
	for rows.Next() {
		var g app.GroupWithCoverage
		var id uuid.UUID
		if err := rows.Scan(&id, &g.Name, &g.PremiumMinor, &g.BenefitMinor, &g.Frequency, &g.CoverageStatus); err != nil {
			return nil, err
		}
		g.GroupID = id.String()
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *GroupRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.IddirGroup, error) {
	g := &domain.IddirGroup{}
	var createdAt time.Time
	err := r.DB.QueryRow(ctx, `
		SELECT id, name, premium_minor, frequency, benefit_minor, status, leader_id, pool_account_id, created_at
		FROM iddir_groups WHERE id=$1 AND deleted_at IS NULL
	`, id).Scan(&g.ID, &g.Name, &g.PremiumMinor, &g.Frequency, &g.BenefitMinor, &g.Status, &g.LeaderID, &g.PoolAccountID, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &notFoundErr{"iddir_group"}
		}
		return nil, err
	}
	g.CreatedAt = createdAt
	return g, nil
}

type notFoundErr struct{ entity string }

func (e *notFoundErr) Error() string { return e.entity + " not found" }
