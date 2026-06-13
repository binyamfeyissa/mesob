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

type UserRepo struct {
	DB *pgxpool.Pool
}

func (r *UserRepo) FindByMSISDN(ctx context.Context, msisdn string) (*domain.User, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, msisdn, kyc_tier, region_id, status, preferred_lang, version,
		       created_at, updated_at,
		       COALESCE(role, 'USER') AS role,
		       wallet_account_id
		FROM users
		WHERE msisdn = $1 AND deleted_at IS NULL
	`, msisdn)
	return scanUser(row)
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, msisdn, kyc_tier, region_id, status, preferred_lang, version,
		       created_at, updated_at,
		       COALESCE(role, 'USER') AS role,
		       wallet_account_id
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return scanUser(row)
}

func (r *UserRepo) Save(ctx context.Context, u *domain.User) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO users (id, msisdn, kyc_tier, region_id, status, preferred_lang, version, role, wallet_account_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			kyc_tier         = EXCLUDED.kyc_tier,
			status           = EXCLUDED.status,
			preferred_lang   = EXCLUDED.preferred_lang,
			version          = EXCLUDED.version,
			role             = EXCLUDED.role,
			wallet_account_id = COALESCE(EXCLUDED.wallet_account_id, users.wallet_account_id),
			updated_at       = EXCLUDED.updated_at
	`, u.ID, u.MSISDN, u.KYCTier, u.RegionID, string(u.Status), u.PreferredLang, u.Version, u.Role, u.WalletAccountID, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *UserRepo) UpdateTier(ctx context.Context, id uuid.UUID, tier int8, version int) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE users SET kyc_tier=$1, version=$2, updated_at=$3 WHERE id=$4`,
		tier, version, time.Now().UTC(), id,
	)
	return err
}

func (r *UserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE users SET status=$1, updated_at=$2 WHERE id=$3`,
		string(status), time.Now().UTC(), id,
	)
	return err
}

func (r *UserRepo) UpdateLanguage(ctx context.Context, id uuid.UUID, lang string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE users SET preferred_lang=$1, updated_at=$2 WHERE id=$3`,
		lang, time.Now().UTC(), id,
	)
	return err
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var status string
	err := row.Scan(
		&u.ID, &u.MSISDN, &u.KYCTier, &u.RegionID,
		&status, &u.PreferredLang, &u.Version,
		&u.CreatedAt, &u.UpdatedAt, &u.Role,
		&u.WalletAccountID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}
	u.Status = domain.UserStatus(status)
	return &u, nil
}
