module github.com/yeqown/opentelemetry-quake/contrib/resty

go 1.15

require (
	github.com/go-resty/resty/v2 v2.7.0
	github.com/yeqown/opentelemetry-quake v1.3.1
	go.opentelemetry.io/otel v1.2.0
)

//replace github.com/yeqown/opentelemetry-quake => ../../
