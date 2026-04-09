package log

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	echov5 "github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSlogLogger(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l)

	sl.Info("hello slog", "key", "value")

	out := w.String()
	assert.Contains(t, out, "hello slog")
	assert.Contains(t, out, "key")
	assert.Contains(t, out, "value")
}

func TestSlogHandler_Levels(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l)

	sl.Debug("debug msg")
	sl.Info("info msg")
	sl.Warn("warn msg")
	sl.Error("error msg")

	out := w.String()
	assert.Contains(t, out, "debug msg")
	assert.Contains(t, out, "info msg")
	assert.Contains(t, out, "warn msg")
	assert.Contains(t, out, "error msg")
}

func TestSlogHandler_WithAttrs(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l).With("persistent", "attr")

	sl.Info("msg with attr")

	out := w.String()
	assert.Contains(t, out, "persistent")
	assert.Contains(t, out, "attr")
}

func TestSlogHandler_WithGroup(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)
	sl := NewSlogLogger(l).WithGroup("grp")

	sl.Info("grouped", slog.String("key", "val"))

	out := w.String()
	assert.Contains(t, out, "grp.key")
	assert.Contains(t, out, "val")
}

func TestAccessLoggerV5(t *testing.T) {
	w := &bytes.Buffer{}
	l := NewWithOutput(w)

	e := echov5.New()
	e.Use(AccessLoggerV5(l))
	e.GET("/test", func(c *echov5.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	out := w.String()
	assert.Contains(t, out, "<-- GET /test")
	assert.Contains(t, out, "--> GET /test")
	assert.Contains(t, out, "method")
	assert.Contains(t, out, "latency")
}
