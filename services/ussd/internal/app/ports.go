package app

import (
	"context"

	"github.com/mesob-wallet/ussd/internal/domain"
)

type SessionStore interface {
	Get(ctx context.Context, sessionID string) (*domain.Session, error)
	Save(ctx context.Context, s *domain.Session) error
	Delete(ctx context.Context, sessionID string) error
}
