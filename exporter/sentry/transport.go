package sentry

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
)

var _ transport = (*sentryTransport)(nil)

type transport interface {
	SendEvents(events []*sentry.Event)
	Configure(options sentry.ClientOptions)
	Flush(ctx context.Context) bool
}

type sentryTransport struct {
	httpTrans *sentry.HTTPTransport
}

func newTransport() *sentryTransport {
	return &sentryTransport{
		httpTrans: sentry.NewHTTPTransport(),
	}
}

func (t *sentryTransport) Flush(ctx context.Context) bool {
	deadline, ok := ctx.Deadline()
	if ok {
		return t.httpTrans.Flush(time.Until(deadline))
	}

	return t.httpTrans.Flush(time.Second)
}

func (t *sentryTransport) Configure(options sentry.ClientOptions) {
	t.httpTrans.Configure(options)
}

func (t *sentryTransport) SendEvents(transactions []*sentry.Event) {
	bufferCounter := 0
	for _, transaction := range transactions {
		// We should flush all events when we send transactions equal to the transport
		// buffer size, so we don't drop transactions.
		if bufferCounter == t.httpTrans.BufferSize {
			t.httpTrans.Flush(time.Second)
			bufferCounter = 0
		}

		t.httpTrans.SendEvent(transaction)
		bufferCounter++
	}
}
