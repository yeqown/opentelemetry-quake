package main

import (
	context "context"
	"net"
	"time"

	"google.golang.org/grpc"

	otelquake "github.com/yeqown/opentelemetry-quake"
	pb "github.com/yeqown/opentelemetry-quake/examples/api"
	otelgrpc "github.com/yeqown/opentelemetry-quake/x/grpc"
)

func main() {
	shutdown, err := otelquake.Setup(
		otelquake.WithServerName("grpc-demo"),
		otelquake.WithOtlpExporter(""),
		otelquake.WithSampleRate(1.0),
	)
	defer shutdown()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(otelgrpc.TracingServerInterceptor(otelgrpc.LogPayloads())),
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
