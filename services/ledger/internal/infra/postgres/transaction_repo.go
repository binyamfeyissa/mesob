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

// TransactionRepo implements app.TransactionRepository.
type TransactionRepo struct {
	DB *pgxpool.Pool
}

// Save atomically persists a transaction, all its ledger entries, updates the
// account_balances cache and records the idempotency key — all within a single
// database transaction.
func (r *TransactionRepo) Save(ctx context.Context, tx *domain.Transaction) error {
	return pgx.BeginTxFunc(ctx, r.DB, pgx.TxOptions{}, func(dbtx pgx.Tx) error {
		// 1. Insert transaction row.
		const insertTxn = `
			INSERT INTO transactions
				(id, type, status, idempotency_key, initiated_by, channel, reverts_txn_id, created_at)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err := dbtx.Exec(ctx, insertTxn,
			tx.ID,
			tx.Type,
			tx.Status,
			tx.IdempotencyKey,
			tx.InitiatedBy,
			tx.Channel,
			tx.RevertsTxnID,
			tx.CreatedAt,
		)
		if err != nil {
			return err
		}

		// 2. Insert entries and update balances.
		const insertEntry = `
			INSERT INTO ledger_entries
				(id, transaction_id, account_id, direction, amount_minor, created_at)
			VALUES
				($1, $2, $3, $4, $5, $6)`

		const upsertBalance = `
			INSERT INTO account_balances (account_id, balance_minor)
			VALUES ($1, $2)
			ON CONFLICT (account_id) DO UPDATE
				SET balance_minor = account_balances.balance_minor +
					CASE WHEN $3 = 'C' THEN $4::bigint ELSE -$4::bigint END`

		for _, e := range tx.Entries {
			entryID, genErr := uuid.NewV7()
			if genErr != nil {
				return genErr
			}

			if _, execErr := dbtx.Exec(ctx, insertEntry,
				entryID,
				tx.ID,
				e.AccountID,
				string(e.Direction),
				e.AmountMinor,
				tx.CreatedAt,
			); execErr != nil {
				return execErr
			}

			if _, execErr := dbtx.Exec(ctx, upsertBalance,
				e.AccountID,
				// Initial balance on insert: credit positive, debit negative.
				balanceDelta(e.Direction, e.AmountMinor),
				string(e.Direction),
				e.AmountMinor,
			); execErr != nil {
				return execErr
			}
		}

		// 3. Store idempotency key.
		const insertIdem = `
			INSERT INTO idempotency_keys (key, created_at)
			VALUES ($1, $2)
			ON CONFLICT (key) DO NOTHING`
		_, err = dbtx.Exec(ctx, insertIdem, tx.IdempotencyKey, tx.CreatedAt)
		return err
	})
}

// balanceDelta returns the signed amount to use as the initial insert value in
// account_balances.  Credits add; debits subtract.
func balanceDelta(dir domain.Direction, amount int64) int64 {
	if dir == domain.Credit {
		return amount
	}
	return -amount
}

// FindByIdempotencyKey retrieves a transaction (with entries) by its
// idempotency key.  Returns kiterr.ErrNotFound if no match.
func (r *TransactionRepo) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	const q = `
		SELECT id, type, status, idempotency_key, initiated_by, channel, reverts_txn_id, created_at
		FROM transactions
		WHERE idempotency_key = $1`

	return r.scanTransaction(ctx, r.DB.QueryRow(ctx, q, key))
}

// FindByID retrieves a transaction (with entries) by its UUID.
// Returns kiterr.ErrNotFound if no match.
func (r *TransactionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	const q = `
		SELECT id, type, status, idempotency_key, initiated_by, channel, reverts_txn_id, created_at
		FROM transactions
		WHERE id = $1`

	return r.scanTransaction(ctx, r.DB.QueryRow(ctx, q, id))
}

// scanTransaction scans a transaction header row and loads its entries.
func (r *TransactionRepo) scanTransaction(ctx context.Context, row pgx.Row) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := row.Scan(
		&tx.ID,
		&tx.Type,
		&tx.Status,
		&tx.IdempotencyKey,
		&tx.InitiatedBy,
		&tx.Channel,
		&tx.RevertsTxnID,
		&tx.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, kiterr.ErrNotFound
		}
		return nil, err
	}

	entries, err := r.loadEntries(ctx, tx.ID)
	if err != nil {
		return nil, err
	}
	tx.Entries = entries
	return &tx, nil
}

// loadEntries fetches all ledger entries for a transaction.
func (r *TransactionRepo) loadEntries(ctx context.Context, txnID uuid.UUID) ([]domain.Entry, error) {
	const q = `
		SELECT account_id, direction, amount_minor
		FROM ledger_entries
		WHERE transaction_id = $1
		ORDER BY id ASC`

	rows, err := r.DB.Query(ctx, q, txnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.Entry
	for rows.Next() {
		var e domain.Entry
		var dir string
		if scanErr := rows.Scan(&e.AccountID, &dir, &e.AmountMinor); scanErr != nil {
			return nil, scanErr
		}
		e.Direction = domain.Direction(dir)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
