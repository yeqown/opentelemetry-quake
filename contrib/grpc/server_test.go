package tracinggrpc_test

import (
	"google.golang.org/grpc"

	tracinggrpc "github.com/yeqown/opentelemetry-quake/contrib/grpc"
)

func ExampleTracingServerInterceptor() {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(tracinggrpc.TracingServerInterceptor(
			tracinggrpc.LogPayloads(),
		)),
	)

	_ = s
}
