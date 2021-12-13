package tracing_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tracing "github.com/yeqown/opentelemetry-quake"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"
)

func ExampleStartSpan() {
	ctx, sp := tracing.StartSpan(context.Background(), "example")
	defer sp.End()

	remoteCall := func(ctx context.Context) {
		// launch a RPC call
	}

	// pass ctx to another internal call or remote call.
	remoteCall(ctx)
}

func ExampleSpanFromContext() {
	ctx, sp := tracing.StartSpan(context.Background(), "example")
	defer sp.End()

	internalCall := func(ctx context.Context) {
		// launch a RPC call
		sp2 := tracing.SpanFromContext(ctx)
		sp2.RecordError(fmt.Errorf("encounter an error in internalCall"))
	}

	// pass ctx to another internal call or remote call.
	internalCall(ctx)
}

func ExampleSpanContextFromContext() {
	ctx, sp := tracing.StartSpan(context.Background(), "example")
	defer sp.End()

	remoteCall := func(ctx context.Context) {
		// launch a RPC call
		sc := tracing.SpanContextFromContext(ctx)
		fmt.Println(sc.TraceID)
	}

	// pass ctx to another internal call or remote call.
	remoteCall(ctx)
}

func ExampleTraceContextFromContext() {
	ctx, sp := tracing.StartSpan(context.Background(), "example")
	defer sp.End()

	remoteCall := func(ctx context.Context) {
		// launch a RPC call
		tc := tracing.TraceContextFromContext(ctx)
		fmt.Println(tc.TraceID)
	}

	// pass ctx to another internal call or remote call.
	remoteCall(ctx)
}

type testContextSuite struct {
	suite.Suite
	shutdown func()
}

func (t testContextSuite) TearDownSuite() {
	t.shutdown()
}

func (t *testContextSuite) SetupSuite() {
	shutdown, err := tracing.SetupDefault()
	if err != nil {
		panic(err)
	}

	t.shutdown = shutdown
}

func (t testContextSuite) Test_Compare_spanContext() {
	ctx, sp := tracing.StartSpan(context.Background(), "example")
	defer sp.End()

	sc := tracing.SpanContextFromContext(ctx)
	sc2 := trace.SpanContextFromContext(ctx)

	t.Equal(sc2.TraceID().String(), sc.TraceID)
	t.Equal(sc2.SpanID().String(), sc.SpanID)
	t.Equal(sc2.IsSampled(), sc.Sampled())
	t.Equal(sc2.IsRemote(), sc.IsRemote())
}

func (t testContextSuite) Test_TraceContextFromContext() {
	ctx, sp := tracing.StartSpan(context.Background(), "example")
	defer sp.End()
	tc1 := sp.SpanContext()
	tc1_2 := tracing.TraceContextFromContext(ctx)
	t.Equal(tc1, tc1_2)
	t.Equal(strings.Repeat("0", 16), tc1.ParentSpanID)

	ctx2, sp2 := tracing.StartSpan(ctx, "test")
	defer sp2.End()
	_ = ctx2
	tc2 := sp2.SpanContext()
	tc2_2 := tracing.TraceContextFromContext(ctx2)
	t.Equal(tc2, tc2_2)

	t.NotEmpty(tc1)
	t.NotEmpty(tc2)
	t.Equal(tc1.TraceID, tc2.TraceID)
	t.NotEqual(tc1.SpanID, tc2.SpanID)
	t.NotEqual(tc1.ParentSpanID, tc2.ParentSpanID)
	t.Equal(tc1.SpanID, tc2.ParentSpanID)
}

func Test_contextSuite(t *testing.T) {
	suite.Run(t, new(testContextSuite))
}
