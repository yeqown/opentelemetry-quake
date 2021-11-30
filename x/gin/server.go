package otelgin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	conventions "go.opentelemetry.io/collector/model/semconv/v1.5.0"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	otelquake "github.com/yeqown/opentelemetry-quake"
	"github.com/yeqown/opentelemetry-quake/pkg"
)

// CarrierFactory is a factory function to tell Tracing middleware how to fill the
// TraceContext from carrier. Checkout pkg/exporter/sentry/adaptor.go#CarrierFactory
// for detail.
type CarrierFactory func(h http.Header) propagation.TextMapCarrier

// Config helps user to control Tracing middleware about how to
// handle the request and response. Such as:
// - log the request and response body or not,
// - how to extract TraceContext from request.
type Config struct {
	Factory                 CarrierFactory
	LogRequest, LogResponse bool
}

func (c *Config) ApplyCarrierFactory(factory CarrierFactory) *Config {
	if factory == nil {
		return c
	}
	c.Factory = factory

	return c
}

func (c *Config) EnableLogPayloads() *Config {
	c.LogRequest = true
	c.LogResponse = true

	return c
}

func DefaultConfig() *Config {
	return &Config{
		Factory:     builtinCarrierFactory,
		LogRequest:  false,
		LogResponse: false,
	}
}

func builtinCarrierFactory(h http.Header) propagation.TextMapCarrier {
	return propagation.HeaderCarrier(h)
}

// Tracing creates a new otel.Tracer if never created and returns a gin.HandlerFunc.
// You only need to specify a CarrierFactory if your frontend doesn't obey TraceContext
// specification https://www.w3.org/TR/trace-context, otherwise you can leave it nil.
func Tracing(config *Config) gin.HandlerFunc {
	if config == nil {
		config = DefaultConfig()
	}
	factory := config.Factory
	if factory == nil {
		factory = builtinCarrierFactory
	}

	tc := propagation.TraceContext{}

	return func(c *gin.Context) {
		// try to extract remote trace from request header.
		parentCtx := tc.Extract(c.Request.Context(), factory(c.Request.Header))
		ctx, sp := otelquake.StartSpan(parentCtx, c.FullPath(),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attribute.String(conventions.AttributeHTTPMethod, c.Request.Method)),
		)
		defer sp.End()

		println("traceId: ", sp.SpanContext().TraceID().String())
		println("sampled: ", sp.SpanContext().TraceFlags().IsSampled())

		// inject trace context to gin context
		inject(c, ctx)

		// FIXME(@yeqown): root span could not hold events, so we need to create a child span to hold events.
		//_, sp2 := otelquake.StartSpan(ctx, "logger")
		//defer sp2.End()

		// use custom writer, so we record the response body.
		rbw := &respBodyWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = rbw

		if config.LogRequest {
			body, err := c.GetRawData()
			if err == nil && len(body) != 0 {
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}

			// add start event and record request body.
			sp.AddEvent("request",
				trace.WithAttributes(attribute.String("raw", pkg.ToString(body))),
			)
		}

		c.Next()

		// add end event and record response body.
		if config.LogResponse {
			sp.AddEvent("response",
				trace.WithAttributes(attribute.String("raw", rbw.body.String())),
			)
		}

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

		var (
			r        interface{}
			panicked bool
		)

		defer func() {
			if r = recover(); r != nil {
				panicked = true
				// FIXED(@yeqown): record stack trace.
				sp.RecordError(fmt.Errorf("%v", r), trace.WithStackTrace(true))
				// TODO(@yeqown): let user to decide whether re-panic or not.
			}
		}()

		c.Next()

		if c.Writer.Status() >= 400 {
			sp.SetStatus(codes.Error, "ERR")
			sp.RecordError(fmt.Errorf("%v", c.Writer.Status()))
		} else {
			sp.SetStatus(codes.Ok, "OK")
		}
	}
}
