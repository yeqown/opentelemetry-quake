// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sentryexporter

import (
	"context"
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/model/pdata"
	conventions "go.opentelemetry.io/collector/model/semconv/v1.5.0"
)

const (
	otelSentryExporterVersion = "0.2.2"
	otelSentryExporterName    = "sentry.opentelemetry"
)

// canonicalCodes maps OpenTelemetry span codes to Sentry's span status.
// See numeric codes in https://github.com/open-telemetry/opentelemetry-proto/blob/6cf77b2f544f6bc7fe1e4b4a8a52e5a42cb50ead/opentelemetry/proto/trace/v1/trace.proto#L303
var canonicalCodes = [...]sentry.SpanStatus{
	sentry.SpanStatusUndefined,
	sentry.SpanStatusOK,
	sentry.SpanStatusCanceled,
	sentry.SpanStatusUnknown,
	sentry.SpanStatusInvalidArgument,
	sentry.SpanStatusDeadlineExceeded,
	sentry.SpanStatusNotFound,
	sentry.SpanStatusAlreadyExists,
	sentry.SpanStatusPermissionDenied,
	sentry.SpanStatusResourceExhausted,
	sentry.SpanStatusFailedPrecondition,
	sentry.SpanStatusAborted,
	sentry.SpanStatusOutOfRange,
	sentry.SpanStatusUnimplemented,
	sentry.SpanStatusInternalError,
	sentry.SpanStatusUnavailable,
	sentry.SpanStatusDataLoss,
	sentry.SpanStatusUnauthenticated,
}

// SentryExporter defines the Sentry Exporter.
type SentryExporter struct {
	transport transport
}

// newSentryExporter returns a new Sentry Exporter.
func newSentryExporter(config *Config, set component.ExporterCreateSettings) (component.TracesExporter, error) {
	tr := newSentryTransport()

	clientOptions := sentry.ClientOptions{
		Dsn: config.DSN,
	}

	if config.InsecureSkipVerify {
		clientOptions.HTTPTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}

	tr.Configure(clientOptions)

	s := &SentryExporter{
		transport: tr,
	}

	return exporterhelper.NewTracesExporter(
		config,
		set,
		s.consumeTraces,
		exporterhelper.WithShutdown(func(ctx context.Context) error {
			allEventsFlushed := tr.Flush(ctx)

			if !allEventsFlushed {
				println("[WARN] could not flush all events, reached timeout")
			}

			return nil
		}),
	)
}

// consumeTraces takes an incoming OpenTelemetry trace, converts them into Sentry spans and transactions
// and sends them using Sentry's transport.
func (s *SentryExporter) consumeTraces(_ context.Context, td pdata.Traces) error {
	events := traceToSentryEvents(td)
	s.transport.SendEvents(events)

	return nil
}

// generateSpanOperation generates span descriptors (op and description)
// from the name, attributes and SpanKind of an otel span based onSemantic Conventions
// described by the open telemetry specification.
//
// See https://github.com/open-telemetry/opentelemetry-specification/tree/5b78ee1/specification/trace/semantic_conventions
// for more details about the semantic conventions.
func generateSpanOperation(span *pdata.Span) (op string, description string) {
	attrs := span.Attributes()
	spanKind := span.Kind()
	op = span.Name()
	description = op

	// Generating span descriptors operates under the assumption that only one of the conventions are present.
	// In the possible case that multiple convention attributes are available, conventions are selected based
	// on what is most likely and what is most useful (ex. http is prioritized over FaaS)

	// If http.method exists, this is a http request span.
	if httpMethod, ok := attrs.Get(conventions.AttributeHTTPMethod); ok {
		op = httpMethod.StringVal() + " " + op
		description = "http"
		switch spanKind {
		case pdata.SpanKindClient:
			description += ".client"
		case pdata.SpanKindServer:
			description += ".server"
		}
		description += " " + op // Ex. description="GET /api/users/{user_id}".

		return op, description
	}

	// If db.type exists then this is a database call span.
	if _, ok := attrs.Get(conventions.AttributeDBSystem); ok {
		// Use DB statement (Ex "SELECT * FROM table") if possible as description.
		if statement, okInst := attrs.Get(conventions.AttributeDBStatement); okInst {
			description = statement.StringVal()
		}

		return "db", description
	}

	// If rpc.service exists then this is a rpc call span.
	if _, ok := attrs.Get(conventions.AttributeRPCService); ok {
		description = "rpc"
		switch spanKind {
		case pdata.SpanKindClient:
			description += ".client"
		case pdata.SpanKindServer:
			description += ".server"
		}
		return op, description
	}

	// Default just use span.name.
	return
}

func statusFromSpanStatus(spanStatus pdata.SpanStatus) (status sentry.SpanStatus, message string) {
	code := spanStatus.Code()
	if code < 0 || int(code) >= len(canonicalCodes) {
		return sentry.SpanStatusUnknown, "ErrorCode(" + strconv.Itoa(int(code)) + ")"
	}

	return canonicalCodes[code], spanStatus.Message()
}

func generateEventID() sentry.EventID {
	return sentry.EventID(uuid())
}
