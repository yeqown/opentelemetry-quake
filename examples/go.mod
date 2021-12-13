module github.com/yeqown/opentelemetry-quake/examples

go 1.16

require (
	github.com/gin-gonic/gin v1.7.4
	github.com/go-resty/resty/v2 v2.7.0
	github.com/yeqown/opentelemetry-quake v1.2.0
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)

replace (
	github.com/yeqown/opentelemetry-quake => ../
)
