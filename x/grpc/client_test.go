package otelgrpc_test

import (
	"google.golang.org/grpc"

	otelgrpc "github.com/yeqown/opentelemetry-quake/x/grpc"
)

func ExampleTracingClientInterceptor() {
	address := "dns:///xxx"

	conn, err := grpc.Dial(
		address,
		grpc.WithUnaryInterceptor(otelgrpc.TracingClientInterceptor(otelgrpc.LogPayloads())),
	)

	_, _ = conn, err
}
