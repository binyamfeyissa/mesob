package postgres

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CycleRepo struct {
	DB *pgxpool.Pool
}

func (r *CycleRepo) Close(ctx context.Context, cycleID uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE iqub_cycles SET status = 'CLOSED', closed_at = NOW()
		WHERE id = $1 AND status = 'OPEN'
	`, cycleID)
	return err
}
