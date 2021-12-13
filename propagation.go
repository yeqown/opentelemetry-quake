package tracing

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

// TraceContextCarrier is the type of the carrier to be used for
// propagating the trace context across processes.
type TraceContextCarrier interface {
	Get(key string) string
	Set(key, value string)
}

// TraceContextPropagator is the type of the propagator to be used for
// propagating the trace context across processes.
type TraceContextPropagator interface {
	Inject(ctx context.Context, carrier TraceContextCarrier)
	Extract(ctx context.Context, carrier TraceContextCarrier) context.Context
}

var _ propagation.TextMapCarrier = (*carrierAdapter)(nil)

type carrierAdapter struct {
	carrier TraceContextCarrier
}

func (c carrierAdapter) Get(key string) string        { return c.carrier.Get(key) }
func (c carrierAdapter) Set(key string, value string) { c.carrier.Set(key, value) }
func (c carrierAdapter) Keys() []string               { return nil } // noop

var _ TraceContextCarrier = (*mapCarrier)(nil)

type mapCarrier map[string]string

// NewMapCarrier returns a new TraceContextCarrier that wraps a map[string]string.
func NewMapCarrier() TraceContextCarrier {
	return mapCarrier(make(map[string]string, 2))
}

func (m mapCarrier) Get(key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

func (m mapCarrier) Set(key, value string) {
	if m == nil {
		return
	}

	m[key] = value
}

var (
	_          TraceContextPropagator = (*defaultTraceContextPropagator)(nil)
	propagator TraceContextPropagator = defaultTraceContextPropagator{}
)

// defaultTraceContextPropagator use propagation.TraceContext directly. At the same time,
// it hides the details of how to load traceparent and tracestate from the carrier.
// To extend the default behavior, just implement the TraceContextCarrier interface, such as:
// http.Header, grpc.Metadata, etc.
type defaultTraceContextPropagator struct{}

func (d defaultTraceContextPropagator) Inject(ctx context.Context, carrier TraceContextCarrier) {
	propagation.TraceContext{}.Inject(ctx, carrierAdapter{carrier})
}

func (d defaultTraceContextPropagator) Extract(ctx context.Context, carrier TraceContextCarrier) context.Context {
	return propagation.TraceContext{}.Extract(ctx, carrierAdapter{carrier})
}

// GetPropagator returns the trace context propagator.
func GetPropagator() TraceContextPropagator {
	return propagator
}

// SetPropagator sets the trace context propagator. if it never called,
// the default propagator will be used.
func SetPropagator(p TraceContextPropagator) {
	propagator = p
}
