module github.com/yeqown/opentelemetry-quake

go 1.15

require (
	github.com/kr/pretty v0.3.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/yeqown/opentelemetry-quake/sentryexporter v1.0.0
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/exporters/jaeger v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.1.0
	go.opentelemetry.io/otel/sdk v1.1.0
	go.opentelemetry.io/otel/trace v1.1.0
	golang.org/x/sys v0.0.0-20210816074244-15123e1e1f71 // indirect
)

replace github.com/yeqown/opentelemetry-quake/sentryexporter => ./exporter/sentry
