package sentry

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/propagation"
)

// HeaderCarrierAdapter is an adapter that let sentry trace is valid for open-telemetry
// specification. open-telemetry specification is based on https://www.w3.org/TR/trace-context/#traceparent-header.
type HeaderCarrierAdapter struct {
	http.Header
}

// Get adaptor, theory links:
// - go.opentelemetry.io/otel@v1.1.0/propagation/trace_context.go#L29-30
// - go.opentelemetry.io/otel@v1.1.0/propagation/trace_context.go#L82
// - go.opentelemetry.io/otel@v1.1.0/propagation/trace_context.go#L145
func (hca HeaderCarrierAdapter) Get(key string) string {
	return hca.Header.Get(key)
}

func (hca HeaderCarrierAdapter) Set(key string, value string) {
	hca.Header.Set(key, value)
}

func (hca HeaderCarrierAdapter) Keys() []string {
	keys := make([]string, 0, len(hca.Header))
	for k := range hca.Header {
		keys = append(keys, k)
	}
	return keys
}

func CarrierFactory(h http.Header) propagation.TextMapCarrier {
	carrier := HeaderCarrierAdapter{Header: h}

	if value := h.Get("sentry-trace"); value != "" {
		carrier.Set("traceparent", convertSentryTraceToParent(value))
	}

	return &carrier
}

// convertSentryTraceToParent matches either
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
func convertSentryTraceToParent(sentryTrace string) string {
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

//updateFromSentryTrace parses a sentry-trace HTTP header (as returned by
//ToSentryTrace) and updates fields of the span. If the header cannot be
//recognized as valid, the span is left unchanged.
//func updateFromSentryTrace(header []byte) {
//	m := sentryTracePattern.FindSubmatch(header)
//	if m == nil {
//		// no match
//		return
//	}
//
//	_, _ = hex.Decode(s.TraceID[:], m[1])
//	_, _ = hex.Decode(s.ParentSpanID[:], m[2])
//	if len(m[3]) != 0 {
//		switch m[3][0] {
//		case '0':
//			s.Sampled = SampledFalse
//		case '1':
//			s.Sampled = SampledTrue
//		}
//	}
//}
