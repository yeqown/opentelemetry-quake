package tracinggrpc

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/runtime/protoimpl"

	tracing "github.com/yeqown/opentelemetry-quake"
	"github.com/yeqown/opentelemetry-quake/pkg"
)

var (
	_ tracing.TraceContextCarrier = (*metadataCarrier)(nil)
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

func (w metadataCarrier) Set(key, val string) {
	// The GRPC HPACK implementation rejects any uppercase keys here.
	//
	// As such, since the HTTP_HEADERS format is case-insensitive anyway, we
	// blindly lowercase the key (which is guaranteed to work in the
	// Inject/Extract sense per the OpenTracing spec).
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func extractSpanContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	return tracing.GetPropagator().Extract(ctx, metadataCarrier{MD: md})
}

func injectSpanContext(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	tracing.GetPropagator().Inject(ctx, metadataCarrier{MD: md})
	return metadata.NewOutgoingContext(ctx, md)
}

func marshalPbMessage(v interface{}) string {
	switch v.(type) {
	case protoiface.MessageV1:
		if bytes, err := jsonMarshallar.Marshal(
			protoimpl.X.ProtoMessageV2Of(v.(protoiface.MessageV1)), // convert to proto.Message
		); err == nil {
			return pkg.ToString(bytes)
		}

	case proto.Message:
		if bytes, err := jsonMarshallar.Marshal(v.(proto.Message)); err == nil {
			return pkg.ToString(bytes)
		}
	}

	return fmt.Sprintf("%s", v)

}
