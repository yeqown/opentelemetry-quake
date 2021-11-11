package opentelemetry

import (
	"context"
	"log"
	"sync"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	sentryexporter "github.com/yeqown/opentelemetry-quake/exporter/sentry"
)

// newExporter returns a console exporterEnum.
func newExporter(so setupOption) (exp trace.SpanExporter, err error) {
	switch so.exporter {
	case SENTRY:
		exp, err = sentryexporter.New(so.sentryDSN)
	case OTLP:

		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			//otlptracegrpc.WithEndpoint(), // TODO(@yeqown): allow parameters, such as host and port
		)
		exp, err = otlptrace.New(context.Background(), client)
	case JAEGER:
		exp, err = jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(so.jaegerAgentHost)))
	default:
		err = errors.New("unknown exporter")
	}

	if err != nil {
		return nil, errors.Wrap(err, "newExporter failed")
	}

	return exp, err
}

// newResource returns a resource describing this application.
// DONE(@yeqown): allow modifying and configured by developer by WithXXX API,
// also try extract from environment variables while some of them are empty.
func newResource(so setupOption) *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(so.serverName),
			semconv.ServiceVersionKey.String(so.version),
			attribute.String("environment", so.env),
		),
	)
	return r
}

var _setupOnce sync.Once

// Setup would only execute once if it is called multiple times. Of course,
// if setup failed , it would return an error and allows the caller to retry.
func Setup(opts ...SetupOption) (shutdown func(), err error) {
	_setupOnce.Do(func() {
		shutdown, err = setup(opts...)
	})

	if err != nil {
		// re-create once avoid setting up once but failed.
		_setupOnce = sync.Once{}
	}

	return shutdown, err
}

func setup(opts ...SetupOption) (func(), error) {
	so := defaultSetupOpt
	for _, o := range opts {
		o.apply(&so)
	}
	if err := fixSetupOption(&so); err != nil {
		return nil, errors.Wrap(err, "setup try to fixSetupOption")
	}

	// DONE(@yeqown): use factory pattern to create exporterEnum. jaeger and sentry are optional.
	exporter, err := newExporter(so)
	if err != nil {
		return nil, errors.Wrap(err, "setup create exporterEnum")
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(newResource(so)),
		trace.WithSampler(trace.TraceIDRatioBased(so.sampleRatio)),
	)
	// generate a shutdown function to close trace provider.
	shutdown := func() {
		if err = provider.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}

	// register tracer provider
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return shutdown, nil
}

// MustSetup same as Setup but panic if setup encounter any error.
func MustSetup(opts ...SetupOption) func() {
	shutdown, err := Setup(opts...)
	if err != nil {
		panic(err)
	}

	return shutdown
}
