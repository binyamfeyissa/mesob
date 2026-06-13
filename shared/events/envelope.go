package events

import (
	"encoding/json"
	"time"
)

type Envelope struct {
	EventID     string          `json:"event_id"`
	Type        string          `json:"type"`
	Version     int             `json:"version"`
	OccurredAt  time.Time       `json:"occurred_at"`
	AggregateID string          `json:"aggregate_id"`
	Payload     json.RawMessage `json:"payload"`
}

func New(eventType, aggregateID string, version int, payload any) (Envelope, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		Type:        eventType,
		Version:     version,
		OccurredAt:  time.Now().UTC(),
		AggregateID: aggregateID,
		Payload:     json.RawMessage(b),
	}, nil
}
