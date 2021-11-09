// Package sentryexporter is a sentry Go SDK client to implement the
// trace.SpanExporter interface.
//
// references:
// - https://develop.sentry.dev/sdk/overview/#parsing-the-dsn
// - https://develop.sentry.dev/sdk/store/
package sentry

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	conventions "go.opentelemetry.io/collector/model/semconv/v1.5.0"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	otelSentryExporterVersion = "0.0.2"
	otelSentryExporterName    = "sentry.opentelemetry"
)

var _ sdktrace.SpanExporter = (*Exporter)(nil)

// canonicalCodes maps OpenTelemetry span codes to Sentry's span status.
// See numeric codes in https://github.com/open-telemetry/opentelemetry-proto/blob/6cf77b2f544f6bc7fe1e4b4a8a52e5a42cb50ead/opentelemetry/proto/trace/v1/trace.proto#L303
var canonicalCodes = [...]sentry.SpanStatus{
	sentry.SpanStatusUndefined,
	sentry.SpanStatusOK,
	sentry.SpanStatusUnknown,
}

func New(dsn string) (*Exporter, error) {
	co := sentry.ClientOptions{
		Dsn: dsn,
		HTTPTransport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	t := newTransport()
	t.Configure(co)

	return &Exporter{
		transport: t,
	}, nil
}

type Exporter struct {
	transport transport
}

// ExportSpans takes an incoming OpenTelemetry trace, converts them into Sentry spans and transactions
// and sends them using Sentry's transport.
func (e Exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if len(spans) == 0 {
		return nil
	}

	exceptionEvents := make([]*sentry.Event, 0, len(spans))
	maybeOrphanSpans := make([]*sentry.Span, 0, len(spans))
	// Maps all child span ids to their root span.
	idMap := make(map[sentry.SpanID]sentry.SpanID)
	// Maps root span id to a transaction.
	transactionMap := make(map[sentry.SpanID]*sentry.Event)
	for k := 0; k < len(spans); k++ {
		otelSpan := spans[k]
		sentrySpan := convertToSentrySpan(otelSpan)
		convertEventsToSentryExceptions(&exceptionEvents, otelSpan.Events(), sentrySpan)

		// If a span is a root span, we consider it the start of a Sentry transaction.
		// We should then create a new transaction for that root span, and keep track of it.
		//
		// If the span is not a root span, we can either associate it with an existing
		// transaction, or we can temporarily consider it an orphan span.
		if spanIsTransaction(otelSpan) {
			transactionMap[sentrySpan.SpanID] = transactionFromSpan(sentrySpan)
			idMap[sentrySpan.SpanID] = sentrySpan.SpanID
		} else {
			if rootSpanID, ok := idMap[sentrySpan.ParentSpanID]; ok {
				idMap[sentrySpan.SpanID] = rootSpanID
				transactionMap[rootSpanID].Spans = append(transactionMap[rootSpanID].Spans, sentrySpan)
			} else {
				maybeOrphanSpans = append(maybeOrphanSpans, sentrySpan)
			}
		}
	}

	if len(transactionMap) == 0 {
		return nil
	}

	// After the first pass through, we can't necessarily make the assumption we have not associated all
	// the spans with a transaction. As such, we must classify the remaining spans as orphans or not.
	orphanSpans := classifyAsOrphanSpans(maybeOrphanSpans, len(maybeOrphanSpans)+1, idMap, transactionMap)
	transactions := generateTransactions(transactionMap, orphanSpans)
	events := append(transactions, exceptionEvents...)
	e.transport.SendEvents(events)

	return nil
}

func (e Exporter) Shutdown(ctx context.Context) error {
	allEventsFlushed := e.transport.Flush(ctx)

	if !allEventsFlushed {
		log.Print("Could not flush all events, reached timeout")
	}

	return nil
}

// transactionFromSpan converts a span to a transaction.
func transactionFromSpan(span *sentry.Span) *sentry.Event {
	transaction := sentry.NewEvent()
	transaction.EventID = generateEventID()

	transaction.Contexts["trace"] = sentry.TraceContext{
		TraceID:      span.TraceID,
		SpanID:       span.SpanID,
		ParentSpanID: span.ParentSpanID,
		Op:           span.Op,
		Description:  span.Description,
		Status:       span.Status,
	}

	transaction.Type = "transaction"

	transaction.Sdk.Name = otelSentryExporterName
	transaction.Sdk.Version = otelSentryExporterVersion

	transaction.StartTime = span.StartTime
	transaction.Tags = span.Tags
	transaction.Timestamp = span.EndTime
	transaction.Transaction = span.Description

	return transaction
}

func uuid() string {
	id := make([]byte, 16)
	// Prefer rand.Read over rand.Reader, see https://go-review.googlesource.com/c/go/+/272326/.
	_, _ = rand.Read(id)
	id[6] &= 0x0F // clear version
	id[6] |= 0x40 // set version to 4 (random uuid)
	id[8] &= 0x3F // clear variant
	id[8] |= 0x80 // set to IETF variant
	return hex.EncodeToString(id)
}

func generateEventID() sentry.EventID {
	return sentry.EventID(uuid())
}

// spanIsTransaction determines if a span should be sent to Sentry as a transaction.
// If parent span id is empty or the span kind allows remote parent spans, then the span is a root span.
func spanIsTransaction(span sdktrace.ReadOnlySpan) bool {
	kind := span.SpanKind()
	return !span.Parent().IsValid() || kind == trace.SpanKindServer || kind == trace.SpanKindConsumer
}

func generateTagsFromAttributes(attrs []attribute.KeyValue) map[string]string {
	tags := make(map[string]string)

	for _, attr := range attrs {
		key := string(attr.Key)
		switch attr.Value.Type() {
		case attribute.STRING:
			tags[key] = attr.Value.AsString()
		case attribute.BOOL:
			tags[key] = strconv.FormatBool(attr.Value.AsBool())
		case attribute.FLOAT64:
			tags[key] = strconv.FormatFloat(attr.Value.AsFloat64(), 'g', -1, 64)
		case attribute.INT64:
			tags[key] = strconv.FormatInt(attr.Value.AsInt64(), 10)
		}
	}

	return tags
}

// generateSpanDescriptors generates span descriptors (op and description)
// from the name, attributes and SpanKind of an otel span based onSemantic Conventions
// described by the open telemetry specification.
//
// See https://github.com/open-telemetry/opentelemetry-specification/tree/5b78ee1/specification/trace/semantic_conventions
// for more details about the semantic conventions.
func generateSpanDescriptors(name string, attrs []attribute.KeyValue, spanKind trace.SpanKind) (op string, description string) {
	var opBuilder strings.Builder
	var dBuilder strings.Builder

	// Generating span descriptors operates under the assumption that only one of the conventions are present.
	// In the possible case that multiple convention attributes are available, conventions are selected based
	// on what is most likely and what is most useful (ex. http is prioritized over FaaS)

	// If http.method exists, this is a http request span.
	for _, attr := range attrs {
		switch attr.Key {
		case conventions.AttributeHTTPMethod:
			opBuilder.WriteString("http")
			switch spanKind {
			case trace.SpanKindClient:
				opBuilder.WriteString(".client")
			case trace.SpanKindServer:
				opBuilder.WriteString(".server")
			}
			// Ex. description="GET /api/users/{user_id}".
			_, _ = fmt.Fprintf(&dBuilder, "%s %s", attr.Value.AsString(), name)
			return opBuilder.String(), dBuilder.String()
		case conventions.AttributeDBSystem:
			opBuilder.WriteString("db")
			dBuilder.WriteString(attr.Value.AsString())
			// Use DB statement (Ex "SELECT * FROM table") if possible as description.
			//if statement, okInst := attrs.Get(conventions.AttributeDBStatement); okInst {
			//	dBuilder.WriteString(statement.StringVal())
			//} else {
			//	dBuilder.WriteString(name)
			//}
			return opBuilder.String(), dBuilder.String()
		case conventions.AttributeRPCService:
			opBuilder.WriteString("rpc")
			return opBuilder.String(), name
		case "messaging.system":
			opBuilder.WriteString("message")
			return opBuilder.String(), name
		case "faas.trigger":
			opBuilder.WriteString(attr.Value.AsString())
			return opBuilder.String(), name
		}
	}

	// Default just use span.name.
	return "", name
}

func convertToSentrySpan(span sdktrace.ReadOnlySpan) (sentrySpan *sentry.Span) {
	attributes := span.Attributes()
	name := span.Name()
	spanKind := span.SpanKind()
	library := span.InstrumentationLibrary()

	op, description := generateSpanDescriptors(name, attributes, spanKind)
	tags := generateTagsFromAttributes(attributes)

	//for k, v := range resourceTags {
	//	tags[k] = v
	//}

	status, message := statusFromSpanStatus(span.Status())
	if message != "" {
		tags["status_message"] = message
	}

	if spanKind != trace.SpanKindUnspecified {
		tags["span_kind"] = spanKind.String()
	}

	tags["library_name"] = library.Name
	tags["library_version"] = library.Version

	sentrySpan = &sentry.Span{
		TraceID:     sentry.TraceID(span.SpanContext().TraceID()),
		SpanID:      sentry.SpanID(span.SpanContext().SpanID()),
		Description: description,
		Op:          op,
		Tags:        tags,
		StartTime:   span.StartTime(),
		EndTime:     span.EndTime(),
		Status:      status,
	}

	if parent := span.Parent(); parent.IsValid() {
		sentrySpan.ParentSpanID = sentry.SpanID(parent.SpanID())
	}

	return sentrySpan
}

func statusFromSpanStatus(s sdktrace.Status) (status sentry.SpanStatus, message string) {
	code := s.Code
	if code < 0 || int(code) >= len(canonicalCodes) {
		return sentry.SpanStatusUnknown, fmt.Sprintf("error code %d", code)
	}

	return canonicalCodes[code], s.Description
}

// convertEventsToSentryExceptions creates a set of sentry events from exception events present in spans.
// These events are stored in a mutated eventList
func convertEventsToSentryExceptions(eventList *[]*sentry.Event, events []sdktrace.Event, sentrySpan *sentry.Span) {
	for i := 0; i < len(events); i++ {
		event := events[i]
		if event.Name != "exception" {
			continue
		}
		var exceptionMessage, exceptionType, exceptionStack string
		for _, attr := range event.Attributes {
			switch string(attr.Key) {
			case conventions.AttributeExceptionMessage:
				exceptionMessage = attr.Value.AsString()
			case conventions.AttributeExceptionType:
				exceptionType = attr.Value.AsString()
			case conventions.AttributeExceptionStacktrace:
				exceptionStack = attr.Value.AsString()
			}
		}

		if exceptionMessage == "" && exceptionType == "" {
			// `At least one of the following sets of attributes is required:
			// - exception.type
			// - exception.message`
			continue
		}
		sentryEvent, _ := sentryEventFromError(exceptionMessage, exceptionType, exceptionStack, sentrySpan)
		*eventList = append(*eventList, sentryEvent)
	}
}

// sentryEventFromError creates a sentry event from error event in a span
func sentryEventFromError(errorMessage, errorType, stack string, span *sentry.Span) (*sentry.Event, error) {
	if errorMessage == "" && errorType == "" {
		err := errors.New("error type and error message were both empty")
		return nil, err
	}
	event := sentry.NewEvent()
	event.EventID = generateEventID()

	event.Contexts["trace"] = sentry.TraceContext{
		TraceID:      span.TraceID,
		SpanID:       span.SpanID,
		ParentSpanID: span.ParentSpanID,
		Op:           span.Op,
		Description:  span.Description,
		Status:       span.Status,
	}

	event.Type = errorType
	event.Message = errorMessage
	if len(stack) != 0 {
		event.Message = stack
	}
	event.Level = "error"
	event.Exception = []sentry.Exception{{
		Value: errorMessage,
		Type:  errorType,
	}}

	event.Sdk.Name = otelSentryExporterName
	event.Sdk.Version = otelSentryExporterVersion

	event.StartTime = span.StartTime
	event.Tags = span.Tags
	event.Timestamp = span.EndTime
	event.Transaction = span.Description

	return event, nil
}

// classifyAsOrphanSpans iterates through a list of possible orphan spans and tries to associate them
// with a transaction. As the order of the spans is not guaranteed, we have to recursively call
// classifyAsOrphanSpans to make sure that we did not leave any spans out of the transaction they belong to.
func classifyAsOrphanSpans(
	orphanSpans []*sentry.Span,
	prevLength int,
	idMap map[sentry.SpanID]sentry.SpanID,
	transactionMap map[sentry.SpanID]*sentry.Event,
) []*sentry.Span {
	if len(orphanSpans) == 0 || len(orphanSpans) == prevLength {
		return orphanSpans
	}

	newOrphanSpans := make([]*sentry.Span, 0, prevLength)

	for _, orphanSpan := range orphanSpans {
		if rootSpanID, ok := idMap[orphanSpan.ParentSpanID]; ok {
			idMap[orphanSpan.SpanID] = rootSpanID
			transactionMap[rootSpanID].Spans = append(transactionMap[rootSpanID].Spans, orphanSpan)
		} else {
			newOrphanSpans = append(newOrphanSpans, orphanSpan)
		}
	}

	return classifyAsOrphanSpans(newOrphanSpans, len(orphanSpans), idMap, transactionMap)
}

// generateTransactions creates a set of Sentry transactions from a transaction map and orphan spans.
func generateTransactions(transactionMap map[sentry.SpanID]*sentry.Event, orphanSpans []*sentry.Span) []*sentry.Event {
	transactions := make([]*sentry.Event, 0, len(transactionMap)+len(orphanSpans))

	for _, t := range transactionMap {
		transactions = append(transactions, t)
	}

	for _, orphanSpan := range orphanSpans {
		t := transactionFromSpan(orphanSpan)
		transactions = append(transactions, t)
	}

	return transactions
}
