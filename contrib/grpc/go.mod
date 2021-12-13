module github.com/yeqown/opentelemetry-quake/contrib/grpc

go 1.15

require (
	github.com/yeqown/opentelemetry-quake v1.3.1
	go.opentelemetry.io/collector/model v0.40.0
	go.opentelemetry.io/otel v1.2.0
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
)

//replace github.com/yeqown/opentelemetry-quake => ../../
