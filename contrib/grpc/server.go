package tracinggrpc

import (
	"context"
	"fmt"

	tracing "github.com/yeqown/opentelemetry-quake"

	conventions "go.opentelemetry.io/collector/model/semconv/v1.5.0"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

// TracingServerInterceptor returns a grpc.UnaryServerInterceptor suitable
// for use in a grpc.NewServer call.
//
// All gRPC server spans will look for an OpenTracing TraceContext in the gRPC
// metadata; if found, the server span will act as the ChildOf that RPC
// TraceContext.
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
		ctxWithSpan, serverSpan := tracing.StartSpan(extractSpanContext(ctx), info.FullMethod,
			tracing.WithSpanKind(tracing.SpanKindServer),
		)
		serverSpan.SetAttributes(attribute.Bool(conventions.AttributeRPCSystem, true))
		defer serverSpan.End()

		if opts.logPayloads {
			serverSpan.LogFields("request",
				attribute.String("raw", marshalPbMessage(req)),
			)
		}

		resp, err = handler(ctxWithSpan, req)
		if err == nil {
			serverSpan.LogFields("response",
				attribute.String("raw", marshalPbMessage(resp)),
			)
			serverSpan.SetStatus(tracing.OK, "")
			return resp, err
		}

		serverSpan.RecordError(err)
		serverSpan.SetStatus(tracing.Error, err.Error())
		return resp, err
	}
}

// ServerCaptureException 用于辅助 open telemetry 采集应用错误和异常信息
// 因此必须在 ServerTracing 之后调用
func ServerCaptureException(repanic bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// get current span to record, if span is nil then return directly.
		sp := tracing.SpanFromContext(ctx)
		if sp == nil {
			return handler(ctx, req)
		}

		defer func() {
			if r := recover(); r != nil {
				// FIXED(@yeqown): record stack trace.
				sp.RecordError(fmt.Errorf("panic: %v", r), tracing.WithStackTrace())
				if repanic {
					panic(r)
				}
			}
		}()

		resp, err = handler(ctx, req)
		if err != nil {
			sp.SetStatus(tracing.Error, "ERR")
			sp.RecordError(err)
		} else {
			sp.SetStatus(tracing.OK, "OK")
		}
		return resp, err
	}
}
