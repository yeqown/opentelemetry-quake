package sentryexporter

import (
	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/collector/model/pdata"
	semconv "go.opentelemetry.io/collector/model/semconv/v1.5.0"
)

// 职责：实现 sentry 的 TraceExporter 并作为插件静态编译到 otelcol 组件中,
// TraceExporter 的职责是将 otelcol 的 trace 数据转换为目标 vendor 的 trace 数据格式，
// 并上报到 vendor。
//
// 转换 open telemetry Span 到 sentry 数据格式, 映射如下:
// TraceData => sentry.Events

// traceToSentryEvents 转换的最外层控制函数
func traceToSentryEvents(td pdata.Traces) []*sentry.Event {
	events := make([]*sentry.Event, 0, td.SpanCount())
	transactionEvMapping := make(map[sentry.SpanID]*sentry.Event)
	spanIdMapping := make(map[sentry.SpanID]sentry.SpanID, td.SpanCount())
	maybeOrphanSpans := make([]*sentry.Span, 0, td.SpanCount()/2)
	markAsOrphanTemporary := func(span *sentry.Span) {
		maybeOrphanSpans = append(maybeOrphanSpans, span)
	}

	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		r := rs.Resource()
		resKeyGen := withPrefixKeyGenerator("resource")
		resourceTags := generateTagsFromAttributes(r.Attributes(), resKeyGen)
		env := extractEnvFromTags(resourceTags, semconv.AttributeDeploymentEnvironment, resKeyGen)

		for j := 0; j < rs.InstrumentationLibrarySpans().Len(); j++ {
			ils := rs.InstrumentationLibrarySpans().At(j)
			lib := ils.InstrumentationLibrary()
			spans := ils.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				spanId := span.SpanID().Bytes()
				// 将 open telemetry span 数据格式转换为 sentry span 数据格式
				sentrySpan := sentrySpanFromOtelSpan(&span, &lib, resourceTags)
				// 检测错误和异常，并组装数据
				spanEvents := span.Events()
				exceptionEv := detectExceptionFromOtelSpan(sentrySpan, &spanEvents)
				if exceptionEv != nil {
					exceptionEv.Environment = env
					events = append(events, exceptionEv)
				}
				if isRootSpan(span) {
					// 如果是 root span, 则转换为 transaction event (sentry 的概念)
					transactionEv := sentryEventFromSentrySpan(sentrySpan)
					transactionEv.Environment = env
					events = append(events, transactionEv)
					transactionEvMapping[spanId] = transactionEv
					spanIdMapping[spanId] = spanId
					continue
				}

				// 如果是 child span, 则需要查找 parentSpanId, 并将 此span 添加到 event.Spans
				if rootSpanId, ok := spanIdMapping[sentrySpan.SpanID]; ok {
					// 如果找到 root span, 那么添加到 event.Spans
					transactionEvMapping[rootSpanId].Spans = append(transactionEvMapping[rootSpanId].Spans, sentrySpan)
					spanIdMapping[sentrySpan.SpanID] = rootSpanId
					continue
				}

				// 没有找到，暂时标记为孤儿span
				markAsOrphanTemporary(sentrySpan)
			}
		}
	}

	if len(events) == 0 {
		// 数据为空，不用继续后续的处理
		return nil
	}

	// 整理 events: 处理孤儿 span
	orphans := mergeOrphanSpans(maybeOrphanSpans, spanIdMapping, transactionEvMapping)
	for i := 0; i < len(orphans); i++ {
		events = append(events, sentryEventFromSentrySpan(orphans[i]))
	}
	debugf("traceToSentryEvents handled %d events, left %d orphan spans", len(events), len(orphans))

	return events
}

// isRootSpan 判断是否是根 span: 没有 parentSpanID 或 Server / Consumer Span 类型
func isRootSpan(span pdata.Span) bool {
	kind := span.Kind()
	return span.ParentSpanID() == pdata.SpanID{} ||
		kind == pdata.SpanKindServer || kind == pdata.SpanKindConsumer
}

// sentrySpanFromOtelSpan
// 1. trace 信息
// 2. tags 标签
func sentrySpanFromOtelSpan(
	span *pdata.Span, library *pdata.InstrumentationLibrary, resourceTags map[string]string) *sentry.Span {

	spanKind := span.Kind()

	op, description := generateSpanOperation(span)
	tags := generateTagsFromAttributes(span.Attributes(), intactKeyGenerator())
	for k, v := range resourceTags {
		tags[k] = v
	}

	tags["library.name"] = library.Name()
	tags["library.version"] = library.Version()

	status, message := statusFromSpanStatus(span.Status())
	if message != "" {
		tags["status.message"] = message
	}
	if spanKind != pdata.SpanKindUnspecified {
		tags["span.kind"] = spanKind.String()
	}

	sentrySpan := &sentry.Span{
		TraceID:      span.TraceID().Bytes(),
		SpanID:       span.SpanID().Bytes(),
		ParentSpanID: sentry.SpanID{},
		Op:           op,
		Description:  description,
		Status:       status,
		Tags:         tags,
		StartTime:    unixNanoToTime(span.StartTimestamp()),
		EndTime:      unixNanoToTime(span.EndTimestamp()),
		Data:         make(map[string]interface{}, 4),
		Sampled:      0, // TODO(@yeqiang): 不设置不会有问题嘛？
	}

	if parentSpanID := span.ParentSpanID(); !parentSpanID.IsEmpty() {
		sentrySpan.ParentSpanID = parentSpanID.Bytes()
	}

	return sentrySpan
}

