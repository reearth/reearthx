package appx_test

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/reearth/reearthx/appx"
)

// ExampleStartServer shows the recommended wiring: an Echo handler served
// over h2c, bound to a SIGTERM-aware context so the server drains
// gracefully on shutdown.
func ExampleStartServer() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	e := echo.New()
	e.GET("/healthz", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	_ = appx.StartServer(ctx, appx.ServerConfig{
		Handler: e,
		Port:    8080,
		H2C:     true,
	})
}

// ExampleStartServer_minimal shows the smallest correct invocation: any
// http.Handler plus a port. All timeouts fall back to Cloud Run-tuned
// defaults.
func ExampleStartServer_minimal() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	_ = appx.StartServer(context.Background(), appx.ServerConfig{
		Handler: mux,
		Port:    8080,
	})
}
