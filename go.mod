module github.com/yeqown/opentelemetry-quake

go 1.18

require (
	github.com/pkg/errors v0.9.1
	github.com/yeqown/opentelemetry-quake/sentryexporter v1.0.0
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/exporters/jaeger v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.1.0
	go.opentelemetry.io/otel/sdk v1.1.0
	go.opentelemetry.io/otel/trace v1.1.0
)

require (
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/getsentry/sentry-go v0.11.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/stretchr/objx v0.1.1 // indirect
	go.opentelemetry.io/collector/model v0.38.0 // indirect
	go.opentelemetry.io/proto/otlp v0.9.0 // indirect
	golang.org/x/net v0.0.0-20210610132358-84b48f89b13b // indirect
	golang.org/x/sys v0.0.0-20210816074244-15123e1e1f71 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20210604141403-392c879c8b08 // indirect
	google.golang.org/grpc v1.41.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/yeqown/opentelemetry-quake/sentryexporter v1.0.0 => ./exporter/sentry
