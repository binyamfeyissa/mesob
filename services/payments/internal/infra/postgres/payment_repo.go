package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/payments/internal/app"
)

type PaymentRepo struct {
	DB *pgxpool.Pool
}

// SaveRef implements app.PaymentRepository.
func (r *PaymentRepo) SaveRef(ctx context.Context, refType, txnID, billerRef, status string) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(ctx,
		`INSERT INTO payment_refs (id, ref_type, txn_id, biller_ref, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, refType, txnID, billerRef, status, time.Now().UTC(),
	)
	return err
}

// UpdateRefStatus implements app.PaymentRepository.
func (r *PaymentRepo) UpdateRefStatus(ctx context.Context, txnID, status string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE payment_refs SET status=$1 WHERE txn_id=$2`,
		status, txnID,
	)
	return err
}

// FindByID implements app.BillerRepository.
func (r *PaymentRepo) FindByID(ctx context.Context, id string) (*app.Biller, error) {
	var b app.Biller
	err := r.DB.QueryRow(ctx,
		`SELECT id, name, adapter_key, status FROM billers WHERE id=$1 AND status='ACTIVE'`,
		id,
	).Scan(&b.ID, &b.Name, &b.AdapterKey, &b.Status)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
