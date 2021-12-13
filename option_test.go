package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StartSpanOption(t *testing.T) {
	o := defaultSpanStartOption()

	assert.Equal(t, SpanKindUnspecified, o.kind)
	WithSpanKind(SpanKindServer).apply(o)
	assert.Equal(t, SpanKindServer, o.kind)
}

func Test_SpanEventOption(t *testing.T) {
	o := defaultSpanEventOption()

	assert.Equal(t, false, o.withStackTrace)
	WithStackTrace().apply(o)
	assert.Equal(t, true, o.withStackTrace)
}
