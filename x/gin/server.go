package otelgin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/yeqown/opentelemetry-quake"
)

// CarrierFactory is a factory function to tell Tracing middleware how to fill the
// TraceContext from carrier. Checkout pkg/exporter/sentry/adaptor.go#CarrierFactory
// for detail.
type CarrierFactory func(h http.Header) propagation.TextMapCarrier

// Tracing creates a new otel.Tracer if never created and returns a gin.HandlerFunc.
// You only need to specify a CarrierFactory if your frontend doesn't obey TraceContext
// specification https://www.w3.org/TR/trace-context, otherwise you can leave it nil.
func Tracing(factory CarrierFactory) gin.HandlerFunc {
	_, err := opentelemetry.Setup()
	if err != nil {
		// FIXED(@yeqown): panic or log records this tip?
		fmt.Printf("WARNNING: setup opentelemetry failed: %v, this will cause traces collecting failed\n", err)
	}
	tc := propagation.TraceContext{}
	if factory == nil {
		factory = func(h http.Header) propagation.TextMapCarrier {
			return propagation.HeaderCarrier(h)
		}
	}

	return func(c *gin.Context) {
		// try to extract remote trace from request header.
		parentCtx := tc.Extract(c.Request.Context(), factory(c.Request.Header))
		ctx, sp := opentelemetry.StartSpan(parentCtx, c.FullPath(),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer sp.End()

		// inject trace context to gin context
		inject(c, ctx)

		c.Next()

		// TODO(@yeqown): add more tags and logs.
		sp.SetAttributes(
			attribute.Bool("http.status.success", c.Writer.Status() < 400),
			attribute.Int64("http.status.code", int64(c.Writer.Status())),
			attribute.String("http.status.message", http.StatusText(c.Writer.Status())),
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
		)
	}
}

const (
	OtelTraceContextKey = "opentelemetry.gin"
)

func inject(c *gin.Context, ctx context.Context) {
	c.Set(OtelTraceContextKey, ctx)
}

func extract(c *gin.Context) context.Context {
	v, ok := c.Get(OtelTraceContextKey)
	if !ok || v == nil {
		return nil
	}

	return v.(context.Context)
}

// StartSpan is a wrapper of opentelemetry.StartSpan, but it extracts span from gin.Context rather
// than context.Context. The return span is that derived by root span which is created by Tracing middleware.
func StartSpan(c *gin.Context, op string, opts ...trace.SpanStartOption) (ctx context.Context, sp trace.Span) {
	ctx, sp = opentelemetry.StartSpan(extract(c), op, opts...)
	return ctx, sp
}

func ContextFrom(c *gin.Context) context.Context {
	return extract(c)
}

// SpanFromContext get the raw span from gin.Context.
func SpanFromContext(c *gin.Context) trace.Span {
	return trace.SpanFromContext(extract(c))
}

// CaptureError captures error and panic to open-telemetry.
func CaptureError() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get current span to record, if span is nil then return directly.
		sp := SpanFromContext(c)
		if sp == nil {
			c.Next()
			return
		}

		defer func() {
			if err := recover(); err != nil {
				// FIXED(@yeqown): record stack trace.
				sp.RecordError(fmt.Errorf("%v", err), trace.WithStackTrace(true))
				// TODO(@yeqown): let user to decide whether re-panic or not.
			}
		}()

		c.Next()

		if c.Writer.Status() >= 400 {
			sp.RecordError(fmt.Errorf("%v", c.Writer.Status()))
			sp.SetStatus(codes.Error, "ERR")
		} else {
			sp.SetStatus(codes.Ok, "OK")
		}
	}
}
