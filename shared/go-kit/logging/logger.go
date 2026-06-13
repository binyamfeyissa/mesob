package logging

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type contextKey struct{}

func New(serviceName string) zerolog.Logger {
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

func WithContext(ctx context.Context, l zerolog.Logger) context.Context {
	return l.WithContext(ctx)
}

func FromContext(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx)
	if l == nil || l.GetLevel() == zerolog.Disabled {
		nop := zerolog.Nop()
		return &nop
	}
	return l
}
