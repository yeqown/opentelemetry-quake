package otelresty_test

import (
	"context"

	"github.com/go-resty/resty/v2"

	otelresty "github.com/yeqown/opentelemetry-quake/x/resty"
)

func ExampleInjectTracing() {
	client := resty.New()
	otelresty.InjectTracing(client)

	// pretend this to be a trace context, passed as parameter.
	ctx := context.Background()

	resp, err := client.
		R().
		// !!! make sure you called set context to pass trace context to middleware.
		SetContext(ctx).
		Get("https://example.com/api/resource/4396")

	// handle response and error
	_, _ = resp, err
}
