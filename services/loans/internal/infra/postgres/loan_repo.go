package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/loans/internal/domain"
)

type LoanRepo struct {
	DB *pgxpool.Pool
}

func (r *LoanRepo) Save(ctx context.Context, l *domain.Loan) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO loans (id, user_id, principal_minor, fee_minor, outstanding_minor, score_id, status, mode, due_date, mfi_facility_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			outstanding_minor = EXCLUDED.outstanding_minor
	`,
		l.ID, l.UserID, l.PrincipalMinor, l.FeeMinor, l.OutstandingMinor,
		l.ScoreID, string(l.Status), l.Mode, l.DueDate, l.MFIFacilityID, l.CreatedAt,
	)
	return err
}

func (r *LoanRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Loan, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, user_id, principal_minor, fee_minor, outstanding_minor, score_id, status, mode, due_date, mfi_facility_id, created_at
		FROM loans WHERE id = $1
	`, id)

	l := &domain.Loan{}
	var status string
	var dueDate time.Time
	err := row.Scan(
		&l.ID, &l.UserID, &l.PrincipalMinor, &l.FeeMinor, &l.OutstandingMinor,
		&l.ScoreID, &status, &l.Mode, &dueDate, &l.MFIFacilityID, &l.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &notFoundError{id: id.String()}
		}
		return nil, err
	}
	l.Status = domain.LoanStatus(status)
	l.DueDate = dueDate
	return l, nil
}

func (r *LoanRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Loan, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, user_id, principal_minor, fee_minor, outstanding_minor, score_id, status, mode, due_date, mfi_facility_id, created_at
		FROM loans WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []domain.Loan
	for rows.Next() {
		l := domain.Loan{}
		var status string
		var dueDate time.Time
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.PrincipalMinor, &l.FeeMinor, &l.OutstandingMinor,
			&l.ScoreID, &status, &l.Mode, &dueDate, &l.MFIFacilityID, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		l.Status = domain.LoanStatus(status)
		l.DueDate = dueDate
		loans = append(loans, l)
	}
	return loans, rows.Err()
}

func (r *LoanRepo) Update(ctx context.Context, l *domain.Loan) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE loans SET status = $1, outstanding_minor = $2
		WHERE id = $3
	`, string(l.Status), l.OutstandingMinor, l.ID)
	return err
}

type notFoundError struct{ id string }

func (e *notFoundError) Error() string { return "not found: " + e.id }
