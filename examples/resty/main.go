package main

import (
	"context"

	"github.com/go-resty/resty/v2"

	otelquake "github.com/yeqown/opentelemetry-quake"
	"github.com/yeqown/opentelemetry-quake/x/resty"
)

func main() {
	shutdown := otelquake.MustSetup(
		otelquake.WithServerName("resty"),
		//opentelemetry.WithSentryExporter("https://SECRECT@sentry.example.com/7"),
		otelquake.WithOtlpExporter(),
		otelquake.WithSampleRate(1.0),
	)
	defer shutdown()

	client := resty.New()
	otelresty.InjectTracing(client)

	// create a root span
	ctx, sp := otelquake.StartSpan(context.Background(), "resty-main")
	defer sp.End()

	resp, err := client.
		R().
		SetContext(ctx).
		Get("http://localhost:8080/ping")
	_ = resp
	if err != nil {
		println(err.Error())
		return
	}

	println("trace info:", sp.SpanContext().TraceID().String())
	println("span info:", sp.SpanContext().SpanID().String())
}
