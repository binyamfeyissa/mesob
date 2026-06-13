package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func TraceID(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.IsValid() {
		return sc.TraceID().String()
	}
	return ""
}
