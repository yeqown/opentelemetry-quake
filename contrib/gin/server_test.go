package tracinggin_test

import (
	"github.com/gin-gonic/gin"

	tracinggin2 "github.com/yeqown/opentelemetry-quake/contrib/gin"
)

func ExampleTracing() {
	r := gin.Default()

	r.Use(
		tracinggin2.Tracing(),
	)

	// sentry trace header
	r.Use(tracinggin2.Tracing(
		tracinggin2.WithCarrierFactory(tracinggin2.SentryCarrierAdaptor),
		tracinggin2.WithRecordPayloads(),
	))
}

func ExampleCaptureException() {
	r := gin.Default()

	r.Use(
		tracinggin2.CaptureException(false),
	)
}
