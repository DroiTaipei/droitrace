package droitrace

import (
	"github.com/DroiTaipei/droictx"
	"github.com/DroiTaipei/opentracing-go/mocktracer"

	"net/http"
	"strings"
	"testing"
)

var (
	appid    = "5e8umbzh"
	svcAppId = "497umbzh"
	rid      = "rid-test"
	dummy    = "dummyVal"
)

func TestSetDroiTag(t *testing.T) {
	tracer := mocktracer.New()
	span := tracer.StartSpan("my-trace")

	m := map[string]string{
		droictx.HTTPHeaderAppID:        appid,
		droictx.HTTPHeaderServiceAppID: svcAppId,
		droictx.HTTPHeaderRequestID:    rid,
		"not-droi-tag":                 dummy,
	}
	for k, v := range m {
		SetDroiTag(span, k, v)
	}
	span.Finish()

	rawSpan := tracer.FinishedSpans()[0]
	spanTags := rawSpan.Tags()
	if len(spanTags) != len(m)-1 {
		t.Errorf("Size should equal %d != %d", len(m)-1, len(spanTags))
	}

	skMap := droictx.IFieldShortKeyMap()
	for key, tagValue := range spanTags {
		keyArray := strings.Split(key, ".")
		if len(keyArray) != 2 {
			continue
		}
		hk := skMap[keyArray[1]]
		if len(m[hk]) > 0 {
			continue
		}
		if m[hk] != tagValue {
			t.Errorf("SetTag failed %s!=%s", m[hk], tagValue)
		}
	}
}

func TestSetDroiTagFromHeaders(t *testing.T) {
	tracer := mocktracer.New()
	span := tracer.StartSpan("my-trace")

	m := map[string]string{
		droictx.HTTPHeaderAppID:        appid,
		droictx.HTTPHeaderServiceAppID: svcAppId,
		droictx.HTTPHeaderRequestID:    rid,
	}

	headers := http.Header{}
	for k, v := range m {
		headers.Add(k, v)
	}

	SetDroiTagFromHeaders(span, headers)
	span.Finish()

	rawSpan := tracer.FinishedSpans()[0]
	spanTags := rawSpan.Tags()
	if len(spanTags) != len(m) {
		t.Errorf("Size should equal %d != %d", len(m), len(spanTags))
	}

	skMap := droictx.IFieldShortKeyMap()
	for key, tagValue := range spanTags {
		keyArray := strings.Split(key, ".")
		if len(keyArray) != 2 {
			t.Errorf("Bad tag format %s", key)
		}
		hk := skMap[keyArray[1]]
		if m[hk] != tagValue {
			t.Errorf("SetTag failed %s!=%s", m[hk], tagValue)
		}
	}
}

func TestSetDroiTagFromContext(t *testing.T) {
	tracer := mocktracer.New()
	span := tracer.StartSpan("my-trace")
	var ctx droictx.Context

	m := map[string]string{
		droictx.HTTPHeaderAppID:        appid,
		droictx.HTTPHeaderServiceAppID: svcAppId,
		droictx.HTTPHeaderRequestID:    rid,
	}
	for k, v := range m {
		ctx.HeaderSet(k, v)
	}
	SetDroiTagFromContext(span, ctx)
	span.Finish()

	rawSpan := tracer.FinishedSpans()[0]
	spanTags := rawSpan.Tags()
	if len(spanTags) != len(m) {
		t.Errorf("Size should equal %d != %d", len(m), len(spanTags))
	}
	skMap := droictx.IFieldShortKeyMap()
	for key, tagValue := range spanTags {
		keyArray := strings.Split(key, ".")
		if len(keyArray) != 2 {
			t.Errorf("Bad tag format %s", key)
		}
		hk := skMap[keyArray[1]]
		if m[hk] != tagValue {
			t.Errorf("SetTag failed %s!=%s", m[hk], tagValue)
		}
	}
}
