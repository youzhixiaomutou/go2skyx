package go2skyx

import (
	"context"
	"errors"
	"reflect"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	agentV3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

type Tracer struct {
	tracerInner *go2sky.Tracer

	addr         string
	service      string
	samplingRate float64
}

type tracerOption func(t *Tracer)

// WithAddr set addr
func WithAddr(addr string) tracerOption {
	return func(t *Tracer) {
		t.addr = addr
	}
}

// WithService set service
func WithService(service string) tracerOption {
	return func(t *Tracer) {
		t.service = service
	}
}

// WithSamplingRate set samplingRate
func WithSamplingRate(samplingRate float64) tracerOption {
	return func(t *Tracer) {
		t.samplingRate = samplingRate
	}
}

// NewTracer new tracer
func NewTracer(options ...tracerOption) (tracer *Tracer, cleanup func(), err error) {
	tracer = &Tracer{
		addr:         "127.0.0.1:11800",
		service:      "go2skyx",
		samplingRate: 1.0,
	}

	for _, option := range options {
		option(tracer)
	}

	var r go2sky.Reporter
	r, err = reporter.NewGRPCReporter(tracer.addr)
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() {
		if r != nil {
			r.Close()
		}
	}

	var tracerInner *go2sky.Tracer
	tracerInner, err = go2sky.NewTracer(tracer.service, go2sky.WithReporter(r), go2sky.WithSampler(tracer.samplingRate))
	if err != nil {
		return nil, nil, err
	}
	tracer.tracerInner = tracerInner

	return tracer, cleanup, nil
}

func (t *Tracer) CreateSpan(ctx context.Context, options ...spanOption) (xCtx context.Context, span *Span, err error) {
	span = &Span{
		spanLayer: SpanLayerUnknown,
		component: componentUnknown,
		endpoint:  "no operation name",
		peer:      "No Peer",
		tagMap:    map[Tag]string{},
	}

	for _, option := range options {
		option(span)
	}

	var spanInner go2sky.Span
	// create span
	if span.extractor != nil {
		spanInner, xCtx, err = t.tracerInner.CreateEntrySpan(ctx, span.endpoint, span.extractor)
	} else if span.injector != nil {
		spanInner, xCtx, err = t.tracerInner.CreateExitSpanWithContext(ctx, span.endpoint, span.peer, span.injector)
	} else {
		spanInner, xCtx, err = t.tracerInner.CreateLocalSpan(ctx)
	}
	if err != nil {
		return nil, nil, err
	}

	spanInner.SetSpanLayer(agentV3.SpanLayer(span.spanLayer))
	spanInner.SetComponent(span.component)
	spanInner.SetOperationName(span.endpoint)
	spanInner.SetPeer(span.peer)
	for tag, v := range span.tagMap {
		spanInner.Tag(go2sky.Tag(tag), v)
	}

	// use reflect to customize TraceID
	if len(span.traceID) > 0 {
		ssi := reflect.ValueOf(spanInner).Elem().FieldByName("segmentSpanImpl")
		if ssi.IsValid() {
			ssi.Elem().FieldByName("SegmentContext").FieldByName("TraceID").SetString(span.traceID)
		} else {
			return nil, nil, errors.New("you can only set traceID with root")
		}
	}

	span.spanInner = spanInner
	return xCtx, span, nil
}
