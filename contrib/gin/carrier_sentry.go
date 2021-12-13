package tracinggin

import (
	"net/http"
	"strings"

	tracing "github.com/yeqown/opentelemetry-quake"
)

// sentryAdapter is an adapter that let sentry trace is valid for open-telemetry
// specification. open-telemetry specification is based on https://www.w3.org/TR/trace-context/#traceparent-header.
type sentryAdapter struct {
	http.Header
}

// Get adaptor, theory links:
// - go.opentelemetry.io/otel@v1.1.0/propagation/trace_context.go#L29-30
// - go.opentelemetry.io/otel@v1.1.0/propagation/trace_context.go#L82
// - go.opentelemetry.io/otel@v1.1.0/propagation/trace_context.go#L145
func (sa sentryAdapter) Get(key string) string {
	return sa.Header.Get(key)
}

func (sa sentryAdapter) Set(key string, value string) {
	sa.Header.Set(key, value)
}

func (sa sentryAdapter) Keys() []string {
	keys := make([]string, 0, len(sa.Header))
	for k := range sa.Header {
		keys = append(keys, k)
	}
	return keys
}

// SentryCarrierAdaptor is an adapter that let sentry trace is valid for open-telemetry.
func SentryCarrierAdaptor(h http.Header) tracing.TraceContextCarrier {
	carrier := sentryAdapter{Header: h}

	if value := h.Get("sentry-trace"); value != "" {
		carrier.Set("traceparent", translateSentryToOpenTelemetry(value))
	}

	return &carrier
}

// translateSentryToOpenTelemetry matches either
//
// 	TRACE_ID - SPAN_ID
// 	[[:xdigit:]]{32}-[[:xdigit:]]{16}
// or
// 	TRACE_ID - SPAN_ID - SAMPLED
// 	[[:xdigit:]]{32}-[[:xdigit:]]{16}-[01]
//
// var sentryTracePattern = regexp.MustCompile(`^([[:xdigit:]]{32})-([[:xdigit:]]{16})(?:-([01]))?$`)
//
// links:
// - github.com/sentry/sentry-go/tracing.go#L229-259
func translateSentryToOpenTelemetry(sentryTrace string) string {
	if len(sentryTrace) == 0 {
		return ""
	}

	arr := strings.Split(sentryTrace, "-")
	if len(arr) != 3 {
		return sentryTrace
	}
	switch arr[2] {
	case "1":
		arr[2] = "01"
	case "0":
		arr[2] = "00"
	}
	arr2 := append([]string{"00"}, arr...)

	return strings.Join(arr2, "-")
}
