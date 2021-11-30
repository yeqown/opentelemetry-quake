package otelresty

import (
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	otelquake "github.com/yeqown/opentelemetry-quake"
	"github.com/yeqown/opentelemetry-quake/pkg"
)

// Tracing middleware for resty client.
func genPreRequestMiddleware() resty.RequestMiddleware {
	tc := propagation.TraceContext{}

	return func(client *resty.Client, request *resty.Request) error {
		// 1. start a new span from request context.
		// 2. inject trace info into request header
		ctx := request.Context()
		ctx, sp := otelquake.StartSpan(ctx, request.URL,
			trace.WithSpanKind(trace.SpanKindClient),
		)

		request.SetContext(ctx)
		tc.Inject(ctx, propagation.HeaderCarrier(request.Header))

		sp.AddEvent("request",
			trace.WithTimestamp(time.Now()),
			trace.WithAttributes(
				//attribute.String("raw",), TODO: get request data from request.Body
				attribute.String("method", request.Method),
				attribute.String("url", request.URL),
			),
		)

		return nil
	}
}

func genPostRequestMiddleware() resty.ResponseMiddleware {
	return func(client *resty.Client, response *resty.Response) error {
		// 1. extract span from context
		// 2. finish span and record response
		ctx := response.Request.Context()
		sp := trace.SpanFromContext(ctx)
		defer sp.End()

		sp.AddEvent("response",
			trace.WithTimestamp(time.Now()),
			trace.WithAttributes(
				attribute.String("raw", pkg.ToString(response.Body())),
				attribute.String("status", response.Status()),
			),
		)

		return nil
	}
}

func genTracingErrorHook() resty.ErrorHook {
	return func(request *resty.Request, err error) {
		ctx := request.Context()
		sp := trace.SpanFromContext(ctx)
		defer sp.End()

		sp.RecordError(err)
	}
}

var (
	_singletonKeeper = map[*resty.Client]struct{}{}
)

// InjectTracing injects, should keep singleton in one resty.Client.
func InjectTracing(c *resty.Client) {
	if _, ok := _singletonKeeper[c]; ok {
		return
	}

	c.OnBeforeRequest(genPreRequestMiddleware())
	c.OnAfterResponse(genPostRequestMiddleware())
	c.OnError(genTracingErrorHook())

	_singletonKeeper[c] = struct{}{}
}
