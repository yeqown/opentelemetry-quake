package main

import (
	"context"

	"github.com/go-resty/resty/v2"

	tracing "github.com/yeqown/opentelemetry-quake"
	tracingresty "github.com/yeqown/opentelemetry-quake/contrib/resty"
)

func main() {
	shutdown := tracing.MustSetup(
		tracing.WithServerName("resty-demo"),
		//opentelemetry.WithSentryExporter("https://SECRECT@sentry.example.com/7"),
		tracing.WithOtlpExporter(""),
		tracing.WithSampleRate(1.0),
	)
	defer shutdown()

	client := resty.New()
	tracingresty.InjectTracing(client)

	// create a root span
	ctx, sp := tracing.StartSpan(context.Background(), "client")
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

	println("trace info:", sp.SpanContext().TraceID)
	println("span info:", sp.SpanContext().SpanID)
}
