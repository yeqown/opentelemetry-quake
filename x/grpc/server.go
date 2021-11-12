package otelgrpc

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// TracingServerInterceptor returns a grpc.UnaryServerInterceptor suitable
// for use in a grpc.NewServer call.
//
// All gRPC server spans will look for an OpenTracing SpanContext in the gRPC
// metadata; if found, the server span will act as the ChildOf that RPC
// SpanContext.
//
// Root or not, the server Span will be embedded in the context.Context for the
// application-specific gRPC handler(s) to access.
func TracingServerInterceptor(optFuncs ...Option) grpc.UnaryServerInterceptor {
	opts := newOptions()
	opts.apply(optFuncs...)

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// try to extract TraceContext from ctx
		ctxWithSpan, serverSpan := otel.Tracer(tracerName).
			Start(extractSpanContext(ctx), info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer serverSpan.End()

		if opts.logPayloads {
			serverSpan.AddEvent("start", trace.WithAttributes(
				attribute.String("request", marshalPbMessage(req))),
			)
		}

		resp, err = handler(ctxWithSpan, req)
		if err == nil {
			serverSpan.AddEvent("end", trace.WithAttributes(
				attribute.String("response", marshalPbMessage(resp))),
			)
		} else {
			serverSpan.RecordError(err, trace.WithAttributes(
				attribute.String("error.message", err.Error()),
			))
		}

		return resp, err
	}
}
