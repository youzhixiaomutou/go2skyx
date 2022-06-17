# go2skyx

`go2skyx` 是一个基于 [GO2Sky](https://github.com/SkyAPM/go2sky) 通过 `gRPC` 来上报链路信息给 [SkyWalking](https://github.com/apache/skywalking) 的工具库。

## 目的

在 `Go` 语言中通过 [GO2Sky](https://github.com/SkyAPM/go2sky) 来使用 `SkyWalking` 是非常值得推荐的。但是它对于没有深入理解 `SkyWalking` 的各种概念的开发者来说，使用起来是有一定难度的。并且由于高度封装，对 `GO2Sky` 作基于业务的拓展来说是非常困难的。所以，`go2skyx` 封装了一些简单、易于使用的 api 并且提供了一些新的特性（例如，自定义链路 ID）。

## 特性

- 通过 `gRPC` 来上报链路信息给 `SkyWalking`
- 通过函数选项来创建并设置 `tracer` 和 `span`
- 自定义 **`traceID`**, `spanLayer`, `component`, `endpoint`, `peer`, `tags` 和其它任意信息

## 使用方法

你可以使用默认配置来创建一个 `tracer`:

```go
tracer, cleanup, err := NewTracer()
```

然后使用 `tracer` 来报告一个 `span`:

```go
xCtx, span, errCreateSpan := tracer.CreateSpan(context.Background())
span.Log("key", "value")
span.End()
```

通过这两个方法，就可以完成所有功能。但是不幸的是，绝大多数时候都使用默认配置肯定是不能符合你的业务需求的，所以你必须通过传递可选的参数来定制化你的链路信息。

`tracer`:

```go
tracer, cleanup, err := NewTracer(
    // SkyWalking gRPC addr, default 127.0.0.1:11800
    WithAddr(addr),
    // service, default go2skyx
    WithService("go2skyx"),
    // samplingRate, default 1.0
    WithSamplingRate(samplingRate),
)
```

本地调用 `span`:

```go
xCtx, spanLocal, errLocal := tracer.CreateSpan(context.Background(),
    WithSpanLayer(SpanLayerRPCFramework),
    WithEndpoint("local invoke"),
)
```

客户端 `span`:

```go
xCtx, span, errCreateSpan := tracer.CreateSpan(context.Background(),
    WithInjector(func(headerKey, headerValue string) error {
        // inject function for propagation with cross_processing
        return nil
    }),
    WithSpanLayer(SpanLayerRPCFramework),
    WithEndpoint("test go2skyx"),
    WithTraceID("TraceID_"+time.Now().Format(time.RFC3339)),
    // you can add any tag
    WithTag(TagURL, "..."),
}
```

服务端 `span`:
```go
xCtx, spanServer, errServer := tracer.CreateSpan(context.Background(),
    WithExtractor(func(headerKey string) (string, error) {
        // extract function for propagation with cross_processing
        return propagationMap[headerKey], nil
    }),
    WithSpanLayer(SpanLayerRPCFramework),
    WithEndpoint(endpoint),
)
```

我们在 [go2skyx_test.go](#) 提供了一个模拟完整链路调用过程的使用示例。运行它之后，你就可以 `SkyWalking Web UI` 界面中看到类似下面的输出：

![SkyWalkingUI.png](SkyWalkingUI.png)

## 待办

- 文档补全
- 增加静态代码检查

## 参考

- [Apache SkyWalking](https://github.com/apache/skywalking)
- [GO2Sky](https://github.com/SkyAPM/go2sky)
- [go2sky-plugins](https://github.com/SkyAPM/go2sky-plugins)
