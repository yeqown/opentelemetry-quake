package tracing

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.5.0"
)

// newExporter returns a console exporterEnum.
func newExporter(so setupOption) (exp trace.SpanExporter, err error) {
	switch so.exporter {
	//case SENTRY:
	//	exp, err = sentryexporter.New(so.sentryDSN)
	//case JAEGER:
	//	exp, err = jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(so.jaegerAgentHost)))
	case OTLP:
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(so.oltpEndpoint),
		)
		exp, err = otlptrace.New(context.Background(), client)
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
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(so.serverName),
		semconv.ServiceVersionKey.String(so.version),
		semconv.ServiceNamespaceKey.String(so.namespace),
		semconv.HostNameKey.String(so.hostname),
		semconv.DeploymentEnvironmentKey.String(so.env),
		attribute.String("pod.ip", so.podIP),
	)

	//r, _ := resource.Merge(
	//	resource.Default(),
	//	resource.NewWithAttributes(
	//		semconv.SchemaURL,
	//		semconv.ServiceNameKey.String(so.serverName),
	//		semconv.ServiceVersionKey.String(so.version),
	//		semconv.ServiceNamespaceKey.String(so.namespace),
	//		attribute.String("environment", so.env),
	//	),
	//)
	//return r
}

// SetupDefault setup default tracing, using:
//
// OLTP exporter with default endpoint: localhost:4317 or OTEL_COLLECTOR_ENDPOINT;
// serverName from environment variable: APP_ID, APP_NAME;
// version from environment variable: APP_VERSION;
// env from environment variable: RUN_ENV, DEPLOY_ENV;
// namespace from environment variable: NAMESPACE;
// sampleRate set 0.2, means 20% of traces will be sampled, or you can set OTEL_SAMPLE_RATE=[0..1.0];
func SetupDefault() (shutdown func(), err error) {
	defaultOrFromEnv := func(_default string, candidateKeys ...string) (value string) {
		value = _default
		for _, key := range candidateKeys {
			if v := os.Getenv(key); v != "" {
				value = v
				return value
			}
		}

		return value
	}

	name := defaultOrFromEnv("unknown", "APP_ID", "APP_NAME")
	version := defaultOrFromEnv("untagged", "APP_VERSION")
	env := defaultOrFromEnv("default", "RUN_ENV", "DEPLOY_ENV")
	ns := defaultOrFromEnv("app", "NAMESPACE")
	hostname := defaultOrFromEnv("unknown", "HOSTNAME", "POD_NAME")
	podIP := defaultOrFromEnv("127.0.0.1", "POD_IP")

	otelCollectorEndpoint := defaultOrFromEnv("localhost:4317", "OTEL_COLLECTOR_ENDPOINT")
	_fraction := defaultOrFromEnv("0.2", "OTEL_SAMPLE_RATE")
	sampleFraction, err := strconv.ParseFloat(_fraction, 64)
	if err != nil {
		fmt.Printf("[med/opentelemetry] WARNNING: OTEL_SAMPLE_RATE must be a float number, "+
			"parse %s failed: %v\n", _fraction, err)
	}

	return Setup(
		WithServerName(name),
		WithServerVersion(version),
		WithEnv(env),
		WithNamespace(ns),
		WithHostname(hostname),
		WithPodIP(podIP),
		WithOtlpExporter(otelCollectorEndpoint),
		WithSampleRate(sampleFraction),
	)
}

var _setupOnce sync.Once

// Setup would only execute once if it is called multiple times. Of course,
// if setup failed, it would return an error and allows the caller to retry.
// After setup, open telemetry's sdk has been initialized with TracerProvider
// and Propagator across processes.
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
	so := defaultSetupOption()
	for _, o := range opts {
		o.apply(&so)
	}
	if err := fixSetupOption(&so); err != nil {
		return nil, errors.Wrap(err, "setup try to fixSetupOption")
	}

	fmt.Printf("[med/opentelemetry] setup with options: %+v\n", so)
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
	// no need to set this, tracing use custom TraceContextPropagator.
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
