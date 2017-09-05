package droitrace

import (
	"fmt"
	jaeger "github.com/DroiTaipei/jaeger-client-go"
	jaegercfg "github.com/DroiTaipei/jaeger-client-go/config"
	_ "github.com/DroiTaipei/jaeger-client-go/log"
	opentracing "github.com/DroiTaipei/opentracing-go"
	ext "github.com/DroiTaipei/opentracing-go/ext"
	zipkin "github.com/DroiTaipei/zipkin-go-opentracing"
	metrics "github.com/uber/jaeger-lib/metrics"
	"net"
	"net/http"
	"strconv"
	"time"
)

func InitJaeger(collectRate float64, componentName string) error {
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: collectRate,
		},
		Reporter: &jaegercfg.ReporterConfig{
			BufferFlushInterval: time.Duration(5 * time.Second),
		},
	}
	// TO-DO: Add droi logger
	//jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	tracer, _, err := cfg.New(componentName,
		//jaegercfg.Logger(jLogger),
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

func CreateSpanFromReq(req *http.Request) opentracing.Span {
	var sp opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		sp = opentracing.StartSpan(
			fmt.Sprintf("%s %s", req.Method, req.URL.String()))
	} else {
		sp = opentracing.StartSpan(
			fmt.Sprintf("%s %s", req.Method, req.URL.String()),
			ext.RPCServerOption(wireContext))
	}
	attachSpanTags(sp, req)
	return sp
}

func CreateRootSpan(req *http.Request) opentracing.Span {
	sp := opentracing.StartSpan(fmt.Sprintf("%s %s", req.Method, req.URL.RequestURI()))
	attachSpanTags(sp, req)
	return sp
}

func CreateChildSpan(parentSpan opentracing.Span, req *http.Request) opentracing.Span {
	sp := opentracing.StartSpan(
		fmt.Sprintf("%s %s", req.Method, req.URL.RequestURI()),
		opentracing.ChildOf(parentSpan.Context()))
	attachSpanTags(sp, req)
	return sp
}

func CreateFollowFromSpan(parentSpan opentracing.Span, req *http.Request) opentracing.Span {
	sp := opentracing.StartSpan(
		fmt.Sprintf("%s %s", req.Method, req.URL.RequestURI()),
		opentracing.FollowsFrom(parentSpan.Context()))
	attachSpanTags(sp, req)
	return sp
}

func attachSpanTags(sp opentracing.Span, req *http.Request) {
	ext.HTTPMethod.Set(sp, req.Method)
	ext.HTTPUrl.Set(sp, req.URL.String())
	if host, portString, err := net.SplitHostPort(req.URL.Host); err == nil {
		ext.PeerHostname.Set(sp, host)
		if port, err := strconv.Atoi(portString); err != nil {
			ext.PeerPort.Set(sp, uint16(port))
		}
	} else {
		ext.PeerHostname.Set(sp, req.URL.Host)
	}
	SetDroiTagFromHeaders(sp, req.Header)
	return
}

func InjectSpan(sp opentracing.Span, req *http.Request) error {
	if err := sp.Tracer().Inject(sp.Context(),
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
		return err
	}
	return nil
}
