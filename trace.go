package droitrace

import (
	"github.com/DroiTaipei/droictx"
	_ "github.com/DroiTaipei/jaeger-client-go"
	jaegercfg "github.com/DroiTaipei/jaeger-client-go/config"
	jaegerlog "github.com/DroiTaipei/jaeger-client-go/log"
	opentracing "github.com/DroiTaipei/opentracing-go"
	ext "github.com/DroiTaipei/opentracing-go/ext"
	zipkin "github.com/DroiTaipei/zipkin-go-opentracing"
	metrics "github.com/uber/jaeger-lib/metrics"
	"net"
	"net/http"
	"strconv"
)

const (
	ParentSpan = "parentSpan"
)

type SpanReference string

func InitJaeger(componentName string, samplerConf *jaegercfg.SamplerConfig, reporterConf *jaegercfg.ReporterConfig) error {
	cfg := jaegercfg.Configuration{
		Sampler:  samplerConf,
		Reporter: reporterConf,
	}
	// TO-DO: Add droi logger
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	tracer, _, err := cfg.New(componentName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.ZipkinSharedRPCSpan(true),
	)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}

func InitZipkin(collector zipkin.Collector, sampler zipkin.Sampler, host, componentName string) error {
	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, host, componentName),
		zipkin.ClientServerSameSpan(true),
		zipkin.TraceID128Bit(true),
		zipkin.WithSampler(sampler),
	)
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}

func ExtractSpanByTagsMap(spanName string, tags *TagsMap) opentracing.Span {
	var sp opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(tags.Header))
	if err != nil {
		sp = opentracing.StartSpan(spanName)
	} else {
		sp = opentracing.StartSpan(
			spanName,
			ext.RPCServerOption(wireContext))
	}
	attachSpanTags(sp, tags)
	return sp
}

func ExtractSpanFromReq(spanName string, req *http.Request) opentracing.Span {
	var sp opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		sp = opentracing.StartSpan(spanName)
	} else {
		sp = opentracing.StartSpan(
			spanName,
			ext.RPCServerOption(wireContext))
	}
	tags := &TagsMap{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}
	attachSpanTags(sp, tags)
	return sp
}

func CreateSpanFromReq(spaneName string, parentSpan opentracing.Span, relation SpanReference, req *http.Request) opentracing.Span {
	tags := &TagsMap{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
	}
	return CreateSpan(spaneName, parentSpan, relation, tags)
}

func CreateSpanByContext(spanName string, ctx droictx.Context, relation SpanReference, tags *TagsMap) opentracing.Span {
	if ctx == nil {
		return CreateSpan(spanName, nil, ReferenceRoot, tags)
	}
	parentSpanTmp := ctx.Get(ParentSpan)
	if parentSpanTmp == nil {
		return CreateSpan(spanName, nil, ReferenceRoot, tags)
	}
	parentSpan, ok := parentSpanTmp.(opentracing.Span)
	if !ok {
		return CreateSpan(spanName, nil, ReferenceRoot, tags)
	}
	return CreateSpan(spanName, parentSpan, relation, tags)
}

func CreateSpan(spanName string, parentSpan opentracing.Span, relation SpanReference, tags *TagsMap) opentracing.Span {
	var sp opentracing.Span

	switch relation {
	case ReferenceRoot:
		sp = opentracing.StartSpan(spanName)
		attachSpanTags(sp, tags)
	case ReferenceChildOf:
		sp = opentracing.StartSpan(
			spanName,
			opentracing.ChildOf(parentSpan.Context()))
		attachSpanTags(sp, tags)
	case ReferenceFollowsFrom:
		sp = opentracing.StartSpan(
			spanName,
			opentracing.FollowsFrom(parentSpan.Context()))
		attachSpanTags(sp, tags)
	}
	return sp
}

func attachSpanTags(sp opentracing.Span, tags *TagsMap) {
	ext.HTTPMethod.Set(sp, tags.Method)
	if tags.URL != nil {
		ext.HTTPUrl.Set(sp, tags.URL.String())
		if host, portString, err := net.SplitHostPort(tags.URL.Host); err == nil {
			ext.PeerHostname.Set(sp, host)
			if port, err := strconv.Atoi(portString); err != nil {
				ext.PeerPort.Set(sp, uint16(port))
			}
		} else {
			ext.PeerHostname.Set(sp, tags.URL.Host)
		}
	}
	if tags.Header != nil {
		SetDroiTagFromHeaders(sp, tags.Header)
	}
	if tags.Others != nil {
		for k, v := range tags.Others {
			sp.SetTag(k, v)
		}
	}
	return
}

func InjectSpan(sp opentracing.Span, header http.Header) error {
	if err := sp.Tracer().Inject(sp.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(header)); err != nil {
		return err
	}
	return nil
}
