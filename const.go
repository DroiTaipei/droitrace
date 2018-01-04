package droitrace

const (
	TagError         = "error"
	TagDroiError     = "droi.error"
	TagDroiErrorCode = "droi.error_code"
	TagHTTPStatus    = "http.status_code"
)

const (
	ReferenceRoot        = SpanReference("root")
	ReferenceChildOf     = SpanReference("childOf")
	ReferenceFollowsFrom = SpanReference("followsFrom")
)
