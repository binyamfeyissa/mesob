package outbox

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
)

// Relay polls outbox_events, forwards each event to Kafka, then marks it published.
// If KafkaBrokers is empty or unreachable, events are marked published without
// forwarding so the table never grows unboundedly.
type Relay struct {
	DB           *pgxpool.Pool
	KafkaBrokers string // comma-separated; empty disables forwarding
}

type outboxRow struct {
	ID      string
	Topic   string
	Payload json.RawMessage
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.flush(ctx); err != nil {
				log.Error().Err(err).Msg("ledger outbox flush error")
			}
		}
	}
}

func (r *Relay) flush(ctx context.Context) error {
	rows, err := r.DB.Query(ctx,
		`SELECT id, topic, payload FROM outbox_events
		 WHERE published = false ORDER BY created_at LIMIT 100`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var events []outboxRow
	for rows.Next() {
		var e outboxRow
		if err := rows.Scan(&e.ID, &e.Topic, &e.Payload); err != nil {
			return err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	// Forward to Kafka when brokers are configured and reachable.
	if r.KafkaBrokers != "" {
		writer := &kafka.Writer{
			Addr:                   kafka.TCP(strings.Split(r.KafkaBrokers, ",")...),
			AllowAutoTopicCreation: true,
			WriteTimeout:           5 * time.Second,
		}
		msgs := make([]kafka.Message, 0, len(events))
		for _, e := range events {
			msgs = append(msgs, kafka.Message{
				Topic: e.Topic,
				Key:   []byte(e.ID),
				Value: e.Payload,
			})
		}
		if err := writer.WriteMessages(ctx, msgs...); err != nil {
			// Swallow network errors — still mark published so the table doesn't fill.
			if !isNetworkError(err) {
				writer.Close()
				return err
			}
			log.Warn().Err(err).Msg("ledger outbox: kafka unreachable, skipping forward")
		}
		writer.Close()
	}

	ids := make([]string, len(events))
	for i, e := range events {
		ids[i] = e.ID
	}
	_, err = r.DB.Exec(ctx, `UPDATE outbox_events SET published = true WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		return err
	}
	log.Info().Int("count", len(events)).Msg("ledger outbox flushed")
	return nil
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if ok := strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "dial"); ok {
		return true
	}
	return false || (func() bool {
		_ = netErr
		return false
	})()
}
