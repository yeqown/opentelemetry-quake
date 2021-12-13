package tracing

import (
	"errors"
	"os"
)

type exporterEnum string

const (
	OTLP exporterEnum = "OTLP"
	//JAEGER exporterEnum = "JAEGER"
	//SENTRY exporterEnum = "SENTRY"
)

type setupOption struct {
	// serverName represents an identity of application.
	serverName string
	version    string
	env        string
	namespace  string
	hostname   string
	podIP      string

	exporter exporterEnum
	//jaegerAgentHost string // jaegerAgentHost is the hostname of the jaeger agent.
	//sentryDSN    string // it could not be empty while exporter is SENTRY.
	oltpEndpoint string // it could not be empty while exporter is OTLP.

	sampleRatio float64 // sampleRatio is the sampling ratio of trace. 1.0 means 100% sampling, 0 means 0% sampling.
}

func defaultSetupOption() setupOption {
	defaultHost := os.Getenv("NODE_IP")
	if defaultHost == "" {
		defaultHost = "127.0.0.1"
	}

	return setupOption{
		serverName: "unknown",
		version:    "v0.0.0",
		env:        "default",
		namespace:  "default",
		hostname:   "unknown",
		podIP:      "127.0.0.1",
		exporter:   OTLP,
		//jaegerAgentHost: "127.0.0.1",
		//sentryDSN:       "",
		// 如果没有指定endpoint，则使用默认的HOST和端口 localhost:4317
		// DONE(@yeqown): 使用 agent 模式部署 otelcol 后，采用 NodeIP:4317 作为默认值
		oltpEndpoint: defaultHost + ":4317",
		sampleRatio:  1.0,
	}
}

var (
	ErrUnknownExporter   = errors.New("unknown exporterEnum type")
	ErrOtlpEndpointEmpty = errors.New("otlp endpoint could not be empty")
	ErrServerNameEmpty   = errors.New("server name could not be empty")
	//ErrJaegerAgentHostEmpty = errors.New("jaeger agent host could not be empty")
)

func fixSetupOption(so *setupOption) error {
	switch so.exporter {
	case OTLP:
	default:
		return ErrUnknownExporter
	}

	//if so.exporter == JAEGER && so.jaegerAgentHost == "" {
	//	return ErrJaegerAgentHostEmpty
	//}

	if so.exporter == OTLP && so.oltpEndpoint == "" {
		return ErrOtlpEndpointEmpty
	}

	if so.serverName == "" {
		return ErrServerNameEmpty
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

func WithHostname(hostname string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.hostname = hostname
	})
}

func WithPodIP(podIP string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		o.podIP = podIP
	})
}

//func WithSentryExporter(dsn string) SetupOption {
//	return fnSetupOption(func(o *setupOption) {
//		o.exporter = SENTRY
//		o.sentryDSN = dsn
//	})
//}

//func WithJaegerExporter(agentHost string) SetupOption {
//	if agentHost == "" {
//		agentHost = os.Getenv("NODE_IP")
//	}
//
//	return fnSetupOption(func(o *setupOption) {
//		o.exporter = JAEGER
//		o.jaegerAgentHost = agentHost
//	})
//}

func WithOtlpExporter(endpoint string) SetupOption {
	return fnSetupOption(func(o *setupOption) {
		if endpoint != "" {
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
