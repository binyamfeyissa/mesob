package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Publisher struct {
	DB *pgxpool.Pool
}

func (p *Publisher) Publish(ctx context.Context, eventType, aggregateID string, payload any) error {
	if p.DB == nil {
		return nil
	}
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = p.DB.Exec(ctx,
		`INSERT INTO outbox_events (id, topic, aggregate_id, payload, published, created_at)
		 VALUES ($1, $2, $3, $4, false, $5)`,
		id, eventType, aggregateID, b, time.Now().UTC(),
	)
	return err
}
