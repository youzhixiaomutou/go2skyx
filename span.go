package go2skyx

import (
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	agentV3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

type Span struct {
	spanInner go2sky.Span

	extractor propagation.Extractor
	injector  propagation.Injector

	traceID   string
	spanLayer SpanLayer
	component int32
	endpoint  string
	peer      string
	tagMap    map[Tag]string
}

type spanOption func(s *Span)

// WithTraceID set traceID
func WithTraceID(traceID string) spanOption {
	return func(s *Span) {
		s.traceID = traceID
	}
}

// WithSpanLayer set span layer
func WithSpanLayer(spanLayer SpanLayer) spanOption {
	return func(s *Span) {
		s.spanLayer = spanLayer
	}
}

// WithExtractor set extractor
func WithExtractor(extractor func(headerKey string) (string, error)) spanOption {
	return func(s *Span) {
		s.extractor = extractor
	}
}

// WithInjector set injector
func WithInjector(injector func(headerKey, headerValue string) error) spanOption {
	return func(s *Span) {
		s.injector = func(headerKey, headerValue string) error {
			if headerKey == propagation.Header {
				sc := propagation.SpanContext{}
				err := sc.DecodeSW8(headerValue)
				if err != nil {
					return err
				}
				sc.TraceID = s.traceID
				headerValue = sc.EncodeSW8()
			}
			return injector(headerKey, headerValue)
		}
	}
}

// WithComponent set component
func WithComponent(component int32) spanOption {
	return func(s *Span) {
		s.component = component
	}
}

// WithEndpoint set endpoint
func WithEndpoint(endpoint string) spanOption {
	return func(s *Span) {
		s.endpoint = endpoint
	}
}

// WithPeer set peer
func WithPeer(peer string) spanOption {
	return func(s *Span) {
		s.peer = peer
	}
}

// WithTag set tag
func WithTag(tag Tag, value string) spanOption {
	return func(s *Span) {
		s.tagMap[tag] = value
	}
}

func (span *Span) Log(s ...string) {
	span.spanInner.Log(time.Now(), s...)
}

func (span *Span) Error(s ...string) {
	span.spanInner.Error(time.Now(), s...)
}

func (span *Span) End() {
	span.spanInner.End()
}

// SpanLayer Map to the layer of span
// see https://github.com/apache/skywalking-goapi/blob/main/collect/language/agent/v3/Tracing.pb.go
type SpanLayer agentV3.SpanLayer

const (
	// SpanLayerUnknown Unknown layer. Could be anything.
	SpanLayerUnknown SpanLayer = 0
	// SpanLayerDatabase A database layer, used in tracing the database client component.
	SpanLayerDatabase SpanLayer = 1
	// SpanLayerRPCFramework A RPC layer, used in both client and server sides of RPC component.
	SpanLayerRPCFramework SpanLayer = 2
	// SpanLayerHttp HTTP is a more specific RPCFramework.
	SpanLayerHttp SpanLayer = 3
	// SpanLayerMQ A MQ layer, used in both producer and consumer sides of the MQ component.
	SpanLayerMQ SpanLayer = 4
	// SpanLayerCache A cache layer, used in tracing the cache client component.
	SpanLayerCache SpanLayer = 5
	// SpanLayerFAAS A FAASlayer, used in function-as-a-Service platform.
	SpanLayerFAAS SpanLayer = 6
)

// component of span
// see https://github.com/apache/skywalking/blob/master/oap-server/server-starter/src/main/resources/component-libraries.yml
// see https://github.com/apache/skywalking/blob/master/docs/en/guides/Component-library-settings.md
const componentUnknown = 0

// Tag are supported by sky-walking engine.
// As default, all Tags will be stored, but these have particular meanings.
// see https://github.com/SkyAPM/go2sky/blob/master/span.go
type Tag go2sky.Tag

const (
	TagURL             Tag = "url"
	TagStatusCode      Tag = "status_code"
	TagHTTPMethod      Tag = "http.method"
	TagDBType          Tag = "db.type"
	TagDBInstance      Tag = "db.instance"
	TagDBStatement     Tag = "db.statement"
	TagDBSqlParameters Tag = "db.sql.parameters"
	TagMQQueue         Tag = "mq.queue"
	TagMQBroker        Tag = "mq.broker"
	TagMQTopic         Tag = "mq.topic"
)
