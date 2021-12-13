package tracinggrpc

import (
	"context"
	"fmt"

	tracing "github.com/yeqown/opentelemetry-quake"

	conventions "go.opentelemetry.io/collector/model/semconv/v1.5.0"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

// TracingClientInterceptor returns a grpc.UnaryClientInterceptor suitable
// for use in a grpc.Dial call.
//
// All gRPC client spans will inject the OpenTracing TraceContext into the gRPC
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
		parent := tracing.SpanFromContext(ctx)
		if !parent.SpanContext().IsValid() {
			// has no parent span, just skip tracing
			return invoker(ctx, method, req, resp, cc, opts...)
		}

		//if opts.inclusionFunc != nil &&
		//	!opts.inclusionFunc(parentCtx, method, req, resp) {
		//	return invoker(ctx, method, req, resp, cc, opts...)
		//}

		ctxWithSpan, clientSpan := tracing.StartSpan(ctx, method,
			tracing.WithSpanKind(tracing.SpanKindClient),
		)
		clientSpan.SetAttributes(attribute.Bool(conventions.AttributeRPCService, true))
		defer clientSpan.End()
		ctx = injectSpanContext(ctxWithSpan)

		if traceOpts.logPayloads {
			clientSpan.LogFields("request",
				attribute.String("raw", marshalPbMessage(req)),
			)
		}

		err = invoker(ctx, method, req, resp, cc, opts...)
		if err == nil {
			if traceOpts.logPayloads {
				clientSpan.LogFields("response",
					attribute.String("raw", marshalPbMessage(resp)),
				)
			}
			clientSpan.SetStatus(tracing.OK, "")
			return nil
		}

		clientSpan.RecordError(err)
		clientSpan.SetStatus(tracing.Error, err.Error())
		return err
	}
}

// ClientCaptureException 用于辅助 open telemetry 采集应用错误和异常信息
// 因此必须在 ClientTracing 之后调用
func ClientCaptureException(repanic bool) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) error {
		// get current span to record, if span is nil then return directly.
		sp := tracing.SpanFromContext(ctx)
		if sp == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		defer func() {
			if r := recover(); r != nil {
				sp.RecordError(fmt.Errorf("panic: %v", r), tracing.WithStackTrace())
				if repanic {
					panic(r)
				}
			}
		}()

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			sp.SetStatus(tracing.Error, "ERR")
			sp.RecordError(err)
		} else {
			sp.SetStatus(tracing.OK, "OK")
		}
		return err
	}
}
