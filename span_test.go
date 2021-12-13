package tracing_test

import (
	"context"
	"strings"
	"testing"

	tracing "github.com/yeqown/opentelemetry-quake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
)

type spanTestSuite struct {
	suite.Suite

	shutdown func()
}

func (s *spanTestSuite) TearDownSuite() {
	s.shutdown()
}

func (s *spanTestSuite) SetupSuite() {
	shutdown, err := tracing.SetupDefault()
	if err != nil {
		panic(err)
	}

	s.shutdown = shutdown
}

func (s spanTestSuite) Test_spanAgent() {
	t := s.T()

	ctx, sp := tracing.StartSpan(context.Background(), "test", tracing.WithSpanKind(tracing.SpanKindClient))
	defer sp.Finish()

	tc1 := tracing.SpanContextFromContext(ctx)
	tc1_2 := tracing.TraceContextFromContext(ctx)
	assert.NotNil(t, tc1)
	assert.NotEmpty(t, tc1.TraceID)
	assert.NotEmpty(t, tc1.SpanID)
	assert.NotEqual(t, strings.Repeat("0", 32), tc1.TraceID)
	assert.Equal(t, tc1, tc1_2)

	ctx2, sp2 := tracing.StartSpan(ctx, "test2")
	defer sp2.Finish()
	assert.Equal(t, sp.SpanContext().TraceID, sp2.SpanContext().TraceID)
	tc2 := tracing.SpanContextFromContext(ctx2)
	assert.NotNil(t, tc2)
	assert.Equal(t, tc1.TraceID, tc2.TraceID)
	assert.NotEqual(t, tc1.SpanID, tc2.SpanID)
}

func (s spanTestSuite) Test_spanAgent_recording() {
	t := s.T()

	ctx, sp := tracing.StartSpan(context.Background(), "test", tracing.WithSpanKind(tracing.SpanKindClient))
	defer sp.Finish()

	assert.NotPanics(t, func() {
		sp.SetTag("key", "value")
		sp.RecordError(nil)
		sp.SetStatus(tracing.OK, "ok")
		sp.LogFields("start", attribute.Bool("key", true), attribute.String("key2", "value2"))
	})

	_ = ctx
}

func Test_suite(t *testing.T) {
	suite.Run(t, new(spanTestSuite))
}
