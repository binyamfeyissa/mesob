package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/branch/internal/domain"
)

// SettlementRepo implements app.SettlementRepository.
type SettlementRepo struct {
	DB *pgxpool.Pool
}

func (r *SettlementRepo) Save(ctx context.Context, s *domain.Settlement) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO branch_settlements (id, agent_id, branch_id, amount_minor, transaction_id, authorised_by, second_authoriser, confirmed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`, s.ID, s.AgentID, s.BranchID, s.AmountMinor, s.TransactionID,
		s.AuthorisedBy, s.SecondAuthoriser, s.ConfirmedAt, s.CreatedAt)
	return err
}

// DisputeRepo implements app.DisputeRepository.
type DisputeRepo struct {
	DB *pgxpool.Pool
}

func (r *DisputeRepo) Save(ctx context.Context, d *domain.Dispute) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO branch_disputes (id, transaction_id, raised_by, reason, resolution, reversal_txn_id, second_authoriser, resolved_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING
	`, d.ID, d.TransactionID, d.RaisedBy, d.Reason, d.Resolution,
		d.ReversalTxnID, d.SecondAuthoriser, d.ResolvedAt, d.CreatedAt)
	return err
}

func (r *DisputeRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Dispute, error) {
	d := &domain.Dispute{}
	var createdAt time.Time
	err := r.DB.QueryRow(ctx, `
		SELECT id, transaction_id, raised_by, reason, resolution, reversal_txn_id, second_authoriser, resolved_at, created_at
		FROM branch_disputes WHERE id=$1
	`, id).Scan(&d.ID, &d.TransactionID, &d.RaisedBy, &d.Reason, &d.Resolution,
		&d.ReversalTxnID, &d.SecondAuthoriser, &d.ResolvedAt, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &notFoundErr{"dispute"}
		}
		return nil, err
	}
	d.CreatedAt = createdAt
	return d, nil
}

func (r *DisputeRepo) Update(ctx context.Context, d *domain.Dispute) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE branch_disputes SET resolution=$1, reversal_txn_id=$2, resolved_at=$3 WHERE id=$4
	`, d.Resolution, d.ReversalTxnID, d.ResolvedAt, d.ID)
	return err
}

type notFoundErr struct{ entity string }

func (e *notFoundErr) Error() string { return e.entity + " not found" }
