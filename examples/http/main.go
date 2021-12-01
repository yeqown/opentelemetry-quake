package main

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/gin-gonic/gin"

	otelquake "github.com/yeqown/opentelemetry-quake"
	pb "github.com/yeqown/opentelemetry-quake/examples/api"
	sentryexporter "github.com/yeqown/opentelemetry-quake/sentryexporter"
	otelgin "github.com/yeqown/opentelemetry-quake/x/gin"
	otelgrpc "github.com/yeqown/opentelemetry-quake/x/grpc"
)

func main() {
	shutdown, err := otelquake.Setup(
		//otelquake.WithSentryExporter("https://SECRECT@sentry.example.com/7"),
		otelquake.WithOtlpExporter(""),
		otelquake.WithServerName("http-demo"),
		otelquake.WithSampleRate(1.0),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown()

	r := gin.Default()
	r.Use(
		otelgin.Tracing(otelgin.DefaultConfig().
			ApplyCarrierFactory(sentryexporter.CarrierFactory).
			EnableLogPayloads(),
		),
		otelgin.CaptureError(),
	)

	cc, err2 := grpc.Dial("localhost:8000",
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(otelgrpc.TracingClientInterceptor(otelgrpc.LogPayloads())),
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
		ctx := otelgin.ContextFrom(c)
		if _, err = client.SayHello(ctx, &pb.HelloRequest{Name: req.Name}); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}

		// internal process, brother span
		processWithSpan(otelgin.ContextFrom(c))

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
	ctx, sp := otelquake.StartSpan(ctx, "processWithSpan")
	defer sp.End()

	// sleep 100ms
	time.Sleep(10 * time.Millisecond)
}
