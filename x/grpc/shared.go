package otelgrpc

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/yeqown/opentelemetry-quake/pkg"
)

const (
	tracerName = "opentelemetry/x/grpc"
)

var (
	_ propagation.TextMapCarrier = (*metadataCarrier)(nil)
	// 将 pb.Message 处理为 JSON string 而不是使用默认的编码
	jsonMarshallar = protojson.MarshalOptions{
		EmitUnpopulated: true, // 打印零值
	}
)

// metadataCarrier satisfies both the opentracing.TextMapReader and
// opentracing.TextMapWriter interfaces.
type metadataCarrier struct {
	metadata.MD
}

func (w metadataCarrier) Get(key string) string {
	values := w.MD.Get(key)
	if len(values) == 0 {
		return ""
	}

	// NOTICE: this function can only be used to implement TextMapCarrier
	return values[0]
}

func (w metadataCarrier) Keys() []string {
	keys := make([]string, 0, w.MD.Len())
	for k := range w.MD {
		keys = append(keys, k)
	}
	return keys
}

func (w metadataCarrier) Set(key, val string) {
	// The GRPC HPACK implementation rejects any uppercase keys here.
	//
	// As such, since the HTTP_HEADERS format is case-insensitive anyway, we
	// blindly lowercase the key (which is guaranteed to work in the
	// Inject/Extract sense per the OpenTracing spec).
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

//func (w metadataCarrier) ForeachKey(handler func(key, val string) error) error {
//	for k, vals := range w.MD {
//		for _, v := range vals {
//			if err := handler(k, v); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

func extractSpanContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	tc := propagation.TraceContext{}
	ctxWithSpan := tc.Extract(ctx, metadataCarrier{MD: md})
	//sc := trace.SpanContextFromContext(ctxWithSpan)

	return ctxWithSpan
}

func injectSpanContext(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	tc := propagation.TraceContext{}
	tc.Inject(ctx, metadataCarrier{MD: md})
	return metadata.NewOutgoingContext(ctx, md)
}

func marshalPbMessage(v interface{}) string {
	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Sprintf("%s", v)
	}

	bytes, err := jsonMarshallar.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%s", v)
	}

	return pkg.ToString(bytes)
}
