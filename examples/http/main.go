package main

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"

	otelquake "github.com/yeqown/opentelemetry-quake"
	pb "github.com/yeqown/opentelemetry-quake/examples/api"
	sentryexporter "github.com/yeqown/opentelemetry-quake/sentryexporter"
	otelgin "github.com/yeqown/opentelemetry-quake/x/gin"
	otelgrpc "github.com/yeqown/opentelemetry-quake/x/grpc"

	"github.com/gin-gonic/gin"
)

func main() {
	shutdown, err := otelquake.Setup(
		//otelquake.WithSentryExporter("https://SECRECT@sentry.example.com/7"),
		otelquake.WithOtlpExporter(),
		otelquake.WithServerName("demo"),
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

		if err = processWithSpan(otelgin.ContextFrom(c), req.Name); err != nil {
			c.JSON(400, gin.H{
				"error": err,
			})
			return
		}

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

func processWithSpan(ctx context.Context, name string) error {
	// start a span
	//ctx, sp := otelquake.StartSpan(ctx, "processWithSpan")
	//defer sp.End()

	cc, err := grpc.Dial("localhost:8000",
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(otelgrpc.TracingClientInterceptor(otelgrpc.LogPayloads())),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}

	if _, err = pb.NewGreeterClient(cc).SayHello(ctx, &pb.HelloRequest{Name: name}); err != nil {
		return err
	}

	// sleep 100ms
	time.Sleep(100 * time.Millisecond)
	return nil
}
