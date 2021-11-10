package otelgin_test

import (
	"github.com/gin-gonic/gin"

	sentryexporter "github.com/yeqown/opentelemetry-quake/exporter/sentry"
	otelgin "github.com/yeqown/opentelemetry-quake/x/gin"
)

func ExampleTracing() {
	r := gin.Default()

	r.Use(
		otelgin.Tracing(otelgin.DefaultConfig()),
	)

	// sentry trace header
	r.Use(
		otelgin.Tracing(otelgin.Config{
			Factory:     sentryexporter.CarrierFactory,
			LogRequest:  true,
			LogResponse: true,
		}),
	)
}

func ExampleCaptureError() {
	r := gin.Default()

	r.Use(
		otelgin.CaptureError(),
	)
}
