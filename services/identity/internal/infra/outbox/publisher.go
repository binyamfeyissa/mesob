package outbox

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Publisher implements app.EventPublisher by inserting events into the outbox_events
// table. The Relay picks them up and marks them published after forwarding.
type Publisher struct {
	DB *pgxpool.Pool
}

// Publish inserts an outbox event into the database as part of the caller's transaction
// context (or as a standalone insert if no transaction is active).
func (p *Publisher) Publish(ctx context.Context, eventType, aggregateID string, payload any) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = p.DB.Exec(ctx,
		`INSERT INTO outbox_events (id, topic, aggregate_id, payload, published, created_at)
		 VALUES ($1, $2, $3, $4, false, now())`,
		id, eventType, aggregateID, payloadJSON,
	)
	return err
}
