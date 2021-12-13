package main

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"

	tracing "github.com/yeqown/opentelemetry-quake"
	tracinggrpc "github.com/yeqown/opentelemetry-quake/contrib/grpc"
	pb "github.com/yeqown/opentelemetry-quake/examples/api"
)

func main() {
	shutdown, err := tracing.Setup(
		tracing.WithServerName("grpc-demo"),
		tracing.WithOtlpExporter(""),
		tracing.WithSampleRate(1.0),
	)
	defer shutdown()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(tracinggrpc.TracingServerInterceptor(tracinggrpc.LogPayloads())),
		grpc.ConnectionTimeout(10*time.Second),
	)
	pb.RegisterGreeterServer(s, new(greeterServer))

	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}

	println("Listening on port 8000")
	if err = s.Serve(l); err != nil {
		panic(err)
	}
}

type greeterServer struct {
	pb.UnimplementedGreeterServer
}

func (g greeterServer) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	println("SayHello")
	return &pb.HelloReply{Message: "Hello " + request.Name}, nil
}
