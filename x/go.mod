module github.com/yeqown/opentelemetry-quake/x

go 1.16

require (
	github.com/yeqown/opentelemetry-quake v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.7.4
	github.com/go-resty/resty/v2 v2.7.0
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/trace v1.1.0
)

// replace github.com/yeqown/opentelemetry-quake => ../
replace github.com/yeqown/opentelemetry-quake/sentryexporter => ../exporter/sentry
