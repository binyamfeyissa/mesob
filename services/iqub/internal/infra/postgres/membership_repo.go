package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/iqub/internal/domain"
)

type MembershipRepo struct {
	DB *pgxpool.Pool
}

func (r *MembershipRepo) Save(ctx context.Context, m *domain.Membership) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO iqub_memberships (id, group_id, user_id, payout_order, cycle_state, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			cycle_state = EXCLUDED.cycle_state
	`, m.ID, m.GroupID, m.UserID, m.PayoutOrder, string(m.CycleState), m.JoinedAt)
	return err
}

func (r *MembershipRepo) FindByGroupAndUser(ctx context.Context, groupID, userID uuid.UUID) (*domain.Membership, error) {
	m := &domain.Membership{}
	var cycleState string
	var joinedAt time.Time
	err := r.DB.QueryRow(ctx, `
		SELECT id, group_id, user_id, payout_order, cycle_state, joined_at
		FROM iqub_memberships WHERE group_id=$1 AND user_id=$2
	`, groupID, userID).Scan(&m.ID, &m.GroupID, &m.UserID, &m.PayoutOrder, &cycleState, &joinedAt)
	if err != nil {
		return nil, err
	}
	m.CycleState = domain.CycleState(cycleState)
	m.JoinedAt = joinedAt
	return m, nil
}

func (r *MembershipRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Membership, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, group_id, user_id, payout_order, cycle_state, joined_at
		FROM iqub_memberships WHERE user_id=$1 ORDER BY joined_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var memberships []domain.Membership
	for rows.Next() {
		m := domain.Membership{}
		var cycleState string
		var joinedAt time.Time
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.PayoutOrder, &cycleState, &joinedAt); err != nil {
			return nil, err
		}
		m.CycleState = domain.CycleState(cycleState)
		m.JoinedAt = joinedAt
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}

func (r *MembershipRepo) ListByGroup(ctx context.Context, groupID uuid.UUID) ([]domain.Membership, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, group_id, user_id, payout_order, cycle_state, joined_at
		FROM iqub_memberships WHERE group_id=$1 ORDER BY payout_order ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var memberships []domain.Membership
	for rows.Next() {
		m := domain.Membership{}
		var cycleState string
		var joinedAt time.Time
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.PayoutOrder, &cycleState, &joinedAt); err != nil {
			return nil, err
		}
		m.CycleState = domain.CycleState(cycleState)
		m.JoinedAt = joinedAt
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}
