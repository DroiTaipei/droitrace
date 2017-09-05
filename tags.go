package droitrace

import (
	"github.com/DroiTaipei/droictx"
	opentracing "github.com/DroiTaipei/opentracing-go"
	ext "github.com/DroiTaipei/opentracing-go/ext"

	"fmt"
	"net/http"
)

func GenDroiTag(sk string) string {
	return fmt.Sprintf("droi.%s", sk)
}

func SetRPCClientTag(span opentracing.Span) {
	ext.SpanKindRPCClient.Set(span)
}

func SetDroiTag(span opentracing.Span, key string, value interface{}) {
	hkMap := droictx.IFieldHeaderKeyMap()
	if sk := hkMap[key]; len(sk) > 0 {
		tag := GenDroiTag(sk)
		span.SetTag(tag, value)
	}
}

func SetDroiTagFromContext(span opentracing.Span, ctx droictx.Context) {
	keyMap := droictx.IFieldHeaderKeyMap()
	headers := ctx.HeaderMap()
	for hk, v := range headers {
		sk := keyMap[hk]
		if len(sk) > 0 {
			tag := GenDroiTag(sk)
			span.SetTag(tag, v)
		}
	}
	return
}

func SetDroiTagFromHeaders(span opentracing.Span, headers http.Header) {
	for hk, sk := range droictx.IFieldHeaderKeyMap() {
		if v := headers.Get(hk); len(v) > 0 {
			tag := GenDroiTag(sk)
			span.SetTag(tag, v)
		}
	}
}
