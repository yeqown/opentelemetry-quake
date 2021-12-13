package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	tracing "github.com/yeqown/opentelemetry-quake"
	tracinggin "github.com/yeqown/opentelemetry-quake/contrib/gin"
	tracinggrpc "github.com/yeqown/opentelemetry-quake/contrib/grpc"
	pb "github.com/yeqown/opentelemetry-quake/examples/api"
)

func main() {
	shutdown, err := tracing.Setup(
		//tracing.WithSentryExporter("https://SECRECT@sentry.example.com/7"),
		tracing.WithOtlpExporter(""),
		tracing.WithServerName("http-demo"),
		tracing.WithSampleRate(1.0),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown()

	r := gin.Default()
	r.Use(
		tracinggin.Tracing(
			tracinggin.WithCarrierFactory(tracinggin.SentryCarrierAdaptor),
			tracinggin.WithRecordPayloads(),
		),
		tracinggin.CaptureException(true),
	)

	cc, err2 := grpc.Dial("localhost:8000",
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(tracinggrpc.TracingClientInterceptor(tracinggrpc.LogPayloads())),
		grpc.WithBlock(),
	)
	if err2 != nil {
		panic(err)
	}
	defer cc.Close()
	client := pb.NewGreeterClient(cc)

	r.POST("/greet", func(c *gin.Context) {
		req := new(struct {
			Name string `json:"name"`
		})

		if err = c.Bind(req); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		// remote process call
		ctx := tracinggin.TracingContextFrom(c)
		if _, err = client.SayHello(ctx, &pb.HelloRequest{Name: req.Name}); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}

		// internal process, brother span
		processWithSpan(ctx)

		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/panic", func(c *gin.Context) {
		panic(errors.New("panic"))
	})

	log.Println("Listening on: http://127.0.0.1:8080")
	if err = r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func processWithSpan(ctx context.Context) {
	// start a span
	ctx, sp := tracing.StartSpan(ctx, "processWithSpan")
	defer sp.End()

	// sleep 100ms
	time.Sleep(10 * time.Millisecond)
}
