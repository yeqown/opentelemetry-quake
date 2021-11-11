package otelquake

import (
	"errors"
	"os"
)

type exporterEnum string

const (
	JAEGER exporterEnum = "JAEGER"
	SENTRY exporterEnum = "SENTRY"
	OTLP   exporterEnum = "OTLP"
)

type setupOption struct {
	// serverName represents an identity of application.
	serverName string
	version    string
	env        string

	exporter exporterEnum

	jaegerAgentHost string // jaegerAgentHost is the hostname of the jaeger agent.
	sentryDSN       string // it could not be empty while useSentry is true.

	sampleRatio float64 // sampleRatio is the sampling ratio of trace. 1.0 means 100% sampling, 0 means 0% sampling.
}

var (
	defaultSetupOpt = setupOption{
		serverName:      "",
		version:         "v0.0.0",
		sentryDSN:       "",
		env:             "",
		exporter:        JAEGER,
		jaegerAgentHost: "127.0.0.1",
	}

	ErrUnknownExporter      = errors.New("unknown exporterEnum type")
	ErrSentryDSNEmpty       = errors.New("sentry DSN could not be empty")
	ErrJaegerAgentHostEmpty = errors.New("jaeger agent host could not be empty")
	ErrServerNameEmpty      = errors.New("server name could not be empty")
)

func fixSetupOption(so *setupOption) error {
	switch so.exporter {
	case JAEGER:
	case SENTRY:
	case OTLP:
	default:
		return ErrUnknownExporter
	}

	if so.exporter == SENTRY && so.sentryDSN == "" {
		return ErrSentryDSNEmpty
	}

	if so.exporter == JAEGER && so.jaegerAgentHost == "" {
		return ErrJaegerAgentHostEmpty
	}

	if so.serverName == "" {
		so.serverName = os.Getenv("APP_NAME")
	}
	if so.serverName == "" {
		return ErrServerNameEmpty
	}

	if so.env == "" {
		so.env = os.Getenv("ENV")
	}

	return nil
}

type SetupOption interface {
	apply(*setupOption)
}

type fnSetupOption func(*setupOption)

func (f fnSetupOption) apply(o *setupOption) {
	f(o)
}

func WithServerName(name string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.serverName = name
	})
}

func WithSentryExporter(dsn string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.exporter = SENTRY
		o.sentryDSN = dsn
	})
}

func WithJaegerExporter(agentHost string) SetupOption {
	if agentHost == "" {
		agentHost = os.Getenv("NODE_IP")
	}

	return fnSetupOption(func(o *setupOption) {
		o.exporter = JAEGER
		o.jaegerAgentHost = agentHost
	})
}

func WithOtlpExporter() SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.exporter = OTLP
	})
}

func WithSampleRate(fraction float64) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.sampleRatio = fraction
	})
}
