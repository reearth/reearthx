package log

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	echov5 "github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

const (
	testTraceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	testSpanID  = "00f067aa0ba902b7"
)

func useGCP(t *testing.T) {
	t.Helper()
	prev := GCP
	GCP = true
	t.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	t.Cleanup(func() { GCP = prev })
}

func contextWithSpan(t *testing.T, ctx context.Context) context.Context {
	t.Helper()
	tid, err := trace.TraceIDFromHex(testTraceID)
	require.NoError(t, err)
	sid, err := trace.SpanIDFromHex(testSpanID)
	require.NoError(t, err)
	return trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	}))
}

func entries(t *testing.T, w *bytes.Buffer) []map[string]any {
	t.Helper()
	var res []map[string]any
	scanner := bufio.NewScanner(w)
	for scanner.Scan() {
		var e map[string]any
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &e), scanner.Text())
		res = append(res, e)
	}
	return res
}

func assertTraceFields(t *testing.T, e map[string]any) {
	t.Helper()
	assert.Equal(t, "projects/test-project/traces/"+testTraceID, e["logging.googleapis.com/trace"])
	assert.Equal(t, testSpanID, e["logging.googleapis.com/spanId"])
	assert.Equal(t, true, e["logging.googleapis.com/trace_sampled"])
}

func assertNoTraceFields(t *testing.T, e map[string]any) {
	t.Helper()
	assert.NotContains(t, e, "logging.googleapis.com/trace")
	assert.NotContains(t, e, "logging.googleapis.com/spanId")
	assert.NotContains(t, e, "logging.googleapis.com/trace_sampled")
}

func TestContextFuncs_GCPTraceFields(t *testing.T) {
	useGCP(t)
	w := &bytes.Buffer{}
	SetOutput(w)
	t.Cleanup(func() { SetOutput(DefaultOutput) })

	ctx := contextWithSpan(t, context.Background())
	Infofc(ctx, "global logger %s", "with span")
	Infofc(ContextWith(ctx, NewWithOutput(w)), "context logger with span")
	Infoc(context.Background(), "no span")

	es := entries(t, w)
	require.Len(t, es, 3)
	assertTraceFields(t, es[0])
	assertTraceFields(t, es[1])
	assertNoTraceFields(t, es[2])
}

func TestContextFuncs_NoProjectNoTraceFields(t *testing.T) {
	useGCP(t)
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	w := &bytes.Buffer{}
	ctx := ContextWith(contextWithSpan(t, context.Background()), NewWithOutput(w))

	Infoc(ctx, "span but no project")

	es := entries(t, w)
	require.Len(t, es, 1)
	assertNoTraceFields(t, es[0])
}

func TestContextFuncs_GCPProjectVar(t *testing.T) {
	useGCP(t)
	t.Setenv("GOOGLE_CLOUD_PROJECT", "")
	prev := GCPProject
	GCPProject = "configured-project"
	t.Cleanup(func() { GCPProject = prev })
	w := &bytes.Buffer{}

	Infoc(ContextWith(contextWithSpan(t, context.Background()), NewWithOutput(w)), "span")

	es := entries(t, w)
	require.Len(t, es, 1)
	assert.Equal(t, "projects/configured-project/traces/"+testTraceID, es[0]["logging.googleapis.com/trace"])
}

func TestDurationString(t *testing.T) {
	for d, want := range map[time.Duration]string{
		0:                                "0s",
		123 * time.Millisecond:           "0.123s",
		time.Second + 5*time.Microsecond: "1.000005s",
		90 * time.Second:                 "90s",
		1234567891 * time.Nanosecond:     "1.234567891s",
		-1500 * time.Millisecond:         "-1.5s",
	} {
		assert.Equal(t, want, durationString(d), d.String())
	}
}

func TestSlogHandler_GCPTraceFields(t *testing.T) {
	useGCP(t)
	w := &bytes.Buffer{}
	l := NewSlogLogger(NewWithOutput(w))

	l.InfoContext(contextWithSpan(t, context.Background()), "with span")
	l.InfoContext(context.Background(), "no span")

	es := entries(t, w)
	require.Len(t, es, 2)
	assertTraceFields(t, es[0])
	assertNoTraceFields(t, es[1])
}

func assertAccessLog(t *testing.T, w *bytes.Buffer) {
	t.Helper()
	assert.Equal(t, 1, strings.Count(w.String(), `"httpRequest":`), w.String())
	es := entries(t, w)
	require.Len(t, es, 2)
	assert.NotContains(t, es[0], "httpRequest", "incoming entry")
	assertTraceFields(t, es[0])
	assertTraceFields(t, es[1])

	out := es[1]
	assert.Contains(t, out, "latency")
	assert.Equal(t, out["latency"], out["letency"])
	hr, ok := out["httpRequest"].(map[string]any)
	require.True(t, ok, "httpRequest object in outgoing entry: %v", out)
	assert.Equal(t, "GET", hr["requestMethod"])
	assert.Equal(t, "http://example.com/test", hr["requestUrl"])
	assert.Equal(t, float64(http.StatusCreated), hr["status"])
	assert.Equal(t, "2", hr["responseSize"])
	assert.Equal(t, "test-agent", hr["userAgent"])
	assert.Equal(t, "192.0.2.1", hr["remoteIp"])
	assert.Equal(t, "HTTP/1.1", hr["protocol"])
	assert.Regexp(t, `^\d+(\.\d{1,9})?s$`, hr["latency"])
}

func newAccessLogRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test?q=1", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	req.Header.Set("User-Agent", "test-agent")
	return req
}

func TestEcho_AccessLogger_GCP(t *testing.T) {
	useGCP(t)
	w := &bytes.Buffer{}

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.SetRequest(c.Request().WithContext(contextWithSpan(t, c.Request().Context())))
			return next(c)
		}
	})
	e.Use(NewEchoWithRaw(NewWithOutput(w)).AccessLogger())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusCreated, "ok")
	})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, newAccessLogRequest())
	require.Equal(t, http.StatusCreated, rec.Code)

	assertAccessLog(t, w)
}

func TestAccessLoggerV5_GCP(t *testing.T) {
	useGCP(t)
	w := &bytes.Buffer{}

	e := echov5.New()
	e.Use(func(next echov5.HandlerFunc) echov5.HandlerFunc {
		return func(c *echov5.Context) error {
			c.SetRequest(c.Request().WithContext(contextWithSpan(t, c.Request().Context())))
			return next(c)
		}
	})
	e.Use(AccessLoggerV5(NewWithOutput(w)))
	e.GET("/test", func(c *echov5.Context) error {
		return c.String(http.StatusCreated, "ok")
	})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, newAccessLogRequest())
	require.Equal(t, http.StatusCreated, rec.Code)

	assertAccessLog(t, w)
}

func TestEcho_AccessLogger_NotGCP(t *testing.T) {
	w := &bytes.Buffer{}
	e := echo.New()
	e.Use(NewEchoWithRaw(NewWithOutput(w)).AccessLogger())
	e.GET("/test", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	e.ServeHTTP(httptest.NewRecorder(), newAccessLogRequest())

	assert.NotContains(t, w.String(), "httpRequest")
	assert.Contains(t, w.String(), "latency")
	assert.Contains(t, w.String(), "letency")
}
