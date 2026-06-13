package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/identity/internal/domain"
)

type CredentialRepo struct {
	DB *pgxpool.Pool
}

func (r *CredentialRepo) FindByUserID(ctx context.Context, id uuid.UUID) (*domain.Credential, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT user_id, pin_hash, failed_count, locked_until, updated_at
		FROM credentials WHERE user_id = $1
	`, id)

	var c domain.Credential
	var lockedUntil *time.Time
	err := row.Scan(&c.UserID, &c.PINHash, &c.FailedCount, &lockedUntil, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("credential not found")
		}
		return nil, err
	}
	c.LockedUntil = lockedUntil
	return &c, nil
}

func (r *CredentialRepo) Save(ctx context.Context, c *domain.Credential) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO credentials (user_id, pin_hash, failed_count, locked_until, updated_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (user_id) DO UPDATE SET
			pin_hash     = EXCLUDED.pin_hash,
			failed_count = EXCLUDED.failed_count,
			locked_until = EXCLUDED.locked_until,
			updated_at   = EXCLUDED.updated_at
	`, c.UserID, c.PINHash, c.FailedCount, c.LockedUntil, c.UpdatedAt)
	return err
}

func (r *CredentialRepo) IncrementFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE credentials SET failed_count = failed_count + 1, updated_at = $1 WHERE user_id = $2`,
		time.Now().UTC(), id,
	)
	return err
}

func (r *CredentialRepo) Lock(ctx context.Context, id uuid.UUID, until time.Time) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE credentials SET locked_until = $1, updated_at = $2 WHERE user_id = $3`,
		until, time.Now().UTC(), id,
	)
	return err
}

func (r *CredentialRepo) ResetFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE credentials SET failed_count = 0, locked_until = NULL, updated_at = $1 WHERE user_id = $2`,
		time.Now().UTC(), id,
	)
	return err
}

func (r *CredentialRepo) UpdatePINHash(ctx context.Context, id uuid.UUID, hash []byte) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE credentials SET pin_hash = $1, updated_at = $2 WHERE user_id = $3`,
		hash, time.Now().UTC(), id,
	)
	return err
}
