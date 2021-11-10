package otelgrpc

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// TracingClientInterceptor returns a grpc.UnaryClientInterceptor suitable
// for use in a grpc.Dial call.
//
// All gRPC client spans will inject the OpenTracing SpanContext into the gRPC
// metadata; they will also look in the context.Context for an active
// in-process parent Span and establish a ChildOf reference if such a parent
// Span could be found.
func TracingClientInterceptor(optFuncs ...Option) grpc.UnaryClientInterceptor {
	traceOpts := newOptions()
	traceOpts.apply(optFuncs...)

	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) (err error) {
		parent := trace.SpanFromContext(ctx)
		if !parent.SpanContext().IsValid() {
			// has no parent span, just skip tracing
			return invoker(ctx, method, req, resp, cc, opts...)
		}

		//if opts.inclusionFunc != nil &&
		//	!opts.inclusionFunc(parentCtx, method, req, resp) {
		//	return invoker(ctx, method, req, resp, cc, opts...)
		//}

		ctxWithSpan, clientSpan := otel.Tracer(tracerName).
			Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
		defer clientSpan.End()
		ctx = injectSpanContext(ctxWithSpan)

		if traceOpts.logPayloads {
			clientSpan.AddEvent("start", trace.WithAttributes(
				attribute.String("request", marshalPbMessage(req))),
			)
		}

		err = invoker(ctx, method, req, resp, cc, opts...)
		if err == nil {
			if traceOpts.logPayloads {
				clientSpan.AddEvent("start", trace.WithAttributes(
					attribute.String("response", marshalPbMessage(resp))),
				)
			}
		} else {
			clientSpan.RecordError(err, trace.WithAttributes(
				attribute.String("error.message", err.Error()),
			))
		}

		//if traceOpts.decorator != nil {
		//	traceOpts.decorator(clientSpan, method, req, resp, err)
		//}

		return err
	}
}
