package tracinggin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"

	tracing "github.com/yeqown/opentelemetry-quake"
	"github.com/yeqown/opentelemetry-quake/pkg"
)

// config helps user to control Tracing middleware about how to
// handle the request and response. Such as:
// - log the request and response body or not,
// - how to extract TraceContext from request.
type config struct {
	carrierFactory          func(h http.Header) tracing.TraceContextCarrier
	logRequest, logResponse bool
}

func defaultConfig() *config {
	return &config{
		carrierFactory: builtinCarrierFactory,
		logRequest:     false,
		logResponse:    false,
	}
}

func builtinCarrierFactory(h http.Header) tracing.TraceContextCarrier {
	return h
}

type TracingOption interface {
	apply(*config)
}

type fnTracingOption func(*config)

func (f fnTracingOption) apply(c *config) { f(c) }

func newFunctionalOption(f func(*config)) TracingOption {
	return fnTracingOption(f)
}

func WithCarrierFactory(adaptor func(h http.Header) tracing.TraceContextCarrier) TracingOption {
	return newFunctionalOption(func(c *config) {
		if adaptor != nil {
			c.carrierFactory = adaptor
		}
	})
}

func WithRecordPayloads() TracingOption {
	return newFunctionalOption(func(c *config) {
		c.logRequest = true
		c.logResponse = true
	})
}

// Tracing creates a new otel.Tracer if never created and returns a gin.HandlerFunc.
// You only need to specify a CarrierFactory if your frontend doesn't obey TraceContext
// specification https://www.w3.org/TR/trace-context, otherwise you can leave it nil.
func Tracing(opts ...TracingOption) gin.HandlerFunc {
	opt := defaultConfig()
	for _, o := range opts {
		o.apply(opt)
	}

	return func(c *gin.Context) {
		// try to extract remote trace from request header.
		parentCtx := tracing.
			GetPropagator().
			Extract(c.Request.Context(), opt.carrierFactory(c.Request.Header))

		ctx, sp := tracing.StartSpan(parentCtx, c.FullPath(),
			tracing.WithSpanKind(tracing.SpanKindServer),
		)
		defer sp.End()

		// inject trace context to gin context
		inject(c, ctx)

		rbw := getResponseBodyWriter(c)
		c.Writer = rbw

		if opt.logRequest {
			body, err := c.GetRawData()
			if err == nil && len(body) != 0 {
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}

			// add start event and record request body.
			sp.LogFields("request",
				attribute.String("query", c.Request.URL.RawQuery),
				attribute.String("raw", pkg.ToString(body)),
			)
		}

		c.Header("x-tracing-id", sp.SpanContext().TraceID)
		c.Next()

		// add end event and record response body.
		if opt.logResponse {
			sp.LogFields("response", attribute.String("raw", rbw.String()))
		}
		// 只是释放 respBodyWriter 中额外存储的内存空间，并不会影响底层的 ResponseWriter
		rbw.releaseBuffer()

		if c.Writer.Status() >= 400 {
			sp.SetStatus(tracing.Error, http.StatusText(c.Writer.Status()))
		} else {
			sp.SetStatus(tracing.OK, "")
		}

		sp.SetAttributes(
			attribute.Bool("http.status.success", c.Writer.Status() < 400),
			attribute.Int64("http.status.code", int64(c.Writer.Status())),
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
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
		return context.Background()
	}

	return v.(context.Context)
}

// TracingContextFrom read trace context from gin context. it never
// returns nil.
func TracingContextFrom(c *gin.Context) context.Context {
	return extract(c)
}

// CaptureException captures error and panic to open-telemetry.
func CaptureException(repanic bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get current span to record, if span is nil then return directly.
		sp := tracing.SpanFromContext(extract(c))
		if sp == nil {
			c.Next()
			return
		}

		defer func() {
			if r := recover(); r != nil {
				// FIXED(@yeqown): record stack trace.
				sp.RecordError(fmt.Errorf("panic %v", r), tracing.WithStackTrace())
				if repanic {
					panic(r)
				}
			}
		}()

		c.Next()

		if c.Writer.Status() >= 400 {
			sp.SetStatus(tracing.Error, "ERR")
			sp.RecordError(fmt.Errorf("%v", c.Writer.Status()))
		} else {
			sp.SetStatus(tracing.OK, "OK")
		}
	}
}
