package otelgrpc

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// A Class is a set of types of outcomes (including errors) that will often
// be handled in the same way.
type Class string

const (
	Unknown Class = "0xx"
	// Success represents outcomes that achieved the desired results.
	Success Class = "2xx"
	// ClientError represents errors that were the client's fault.
	ClientError Class = "4xx"
	// ServerError represents errors that were the server's fault.
	ServerError Class = "5xx"
)

// ErrorClass returns the class of the given error
func ErrorClass(err error) Class {
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		// Success or "success"
		case codes.OK, codes.Canceled:
			return Success

		// Client errors
		case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
			codes.PermissionDenied, codes.Unauthenticated, codes.FailedPrecondition,
			codes.OutOfRange:
			return ClientError

		// Server errors
		case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted,
			codes.Unimplemented, codes.Internal, codes.Unavailable, codes.DataLoss:
			return ServerError

		// Not sure
		case codes.Unknown:
			fallthrough
		default:
			return Unknown
		}
	}
	return Unknown
}

// setSpanTags sets one or more tags on the given span according to the
// error.
func setSpanTags(span trace.Span, err error, client bool) {
	c := ErrorClass(err)
	code := codes.Unknown
	if s, ok := status.FromError(err); ok {
		code = s.Code()
	}

	span.SetAttributes(
		attribute.String("response_code", code.String()),
		attribute.String("response_class", string(c)),
	)

	if err == nil {
		return
	}
	if client || c == ServerError {
		span.RecordError(err)
	}
}