func sentryEventFromSentrySpan(span *sentry.Span) *sentry.Event {
	sentryEvent := &sentry.Event{
		Contexts:    make(map[string]interface{}),
		Extra:       span.Data,
		Tags:        span.Tags,
		Modules:     nil,
		Breadcrumbs: nil,
		Dist:        "",
		Environment: "",
		EventID:     generateEventID(),
		Fingerprint: nil,
		Level:       "info",
		Message:     "",
		Platform:    "",
		Release:     "",
		Sdk: sentry.SdkInfo{
			Name:    otelSentryExporterName,
			Version: otelSentryExporterVersion,
		},
		ServerName:  span.Tags["server_name"],
		Threads:     nil,
		Timestamp:   span.EndTime,
		Transaction: span.Op,
		User:        sentry.User{},
		Logger:      "",
		Request:     nil,
		Exception:   nil,
		Type:        "transaction",
		StartTime:   span.StartTime,
		Spans:       nil,
	}

	sentryEvent.Contexts["trace"] = sentry.TraceContext{
		TraceID:      span.TraceID,
		SpanID:       span.SpanID,
		ParentSpanID: span.ParentSpanID,
		Op:           span.Op,
		Description:  span.Description,
		Status:       span.Status,
	}

	return sentryEvent
}

func detectExceptionFromOtelSpan(
	span *sentry.Span, events *pdata.SpanEventSlice) (sentryEvent *sentry.Event) {
	debugf("detectExceptionFromOtelSpan called")

	for i := 0; i < events.Len(); i++ {
		event := events.At(i)
		switch event.Name() {
		case "exception":
			sentryEvent = generateExceptionEvent(span, &event)
		case "request", "response":
			debugf("detectExceptionFromOtelSpan detect request or response")
			event.Attributes().Range(func(key string, value pdata.AttributeValue) bool {
				span.Data[event.Name()+"."+key] = value.AsString()
				return true
			})
		}
	}

	return sentryEvent
}

func generateExceptionEvent(span *sentry.Span, event *pdata.SpanEvent) *sentry.Event {
	var exceptionMessage, exceptionType, exceptionStack string
	event.Attributes().Range(func(k string, v pdata.AttributeValue) bool {
		switch k {
		case semconv.AttributeExceptionMessage:
			exceptionMessage = v.StringVal()
		case semconv.AttributeExceptionType:
			exceptionType = v.StringVal()
		case semconv.AttributeExceptionStacktrace:
			exceptionStack = v.StringVal()
		}
		return true
	})

	debugf("exceptionMessage: %s, stack: %s\n", exceptionMessage, exceptionStack)
	if exceptionMessage == "" && exceptionType == "" {
		// `At least one of the following sets of attributes is required:
		// - exception.type
		// - exception.message`
		debugf("exceptionMessage and exceptionType are empty, skip this event\n")
		return nil
	}

	sentryEvent := &sentry.Event{
		Contexts:    make(map[string]interface{}),
		Extra:       span.Data,
		Tags:        span.Tags,
		Modules:     nil,
		Breadcrumbs: nil,
		Dist:        "",
		Environment: "",
		EventID:     generateEventID(),
		Fingerprint: nil,
		Level:       "error",
		Message:     exceptionStack,
		Platform:    "",
		Release:     "",
		Sdk: sentry.SdkInfo{
			Name:    otelSentryExporterName,
			Version: otelSentryExporterVersion,
		},
		ServerName:  span.Tags["server_name"],
		Threads:     nil,
		Timestamp:   span.EndTime,
		Transaction: span.Op,
		User:        sentry.User{},
		Logger:      "",
		Request:     nil,
		Exception: []sentry.Exception{
			{Value: exceptionMessage, Type: exceptionType},
		},
		Type:      exceptionType,
		StartTime: span.StartTime,
		Spans:     nil,
	}

	sentryEvent.Contexts["trace"] = sentry.TraceContext{
		TraceID:      span.TraceID,
		SpanID:       span.SpanID,
		ParentSpanID: span.ParentSpanID,
		Op:           span.Op,
		Description:  span.Description,
		Status:       span.Status,
	}

	return sentryEvent
}

// mergeOrphanSpans 孤儿节点合并
func mergeOrphanSpans(
	orphanSpans []*sentry.Span,
	spanIdMapping map[sentry.SpanID]sentry.SpanID,
	txEvMapping map[sentry.SpanID]*sentry.Event,
) []*sentry.Span {
	if len(orphanSpans) == 0 {
		// 全部合并完毕
		return orphanSpans
	}

	left := make([]*sentry.Span, 0, len(orphanSpans))
	for i := 0; i < len(orphanSpans); i++ {
		spanId := orphanSpans[i].SpanID
		if rootSpanID, ok := spanIdMapping[spanId]; ok {
			spanIdMapping[spanId] = rootSpanID
			txEvMapping[rootSpanID].Spans = append(txEvMapping[rootSpanID].Spans, orphanSpans[i])
			continue
		}

		left = append(left, orphanSpans[i])
	}
	if len(left) == len(orphanSpans) {
		// 经过处理后，无法再继续合并
		return orphanSpans
	}

	return mergeOrphanSpans(left, spanIdMapping, txEvMapping)
}

func extractEnvFromTags(resourceTags map[string]string, tagName string, kengen KeyGenerator) string {
	if v, ok := resourceTags[kengen.generate(tagName)]; ok {
		return v
	}

	return "unknown"
}

func generateTagsFromAttributes(attrs pdata.AttributeMap, keygen KeyGenerator) map[string]string {
	tags := make(map[string]string, attrs.Len())

	debugf("generateTagsFromAttributes called")

	attrs.Range(func(key string, attr pdata.AttributeValue) bool {
		tags[keygen.generate(key)] = attr.AsString()
		return true
	})

	return tags
}
