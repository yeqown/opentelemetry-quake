package main

import (
	"context"
	"errors"
	"log"
	"time"

	otelquake "github.com/yeqown/opentelemetry-quake"
	sentryexporter "github.com/yeqown/opentelemetry-quake/exporter/sentry"
	otelgin "github.com/yeqown/opentelemetry-quake/x/gin"

	"github.com/gin-gonic/gin"
)

func main() {
	shutdown, err := otelquake.Setup(
		otelquake.WithSentryExporter("https://SECRECT@sentry.example.com/7"),
		//opentelemetry.WithOtlpExporter(),
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
	r.GET("/ping", func(c *gin.Context) {
		processWithSpan(otelgin.ContextFrom(c))
		c.String(200, "pong")
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
	println("traceId: ", sp.SpanContext().TraceID().String())
	println("sampled: ", sp.SpanContext().TraceFlags().IsSampled())
	defer sp.End()

	// sleep 100ms
	time.Sleep(100 * time.Millisecond)
}
