# opentelemetry

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
- [ ] 中间件完成与优化
- [ ] 处理代码中 `TODO` 和 `FIXME`
- [ ] 记录请求和响应（可选）
- [ ] 补充测试用例