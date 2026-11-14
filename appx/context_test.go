package appx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/reearth/reearthx/log"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestContextMiddleware(t *testing.T) {
	key := struct{}{}
	value := lo.ToPtr("a")
	ts := httptest.NewServer(ContextMiddleware(key, value)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(*(r.Context().Value(key).(*string))))
	})))
	defer ts.Close()

	res := lo.Must(http.Get(ts.URL))
	body := string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "a", body)
}

func TestContextMiddlewareBy(t *testing.T) {
	type keys struct{}
	key := keys{}
	ts := httptest.NewServer(ContextMiddlewareBy(func(w http.ResponseWriter, r *http.Request) context.Context {
		if r.Method == http.MethodPost {
			return context.WithValue(r.Context(), key, "aaa")
		}
		return nil
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v, ok := r.Context().Value(key).(string); ok {
			_, _ = w.Write([]byte(v))
		}
	})))
	defer ts.Close()

	res := lo.Must(http.Get(ts.URL))
	body := string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "", body)

	res = lo.Must(http.Post(ts.URL, "", nil))
	body = string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "aaa", body)
}

func TestRequestIDMiddleware(t *testing.T) {
	ts := httptest.NewServer(RequestIDMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(GetRequestIDFromContext(r.Context())))
	})))
	defer ts.Close()

	req := lo.Must(http.NewRequest(http.MethodGet, ts.URL, nil))
	req.Header.Set("X-Request-ID", "aaa")
	res := lo.Must(http.DefaultClient.Do(req))
	body := string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "aaa", body)

	req = lo.Must(http.NewRequest(http.MethodGet, ts.URL, nil))
	req.Header.Set("x-cloud-trace-context", "xxx")
	res = lo.Must(http.DefaultClient.Do(req))
	body = string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "xxx", body)
}

func TestRequestIDMiddleware_TraceFieldOnceWithSpan(t *testing.T) {
	prev := log.GCP
	log.GCP = true
	t.Cleanup(func() { log.GCP = prev })
	t.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")

	w := &bytes.Buffer{}
	tid := lo.Must(trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736"))
	sid := lo.Must(trace.SpanIDFromHex("00f067aa0ba902b7"))
	withSpan := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := log.ContextWith(r.Context(), log.NewWithOutput(w))
			ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{TraceID: tid, SpanID: sid}))
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
	h := withSpan(RequestIDMiddleware()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		log.Infoc(r.Context(), "handled")
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")
	h.ServeHTTP(httptest.NewRecorder(), req)

	out := w.String()
	require.Contains(t, out, "handled")
	assert.Equal(t, 1, strings.Count(out, `"logging.googleapis.com/trace":`), out)
}

func TestRequestIDMiddleware_HeaderTraceUsesGCPProject(t *testing.T) {
	prevGCP, prevProject := log.GCP, log.GCPProject
	log.GCP, log.GCPProject = true, "configured-project"
	t.Cleanup(func() { log.GCP, log.GCPProject = prevGCP, prevProject })
	t.Setenv("GOOGLE_CLOUD_PROJECT", "env-project")

	w := &bytes.Buffer{}
	withLogger := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(rw, r.WithContext(log.ContextWith(r.Context(), log.NewWithOutput(w))))
		})
	}
	h := withLogger(RequestIDMiddleware()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		log.Infoc(r.Context(), "handled")
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00")
	h.ServeHTTP(httptest.NewRecorder(), req)

	assert.Contains(t, w.String(), `"logging.googleapis.com/trace":"projects/configured-project/traces/4bf92f3577b34da6a3ce929d0e0e4736"`, w.String())
}
