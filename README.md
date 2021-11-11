<div align="center"><img src="./assets/logo.png" width="60%"/></div>

# opentelemetry

[![Go Report Card](https://goreportcard.com/badge/github.com/yeqown/opentelemetry-quake)](https://goreportcard.com/report/github.com/yeqown/opentelemetry-quake) [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/yeqown/opentelemetry-quake)

适用 opentelemetry 规范的链路追踪仓库，开发这个仓库的目的是：简化项目中对接链路追踪的代码，并统一搜集行为。

预期：

- 遵守 opentelemetry 的使用规范
- 能够同时支持 sentry 和 jaeger 后端服务，对于 sentry 同时支持采集panic时的堆栈信息
- 不需要关心 trace provider 的注册和使用
- 没有带来明显的延迟增长
- 简洁的API

### TODO:

- [x] sdk 可用
- [x] 采集堆栈信息
- [x] 跟前端链路打通
- [x] 中间件完成与优化
- [ ] 处理代码中 `TODO` 和 `FIXME`
- [x] 记录请求和响应（可选）
- [ ] 补充测试用例


### Deploy OpenTelemetry Collector

https://opentelemetry.io/docs/collector/getting-started/

***k8s***

```sh
kubectl apply -f ./.deploy/k8s-otelcol.yaml
```

### Supplement

***1. How build you custom collector***

> For more details, please check out the .build folder.

```sh
## install build
go install go.opentelemetry.io/collector/cmd/builder@latest

## write build config
cat > ~/.otelcol-builder.yaml <<EOF
exporters:
  - gomod: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/alibabacloudlogserviceexporter v0.37.0"
EOF

## execute build command

builder --output-path=.
# or builder --config ~/.otelcol-builder.yaml
```

### Reference

- https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder
- https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/6218