## Sentry Exporter

支持的选项：

- `dsn`: DSN告诉出口商将事件发送到哪里。你可以在哨兵项目的“项目设置”中的“客户端密钥”部分找到。
- `insecure_skip_verify`: 如果将其设置为true，则不会检查ssl证书。用于测试目的，以及部署在私有云中的Sentry安装。

范例如下：

```yaml
exporters:
  sentry:
    dsn: https://key@host/path/42
    insecure_skip_verify: true
```

> See the [docs](./docs/transformation.md) for more details on how this transformation is working.

### 已知的限制

sentry 中的链路展示和opentracing/opentelemetry的设计并不一致，因此在exporter中会根据需要转换一些数据。
尤其是对于事务（sentry中的概念）的拆分上，可能会导致一个具有大量(500+)跨度(span)，且只有一个根跨度的事务，可能被
拆分为大量事务。

### 如何采集错误和异常？

范例

```go
// TODO
```

### 如何编译和部署 opentelemetry-collector

**_编译:_**
```sh
cd ./build && bash ./build.sh
# 执行之后，你会在 .build 目录下看到 otelcol 可执行文件
```

**_部署:_**

1. 直接执行
```sh
cp ./build/otelcol ./deploy
# then
cd ./deploy && ./otelcol -config=otelcol-config.yaml
```

2. k8s 部署 
   1. 那么需要使用 **_.build_** 文件夹下的 **_Dockerfile_** 打包
   2. 然后将镜像上传到 **_k8s_** 集群可访问的仓库
   3. 使用 _**.deploy**_ 中的yaml配置进行部署 

### 参考

- [Sentry Docs](https://docs.sentry.io/docs/)
- [Opentelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector-contrib)
- [My Practice](https://github.com/yeqown/opentelemetry-quake)