package tracing

import "go.opentelemetry.io/otel/attribute"

var _ Span = (*noopSpan)(nil)

type noopSpan struct{}

func (n noopSpan) SpanContext() *TraceContext                               { return &TraceContext{} }
func (n noopSpan) RecordError(err error, opts ...SpanEventOption)           {}
func (n noopSpan) SetTag(key string, value string)                          {}
func (n noopSpan) SetAttributes(attributes ...attribute.KeyValue)           {}
func (n noopSpan) LogFields(event string, attributes ...attribute.KeyValue) {}
func (n noopSpan) SetStatus(code Code, message string)                      {}
func (n noopSpan) Finish()                                                  {}
func (n noopSpan) End()                                                     {}
