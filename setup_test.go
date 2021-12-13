package tracing_test

import (
	tracing "github.com/yeqown/opentelemetry-quake"
)

func ExampleSetupDefault() {
	name := "default_or_from_env"
	version := "1.0.0"
	env := "prod"
	ns := "my_namespace"
	otelCollectorEndpoint := "localhost:4317"

	shutdown, err := tracing.Setup(
		tracing.WithServerName(name),
		tracing.WithServerVersion(version),
		tracing.WithEnv(env),
		tracing.WithNamespace(ns),
		tracing.WithOtlpExporter(otelCollectorEndpoint),
		tracing.WithSampleRate(0.2),
	)
	_, _ = shutdown, err
}
