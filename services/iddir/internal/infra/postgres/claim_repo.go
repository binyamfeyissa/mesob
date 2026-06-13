package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/iddir/internal/domain"
)

type ClaimRepo struct {
	DB *pgxpool.Pool
}

func (r *ClaimRepo) Save(ctx context.Context, c *domain.Claim) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO iddir_claims (id, group_id, member_id, type, description, evidence_ref, status, settled_minor, transaction_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			settled_minor = EXCLUDED.settled_minor,
			transaction_id = EXCLUDED.transaction_id
	`, c.ID, c.GroupID, c.MemberID, c.Type, c.Description, c.EvidenceRef,
		c.Status, c.SettledMinor, c.TransactionID, c.CreatedAt)
	return err
}

func (r *ClaimRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Claim, error) {
	c := &domain.Claim{}
	var createdAt time.Time
	err := r.DB.QueryRow(ctx, `
		SELECT id, group_id, member_id, type, description, evidence_ref, status, settled_minor, transaction_id, created_at
		FROM iddir_claims WHERE id=$1
	`, id).Scan(&c.ID, &c.GroupID, &c.MemberID, &c.Type, &c.Description, &c.EvidenceRef,
		&c.Status, &c.SettledMinor, &c.TransactionID, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &notFoundErr{"claim"}
		}
		return nil, err
	}
	c.CreatedAt = createdAt
	return c, nil
}

func (r *ClaimRepo) ListByGroupAndMember(ctx context.Context, groupID, memberID uuid.UUID) ([]domain.Claim, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, group_id, member_id, type, description, evidence_ref, status, settled_minor, transaction_id, created_at
		FROM iddir_claims WHERE group_id = $1 AND member_id = $2 ORDER BY created_at DESC
	`, groupID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var claims []domain.Claim
	for rows.Next() {
		c := domain.Claim{}
		var createdAt time.Time
		if err := rows.Scan(&c.ID, &c.GroupID, &c.MemberID, &c.Type, &c.Description,
			&c.EvidenceRef, &c.Status, &c.SettledMinor, &c.TransactionID, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt = createdAt
		claims = append(claims, c)
	}
	return claims, rows.Err()
}

func (r *ClaimRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, settledMinor int64, txnID *uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE iddir_claims SET status=$1, settled_minor=$2, transaction_id=$3 WHERE id=$4
	`, status, settledMinor, txnID, id)
	return err
}
