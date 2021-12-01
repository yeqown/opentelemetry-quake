package main

import (
	"context"

	"github.com/go-resty/resty/v2"

	otelquake "github.com/yeqown/opentelemetry-quake"
	"github.com/yeqown/opentelemetry-quake/x/resty"
)

func main() {
	shutdown := otelquake.MustSetup(
		otelquake.WithServerName("resty-demo"),
		//opentelemetry.WithSentryExporter("https://SECRECT@sentry.example.com/7"),
		otelquake.WithOtlpExporter(""),
		otelquake.WithSampleRate(1.0),
	)
	defer shutdown()

	client := resty.New()
	otelresty.InjectTracing(client)

	// create a root span
	ctx, sp := otelquake.StartSpan(context.Background(), "client")
	defer sp.End()

	resp, err := client.
		R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"name": "resty",
		}).
		Post("http://localhost:8080/greet?name=foo")
	_ = resp
	if err != nil {
		println(err.Error())
		return
	}

	println("trace info:", sp.SpanContext().TraceID().String())
	println("span info:", sp.SpanContext().SpanID().String())
}
