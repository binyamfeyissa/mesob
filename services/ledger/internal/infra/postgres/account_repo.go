package postgres

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/ledger/internal/domain"
	kiterr "github.com/mesob-wallet/go-kit/errors"
)

// AccountRepo implements app.AccountRepository and app.EntryRepository.
type AccountRepo struct {
	DB *pgxpool.Pool
}

// Save inserts or updates an account row (upsert on id).
func (r *AccountRepo) Save(ctx context.Context, a *domain.Account) error {
	const q = `
		INSERT INTO accounts
			(id, owner_type, owner_id, type, currency, status, allow_negative, version, created_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE
			SET status        = EXCLUDED.status,
			    allow_negative = EXCLUDED.allow_negative,
			    version       = EXCLUDED.version`

	_, err := r.DB.Exec(ctx, q,
		a.ID,
		a.OwnerType,
		a.OwnerID,
		a.Type,
		a.Currency,
		string(a.Status),
		a.AllowNegative,
		a.Version,
		a.CreatedAt,
	)
	return err
}

// FindByID retrieves a single account by its UUID.
func (r *AccountRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	const q = `
		SELECT id, owner_type, owner_id, type, currency, status, allow_negative, version, created_at
		FROM accounts
		WHERE id = $1`

	row := r.DB.QueryRow(ctx, q, id)
	var a domain.Account
	var status string
	err := row.Scan(
		&a.ID,
		&a.OwnerType,
		&a.OwnerID,
		&a.Type,
		&a.Currency,
		&status,
		&a.AllowNegative,
		&a.Version,
		&a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, kiterr.ErrNotFound
		}
		return nil, err
	}
	a.Status = domain.AccountStatus(status)
	return &a, nil
}

// GetBalance returns the current balance_minor for an account from the
// account_balances cache table.  Returns 0 if no row exists yet (account has
// had no transactions).
func (r *AccountRepo) GetBalance(ctx context.Context, id uuid.UUID) (int64, error) {
	const q = `SELECT balance_minor FROM account_balances WHERE account_id = $1`
	var bal int64
	err := r.DB.QueryRow(ctx, q, id).Scan(&bal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return bal, nil
}

// EntryRepo implements app.EntryRepository.
type EntryRepo struct {
	DB *pgxpool.Pool
}

// ListByAccount returns ledger entries for an account with forward cursor
// pagination.  cursor is the UUID string of the last seen entry id.
// Returns (entries, nextCursor, error).
func (r *EntryRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, limit int, cursor string) ([]domain.Entry, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var (
		rows pgx.Rows
		err  error
	)

	if cursor == "" {
		const q = `
			SELECT id, account_id, direction, amount_minor
			FROM ledger_entries
			WHERE account_id = $1
			ORDER BY id ASC
			LIMIT $2`
		rows, err = r.DB.Query(ctx, q, accountID, limit+1)
	} else {
		const q = `
			SELECT id, account_id, direction, amount_minor
			FROM ledger_entries
			WHERE account_id = $1
			  AND id > $2
			ORDER BY id ASC
			LIMIT $3`
		rows, err = r.DB.Query(ctx, q, accountID, cursor, limit+1)
	}
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var entries []domain.Entry
	var lastID string
	for rows.Next() {
		var e domain.Entry
		var entryID string
		if scanErr := rows.Scan(&entryID, &e.AccountID, &e.Direction, &e.AmountMinor); scanErr != nil {
			return nil, "", scanErr
		}
		entries = append(entries, e)
		lastID = entryID
	}
	if rows.Err() != nil {
		return nil, "", rows.Err()
	}

	var nextCursor string
	if len(entries) > limit {
		entries = entries[:limit]
		nextCursor = lastID
	}
	return entries, nextCursor, nil
}
