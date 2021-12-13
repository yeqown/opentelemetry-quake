package tracing

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type SpanStartOption interface {
	apply(o *startSpanOption)
}

type startSpanOption struct {
	kind spanKind
}

func defaultSpanStartOption() *startSpanOption {
	return &startSpanOption{
		kind: SpanKindUnspecified,
	}
}

func (o *startSpanOption) translateToTraceOptions() []trace.SpanStartOption {
	traceOptions := make([]trace.SpanStartOption, 0, 2)
	traceOptions = append(traceOptions, trace.WithSpanKind(o.kind), trace.WithTimestamp(time.Now()))
	return traceOptions
}

type fnStartSpanOption func(opts *startSpanOption)

func (fn fnStartSpanOption) apply(opts *startSpanOption) { fn(opts) }
func newFnStartSpanOption(fn func(option *startSpanOption)) SpanStartOption {
	return fnStartSpanOption(fn)
}

func WithSpanKind(kind spanKind) SpanStartOption {
	return newFnStartSpanOption(func(option *startSpanOption) {
		option.kind = kind
	})
}

type SpanEventOption interface {
	apply(o *spanEventOption)
}

func defaultSpanEventOption() *spanEventOption {
	return &spanEventOption{
		withStackTrace: false,
	}
}

type spanEventOption struct {
	withStackTrace bool
}

func (o *spanEventOption) translateToEventOptions() []trace.EventOption {
	eventOptions := make([]trace.EventOption, 0, 1)
	if o.withStackTrace {
		eventOptions = append(eventOptions, trace.WithStackTrace(true))
	}
	return eventOptions
}

type fnSpanEventOption func(opts *spanEventOption)

func (fn fnSpanEventOption) apply(opts *spanEventOption) { fn(opts) }
func newFnSpanEventOption(fn func(option *spanEventOption)) SpanEventOption {
	return fnSpanEventOption(fn)
}

func WithStackTrace() SpanEventOption {
	return newFnSpanEventOption(func(option *spanEventOption) {
		option.withStackTrace = true
	})
}
