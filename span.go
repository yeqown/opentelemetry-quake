package otelquake

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerNameInternal = "github.com/yeqown/opentelemetry-quake.internal"
)

// StartSpan is alias of otel.Tracer("tracerName").Start() to avoid importing otel library in you project code.
func StartSpan(ctx context.Context, operation string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx2, sp := otel.Tracer(tracerNameInternal).Start(ctx, operation, opts...)
	return ctx2, sp
}

// SpanFromContext is alias of otel.SpanFromContext() to avoid importing otel library in you project code.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}
