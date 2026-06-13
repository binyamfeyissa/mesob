package app

import (
	"context"
	"io"
)

type NIDProvider interface {
	Verify(ctx context.Context, fan string, name string, dob string) (verified bool, matchScore float64, err error)
}

type MFIProvider interface {
	Originate(ctx context.Context, userRef string, principalMinor int64, termDays int, scoreRef string) (facilityID string, err error)
}

type WebhookProcessor interface {
	Process(ctx context.Context, partner string, body io.Reader, signature string, timestamp string) error
}
