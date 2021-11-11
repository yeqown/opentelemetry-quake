module github.com/yeqown/opentelemetry-quake/examples

go 1.16

require (
	github.com/gin-gonic/gin v1.7.4
	github.com/go-resty/resty/v2 v2.7.0
	github.com/yeqown/opentelemetry-quake v1.0.0
	github.com/yeqown/opentelemetry-quake/x v1.0.0
)

replace (
	github.com/yeqown/opentelemetry-quake => ../
	github.com/yeqown/opentelemetry-quake/pkg/sentry-exporter => ../pkg/sentry-exporter
	github.com/yeqown/opentelemetry-quake/x => ../x
)
