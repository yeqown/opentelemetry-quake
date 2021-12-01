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
	namespace  string

	exporter exporterEnum

	jaegerAgentHost string // jaegerAgentHost is the hostname of the jaeger agent.
	sentryDSN       string // it could not be empty while exporter is SENTRY.
	oltpEndpoint    string // it could not be empty while exporter is OTLP.

	sampleRatio float64 // sampleRatio is the sampling ratio of trace. 1.0 means 100% sampling, 0 means 0% sampling.
}

var (
	defaultSetupOpt = setupOption{
		serverName:      "unknown",
		version:         "v0.0.0",
		env:             "default",
		exporter:        JAEGER,
		jaegerAgentHost: "127.0.0.1",
		sentryDSN:       "",
		oltpEndpoint:    "localhost:4317",
		sampleRatio:     .2,
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

func WithServerVersion(version string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.version = version
	})
}

func WithEnv(env string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.env = env
	})
}

func WithNamespace(namespace string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.namespace = namespace
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

func WithOtlpExporter(endpoint string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		if endpoint != "" {
			// 如果没有指定endpoint，则使用默认的HOST和端口 localhost:4317
			// TODO(@yeqown): 使用 agent 模式部署 otelcol 后，采用 NodeIP:4317 作为默认值
			o.oltpEndpoint = endpoint
		}
		o.exporter = OTLP
	})
}

func WithSampleRate(fraction float64) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.sampleRatio = fraction
	})
}
