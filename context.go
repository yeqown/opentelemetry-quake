package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerInstrumentationLibName = "github.com/yeqown/opentelemetry-quake"
	tracerInstrumentationVersion = "v1.2.4"
)

var (
	tracingSpanKey = struct{}{}
)

func contextWithWrapSpan(ctx context.Context, sp Span) context.Context {
	if sp == nil {
		return ctx
	}

	return context.WithValue(ctx, tracingSpanKey, sp)
}

func spanFromContext(ctx context.Context) Span {
	if ctx == nil {
		return noopSpan{}
	}

	if sp, ok := ctx.Value(tracingSpanKey).(Span); ok {
		return sp
	}

	return noopSpan{}
}

// StartSpan is alias of otel.Tracer("tracerName").Start() to avoid importing otel library in you project code.
func StartSpan(ctx context.Context, operation string, opts ...SpanStartOption) (context.Context, Span) {
	o := defaultSpanStartOption()
	for _, opt := range opts {
		opt.apply(o)
	}

	traceOptions := o.translateToTraceOptions()

	ctx2, sp := otel.
		Tracer(tracerInstrumentationLibName,
			trace.WithInstrumentationVersion(tracerInstrumentationVersion),
		).
		Start(ctx, operation, traceOptions...)

	var psc *trace.SpanContext
	if ro, ok := sp.(sdktrace.ReadOnlySpan); ok && ro.Parent().IsValid() {
		psc = new(trace.SpanContext)
		*psc = ro.Parent()
	}

	spw := wrapSpan(sp, psc)
	ctx3 := contextWithWrapSpan(ctx2, spw)

	return ctx3, spw
}

// SpanFromContext is alias of otel.SpanFromContext() to avoid importing
// otel library in you project code.
func SpanFromContext(ctx context.Context) Span {
	return spanFromContext(ctx)
}

// TraceContext span 的上下文信息，包含 span 的 traceID 和 spanID.
type TraceContext struct {
	TraceID, SpanID, ParentSpanID string
	isRemote                      bool
	sampled                       bool
	isValid                       bool
}

func (tc *TraceContext) Sampled() bool {
	if tc == nil {
		return false
	}

	return tc.sampled
}

func (tc *TraceContext) IsValid() bool {
	if tc == nil {
		return false
	}

	return tc.isValid
}

func (tc *TraceContext) IsRemote() bool {
	if tc == nil {
		return false
	}

	return tc.isRemote
}

func SpanContextFromContext(ctx context.Context) *TraceContext {
	return TraceContextFromContext(ctx)
}

// TraceContextFromContext 从 ctx 中尝试获取 TraceContext 方便日志记录
func TraceContextFromContext(ctx context.Context) *TraceContext {
	sp := SpanFromContext(ctx)
	return sp.SpanContext()
}
