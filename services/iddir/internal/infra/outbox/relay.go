package outbox

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
)

type Relay struct {
	DB           *pgxpool.Pool
	KafkaBrokers string
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
				log.Error().Err(err).Msg("iddir outbox flush error")
			}
		}
	}
}

func (r *Relay) flush(ctx context.Context) error {
	if r.DB == nil {
		return nil
	}
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
			if !strings.Contains(err.Error(), "connection refused") &&
				!strings.Contains(err.Error(), "no such host") {
				writer.Close()
				return err
			}
			log.Warn().Err(err).Msg("iddir outbox: kafka unreachable, skipping forward")
		}
		writer.Close()
	}

	ids := make([]string, len(events))
	for i, e := range events {
		ids[i] = e.ID
	}
	tag, err := r.DB.Exec(ctx, `UPDATE outbox_events SET published=true WHERE id = ANY($1::uuid[])`, ids)
	if err != nil {
		return err
	}
	log.Info().Int64("count", tag.RowsAffected()).Msg("iddir outbox flushed")
	return nil
}
