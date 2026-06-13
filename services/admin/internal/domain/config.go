package domain

import (
	"encoding/json"
	"time"
	"github.com/gofrs/uuid"
)

type ConfigItem struct {
	Key       string
	Value     json.RawMessage
	Version   int
	UpdatedBy *uuid.UUID
	UpdatedAt time.Time
}

type FeatureFlag struct {
	ID         uuid.UUID
	Name       string
	Enabled    bool
	RolloutPct int
	UpdatedBy  *uuid.UUID
	UpdatedAt  time.Time
}
