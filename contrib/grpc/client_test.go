package tracinggrpc_test

import (
	"google.golang.org/grpc"

	tracinggrpc "github.com/yeqown/opentelemetry-quake/contrib/grpc"
)

func ExampleTracingClientInterceptor() {
	address := "dns:///xxx"

	conn, err := grpc.Dial(
		address,
		grpc.WithUnaryInterceptor(tracinggrpc.TracingClientInterceptor(tracinggrpc.LogPayloads())),
	)

	_, _ = conn, err
}
