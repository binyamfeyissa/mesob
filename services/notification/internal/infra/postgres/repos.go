package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mesob-wallet/notification/internal/domain"
)

// TemplateRepo persists and retrieves notification templates.
type TemplateRepo struct {
	DB *pgxpool.Pool
}

func (r *TemplateRepo) FindByKeyLangChannel(
	ctx context.Context,
	key string,
	lang domain.Lang,
	channel domain.Channel,
) (*domain.Template, error) {
	t := &domain.Template{}
	err := r.DB.QueryRow(ctx,
		`SELECT key, lang, channel, body
		 FROM notification_templates
		 WHERE key=$1 AND lang=$2 AND channel=$3`,
		key, string(lang), string(channel),
	).Scan(&t.Key, &t.Lang, &t.Channel, &t.Body)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return t, nil
}

func (r *TemplateRepo) Upsert(ctx context.Context, t *domain.Template) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO notification_templates (key, lang, channel, body, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (key, lang, channel)
		 DO UPDATE SET body = EXCLUDED.body, updated_at = EXCLUDED.updated_at`,
		t.Key, string(t.Lang), string(t.Channel), t.Body, time.Now().UTC(),
	)
	return err
}

// DeliveryRepo persists and retrieves notification deliveries.
type DeliveryRepo struct {
	DB *pgxpool.Pool
}

func (r *DeliveryRepo) Save(ctx context.Context, d *domain.Delivery) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO notification_deliveries
		 (id, user_id, template_key, channel, status, attempts, last_error, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		d.ID, d.UserID, d.TemplateKey, string(d.Channel),
		string(d.Status), d.Attempts, d.LastError, d.CreatedAt,
	)
	return err
}

func (r *DeliveryRepo) FindByID(ctx context.Context, id string) (*domain.Delivery, error) {
	d := &domain.Delivery{}
	var ch, status string
	var createdAt time.Time
	err := r.DB.QueryRow(ctx,
		`SELECT id, user_id, template_key, channel, status, attempts, last_error, created_at
		 FROM notification_deliveries WHERE id=$1`,
		id,
	).Scan(&d.ID, &d.UserID, &d.TemplateKey, &ch, &status, &d.Attempts, &d.LastError, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("delivery not found")
		}
		return nil, err
	}
	d.Channel = domain.Channel(ch)
	d.Status = domain.DeliveryStatus(status)
	d.CreatedAt = createdAt
	return d, nil
}

func (r *DeliveryRepo) UpdateStatus(ctx context.Context, id string, status domain.DeliveryStatus, lastErr string) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE notification_deliveries SET status=$1, last_error=$2 WHERE id=$3`,
		string(status), lastErr, id,
	)
	return err
}
