package tracing

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type spanKind = trace.SpanKind

const (
	// SpanKindUnspecified is an unspecified SpanKind and is not a valid
	// SpanKind. SpanKindUnspecified should be replaced with SpanKindInternal
	// if it is received.
	SpanKindUnspecified spanKind = 0
	// SpanKindInternal is a SpanKind for a Span that represents an internal
	// operation within an application.
	SpanKindInternal spanKind = 1
	// SpanKindServer is a SpanKind for a Span that represents the operation
	// of handling a request from a client.
	SpanKindServer spanKind = 2
	// SpanKindClient is a SpanKind for a Span that represents the operation
	// of client making a request to a server.
	SpanKindClient spanKind = 3
	// SpanKindProducer is a SpanKind for a Span that represents the operation
	// of a producer sending a message to a message broker. Unlike
	// SpanKindClient and SpanKindServer, there is often no direct
	// relationship between this kind of Span and a SpanKindConsumer kind. A
	// SpanKindProducer Span will end once the message is accepted by the
	// message broker which might not overlap with the processing of that
	// message.
	SpanKindProducer spanKind = 4
	// SpanKindConsumer is a SpanKind for a Span that represents the operation
	// of a consumer receiving a message from a message broker. Like
	// SpanKindProducer Spans, there is often no direct relationship between
	// this Span and the Span that produced the message.
	SpanKindConsumer spanKind = 5
)

// Span is a specification for internal tracing client to use.
type Span interface {
	// SpanContext returns the SpanContext of the span, including the
	// trace ID, span ID, and whether the span is a root span.
	SpanContext() *TraceContext

	// RecordError records a span event capture an error. And it always
	// records stacktrace.
	RecordError(err error, opts ...SpanEventOption)

	// SetTag sets a key value pair on the span. to be used for
	// compatibility with OpenTracing.
	SetTag(key string, value string)

	// SetAttributes sets a key value pair on the span.
	SetAttributes(attributes ...attribute.KeyValue)

	// LogFields sets a key value pair on the span.
	LogFields(event string, attributes ...attribute.KeyValue)

	// SetStatus sets the status of the span. OK, Error, etc.
	SetStatus(code Code, message string)

	// Finish ends the span. same as span.End()
	Finish()

	// End ends the span. same to Finish
	End()
}

func wrapSpan(sp trace.Span, psc *trace.SpanContext) Span {
	if psc == nil {
		return spanAgent{
			root: sp,
			psc:  trace.SpanContext{},
		}
	}

	return spanAgent{root: sp, psc: *psc}
}

// spanAgent is the interface implemented by the agent that is
// responsible for propagating spans. at the same time, it contains
//
type spanAgent struct {
	root trace.Span
	psc  trace.SpanContext
}

func (s spanAgent) SpanContext() *TraceContext {
	sc := s.root.SpanContext()
	return traceSpanContextToTraceContext(sc, s.psc)
}
func (s spanAgent) RecordError(err error, opts ...SpanEventOption) {
	o := defaultSpanEventOption()
	for _, opt := range opts {
		opt.apply(o)
	}
	eventOptions := o.translateToEventOptions()

	s.root.RecordError(err, eventOptions...)
}
func (s spanAgent) SetTag(key, value string) { s.root.SetAttributes(attribute.String(key, value)) }
func (s spanAgent) SetAttributes(attributes ...attribute.KeyValue) {
	s.root.SetAttributes(attributes...)
}
func (s spanAgent) LogFields(event string, attrs ...attribute.KeyValue) {
	s.root.AddEvent(event, trace.WithAttributes(attrs...))
}
func (s spanAgent) SetStatus(code Code, message string) { s.root.SetStatus(code, message) }
func (s spanAgent) Finish()                             { s.root.End(trace.WithTimestamp(time.Now())) }
func (s spanAgent) End()                                { s.root.End(trace.WithTimestamp(time.Now())) }

func traceSpanContextToTraceContext(sc, psc trace.SpanContext) *TraceContext {
	return &TraceContext{
		TraceID:      sc.TraceID().String(),
		SpanID:       sc.SpanID().String(),
		ParentSpanID: psc.SpanID().String(),
		sampled:      sc.IsSampled(),
		isValid:      sc.IsValid(),
		isRemote:     sc.IsRemote(),
	}
}
