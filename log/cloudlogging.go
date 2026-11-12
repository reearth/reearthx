package log

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
)

// https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
const (
	gcpTraceKey        = "logging.googleapis.com/trace"
	gcpSpanIDKey       = "logging.googleapis.com/spanId"
	gcpTraceSampledKey = "logging.googleapis.com/trace_sampled"
	gcpHTTPRequestKey  = "httpRequest"
)

func traceFields(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() || !isGCP() {
		return nil
	}
	googleTrace := GoogleTraceFromTraceID(sc.TraceID().String(), "")
	if googleTrace == "" {
		return nil
	}
	return []any{
		gcpTraceKey, googleTrace,
		gcpSpanIDKey, sc.SpanID().String(),
		gcpTraceSampledKey, sc.IsSampled(),
	}
}

func withTrace(l *Logger, ctx context.Context) *Logger {
	if f := traceFields(ctx); f != nil {
		return l.With(f...)
	}
	return l
}

// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
type httpRequest struct {
	method       string
	url          string
	requestSize  int64
	status       int
	responseSize int64
	userAgent    string
	remoteIP     string
	referer      string
	latency      time.Duration
	protocol     string
}

func (h httpRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("requestMethod", h.method)
	enc.AddString("requestUrl", h.url)
	if h.requestSize >= 0 {
		enc.AddString("requestSize", strconv.FormatInt(h.requestSize, 10))
	}
	enc.AddInt("status", h.status)
	enc.AddString("responseSize", strconv.FormatInt(h.responseSize, 10))
	if h.userAgent != "" {
		enc.AddString("userAgent", h.userAgent)
	}
	if h.remoteIP != "" {
		enc.AddString("remoteIp", h.remoteIP)
	}
	if h.referer != "" {
		enc.AddString("referer", h.referer)
	}
	enc.AddString("latency", strconv.FormatFloat(h.latency.Seconds(), 'f', -1, 64)+"s")
	enc.AddString("protocol", h.protocol)
	return nil
}

func httpRequestFields(req *http.Request, scheme, remoteIP string, status int, responseSize int64, latency time.Duration) []any {
	if !isGCP() {
		return nil
	}
	return []any{gcpHTTPRequestKey, httpRequest{
		method:       req.Method,
		url:          scheme + "://" + req.Host + req.URL.RequestURI(),
		requestSize:  req.ContentLength,
		status:       status,
		responseSize: responseSize,
		userAgent:    req.UserAgent(),
		remoteIP:     remoteIP,
		referer:      req.Referer(),
		latency:      latency,
		protocol:     req.Proto,
	}}
}
