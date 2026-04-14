package log

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	echov5 "github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Contains(t, out, "--> GET 200 /test")
	assert.Contains(t, out, "method")
	assert.Contains(t, out, "latency")
}
