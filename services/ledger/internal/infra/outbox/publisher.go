package outbox

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Publisher implements app.EventPublisher by writing events to the outbox
// table.  The Relay picks them up and forwards them to Kafka.
type Publisher struct {
	DB *pgxpool.Pool
}

// Publish inserts an outbox event row.  payload is serialised to JSON.
func (p *Publisher) Publish(ctx context.Context, eventType, aggregateID string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	const q = `
		INSERT INTO outbox_events
			(id, topic, aggregate_id, payload, published, created_at)
		VALUES
			($1, $2, $3, $4, false, now())`

	_, err = p.DB.Exec(ctx, q, id, eventType, aggregateID, raw)
	return err
}
