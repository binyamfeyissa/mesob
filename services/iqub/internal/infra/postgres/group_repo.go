package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/iqub/internal/app"
	"github.com/mesob-wallet/iqub/internal/domain"
)

type GroupRepo struct {
	DB *pgxpool.Pool
}

func (r *GroupRepo) Save(ctx context.Context, g *domain.Group) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO iqub_groups (id, name, cycle_minor, frequency, member_limit, payout_order, status, leader_id, pool_account_id, join_code, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			pool_account_id = EXCLUDED.pool_account_id
	`, g.ID, g.Name, g.CycleMinor, g.Frequency, g.MemberLimit, g.PayoutOrder,
		string(g.Status), g.LeaderID, g.PoolAccountID, g.JoinCode, g.CreatedAt)
	return err
}

func (r *GroupRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]app.GroupWithCycleInfo, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT
			g.id, g.name, g.cycle_minor, g.frequency, g.member_limit,
			c.id,
			COALESCE(c.number, 0),
			c.due_date,
			COALESCE(c.recipient_id::text, ''),
			(SELECT COUNT(*) FROM iqub_memberships WHERE group_id = g.id AND cycle_state = 'PAID'),
			(SELECT COUNT(*) FROM iqub_memberships WHERE group_id = g.id)
		FROM iqub_groups g
		JOIN iqub_memberships m ON g.id = m.group_id AND m.user_id = $1
		LEFT JOIN iqub_cycles c ON c.group_id = g.id AND c.status = 'OPEN'
		WHERE g.deleted_at IS NULL
		ORDER BY g.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []app.GroupWithCycleInfo
	for rows.Next() {
		var g app.GroupWithCycleInfo
		var gid uuid.UUID
		var cycleID *uuid.UUID
		var cycleNumber int
		var dueDate *time.Time
		var nextPayout string
		var paid, total int64
		if err := rows.Scan(&gid, &g.Name, &g.CycleMinor, &g.Frequency, &g.MemberLimit,
			&cycleID, &cycleNumber, &dueDate, &nextPayout, &paid, &total); err != nil {
			return nil, err
		}
		g.GroupID = gid.String()
		if cycleID != nil {
			dueDateStr := ""
			if dueDate != nil {
				dueDateStr = dueDate.Format("2006-01-02")
			}
			g.Cycle = &app.CycleInfo{
				ID:               cycleID.String(),
				Number:           cycleNumber,
				Paid:             int(paid),
				Total:            int(total),
				NextPayoutMember: nextPayout,
				DueDate:          dueDateStr,
			}
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *GroupRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, name, cycle_minor, frequency, member_limit, payout_order, status, leader_id, pool_account_id, join_code, created_at
		FROM iqub_groups WHERE id=$1 AND deleted_at IS NULL
	`, id)
	return scanGroup(row)
}

func (r *GroupRepo) FindByJoinCode(ctx context.Context, code string) (*domain.Group, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, name, cycle_minor, frequency, member_limit, payout_order, status, leader_id, pool_account_id, join_code, created_at
		FROM iqub_groups WHERE join_code=$1 AND deleted_at IS NULL
	`, code)
	return scanGroup(row)
}

func scanGroup(row pgx.Row) (*domain.Group, error) {
	g := &domain.Group{}
	var status string
	var createdAt time.Time
	err := row.Scan(&g.ID, &g.Name, &g.CycleMinor, &g.Frequency, &g.MemberLimit,
		&g.PayoutOrder, &status, &g.LeaderID, &g.PoolAccountID, &g.JoinCode, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &notFoundErr{"group"}
		}
		return nil, err
	}
	g.Status = domain.GroupStatus(status)
	g.CreatedAt = createdAt
	return g, nil
}

type notFoundErr struct{ entity string }

func (e *notFoundErr) Error() string { return e.entity + " not found" }
