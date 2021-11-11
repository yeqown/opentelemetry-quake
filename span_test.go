package otelquake_test

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"

	otelquake "github.com/yeqown/opentelemetry-quake"
)

func ExampleStartSpan() {
	ctx, sp := otelquake.StartSpan(context.Background(), "example")
	defer sp.End()

	remoteCall := func(ctx context.Context) {
		// launch a RPC call
	}

	// pass ctx to another internal call or remote call.
	remoteCall(ctx)
}

func ExampleSpanFromContext() {
	ctx, sp := otelquake.StartSpan(context.Background(), "example")
	defer sp.End()

	internalCall := func(ctx context.Context) {
		// launch a RPC call
		sp2 := otelquake.SpanFromContext(ctx)
		sp2.RecordError(fmt.Errorf("encounter an error in internalCall"), trace.WithStackTrace(true))
	}

	// pass ctx to another internal call or remote call.
	internalCall(ctx)
}
