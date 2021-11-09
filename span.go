package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func StartSpan(ctx context.Context, operation string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, sp := otel.Tracer("").
		Start(ctx, operation, opts...)
	return ctx, sp
}
