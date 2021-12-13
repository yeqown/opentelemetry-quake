package tracingresty

import (
	tracing "github.com/yeqown/opentelemetry-quake"
	"github.com/yeqown/opentelemetry-quake/pkg"

	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel/attribute"
)

// Tracing middleware for resty client.
func genPreRequestMiddleware() resty.RequestMiddleware {

	return func(client *resty.Client, request *resty.Request) error {
		// 1. start a new span from request context.
		// 2. inject trace info into request header
		ctx := request.Context()
		ctx, sp := tracing.StartSpan(ctx, "resty.request",
			tracing.WithSpanKind(tracing.SpanKindClient),
		)

		request.SetContext(ctx)
		tracing.GetPropagator().Inject(ctx, request.Header)

		attrs := []attribute.KeyValue{
			//attribute.String("raw",), TODO: get request data from request.Body
			attribute.String("method", request.Method),
			attribute.String("url", request.URL),
		}
		if request.QueryParam != nil {
			attrs = append(attrs, attribute.String("query", request.QueryParam.Encode()))
		}
		if request.FormData != nil {
			attrs = append(attrs, attribute.String("form", request.FormData.Encode()))
		}

		sp.LogFields("request", attrs...)

		return nil
	}
}

func genPostRequestMiddleware() resty.ResponseMiddleware {
	return func(client *resty.Client, response *resty.Response) error {
		// 1. extract span from context
		// 2. finish span and record response
		ctx := response.Request.Context()
		sp := tracing.SpanFromContext(ctx)
		defer sp.End()

		sp.LogFields("response",
			attribute.String("raw", pkg.ToString(response.Body())),
			attribute.String("status", response.Status()),
		)
		if response.StatusCode() >= 400 {
			sp.SetStatus(tracing.Error, response.Status())
			return nil
		}

		sp.SetStatus(tracing.OK, "")
		return nil
	}
}

func genTracingErrorHook() resty.ErrorHook {
	return func(request *resty.Request, err error) {
		ctx := request.Context()
		sp := tracing.SpanFromContext(ctx)
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
