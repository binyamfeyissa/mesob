package postgres

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/payments/internal/app"
)

type MerchantRepo struct {
	DB *pgxpool.Pool
}

func (r *MerchantRepo) FindByID(ctx context.Context, id uuid.UUID) (*app.Merchant, error) {
	var m app.Merchant
	err := r.DB.QueryRow(ctx,
		`SELECT id, name, account_id, commission_pct FROM merchants WHERE id=$1`,
		id,
	).Scan(&m.ID, &m.Name, &m.AccountID, &m.CommissionPct)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
