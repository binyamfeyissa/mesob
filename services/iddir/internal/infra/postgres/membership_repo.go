package postgres

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MembershipRepo struct {
	DB *pgxpool.Pool
}

func (r *MembershipRepo) Save(ctx context.Context, groupID, userID uuid.UUID) error {
	id, _ := uuid.NewV7()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO iddir_memberships (id, group_id, user_id, coverage_status, joined_at)
		VALUES ($1, $2, $3, 'ACTIVE', NOW())
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, id, groupID, userID)
	return err
}
