package otelgrpc_test

import (
	"google.golang.org/grpc"

	otelgrpc "github.com/yeqown/opentelemetry-quake/x/grpc"
)

func ExampleTracingServerInterceptor() {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.TracingServerInterceptor(
			otelgrpc.LogPayloads(),
		)),
	)

	_ = s
}
