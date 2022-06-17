package go2skyx

import (
	"context"
	"testing"
	"time"
)

var (
	addr         = "127.0.0.1:11800"
	samplingRate = 1.0

	// mock cross_processing to pass propagation context
	propagationMap = map[string]string{}
)

func TestGo2skyx(t *testing.T) {
	tracer, cleanup, err := NewTracer(
		// SkyWalking gRPC addr, default 127.0.0.1:11800
		WithAddr(addr),
		// service, default go2skyx
		WithService("go2skyx"),
		// samplingRate, default 1.0
		WithSamplingRate(samplingRate),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		// wait 1s to cleanup
		time.Sleep(1 * time.Second)
		cleanup()
	}()

	_, span, errCreateSpan := tracer.CreateSpan(context.Background(),
		WithInjector(func(headerKey, headerValue string) error {
			propagationMap[headerKey] = headerValue
			return nil
		}),
		WithSpanLayer(SpanLayerRPCFramework),
		WithEndpoint("test go2skyx"),
		WithTraceID("TraceID_"+time.Now().Format(time.RFC3339)),
		// you can add any tag
		WithTag(TagURL, "..."),
	)
	if errCreateSpan != nil {
		panic(errCreateSpan)
	}

	// call server async
	go server("server processing async")
	// call server sync
	server("server processing sync")
	span.End()
}

// server mock cross_processing
func server(endpoint string) context.Context {
	tracer, cleanup, err := NewTracer(
		WithAddr(addr),
		WithService("go2skyx_server_"+endpoint),
		WithSamplingRate(samplingRate),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		// wait 1s to cleanup
		time.Sleep(1 * time.Second)
		cleanup()
	}()

	// 1. create an entry span
	xCtx, spanServer, errServer := tracer.CreateSpan(context.Background(),
		WithExtractor(func(headerKey string) (string, error) {
			return propagationMap[headerKey], nil
		}),
		WithSpanLayer(SpanLayerRPCFramework),
		WithEndpoint(endpoint),
	)
	if errServer != nil {
		panic(errServer)
	}

	// 2. mock redis request
	_, spanRedis, errRedis := tracer.CreateSpan(xCtx,
		WithSpanLayer(SpanLayerCache),
		WithComponent(7),
		WithEndpoint("get cache"),
	)
	if errRedis != nil {
		panic(errRedis)
	}
	// sleep for redis exec
	time.Sleep(1 * time.Second)
	spanRedis.Log("cmd", "get")
	spanRedis.End()

	// 3. mock database request
	_, spanDatabase, errDatabase := tracer.CreateSpan(xCtx,
		WithSpanLayer(SpanLayerDatabase),
		WithComponent(5),
		WithEndpoint("query DB"),
	)
	if errDatabase != nil {
		panic(errDatabase)
	}
	// sleep for database exec
	time.Sleep(1 * time.Second)
	// log something
	spanDatabase.Log("cmd", "query")
	spanDatabase.End()

	// 4. mock local exec
	_, spanLocal, errLocal := tracer.CreateSpan(xCtx,
		WithSpanLayer(SpanLayerRPCFramework),
		WithEndpoint("local invoke"),
	)
	if errLocal != nil {
		panic(errLocal)
	}
	// sleep for local invoke
	time.Sleep(1 * time.Second)
	// log something
	spanLocal.Log("local", "invoke")
	spanLocal.End()

	spanServer.Log("server tracing", "done")
	spanServer.End()
	return xCtx
}
