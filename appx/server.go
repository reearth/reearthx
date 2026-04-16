package appx

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/reearth/reearthx/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// StartServer is the unified HTTP entry point for reearth services.
//
// It exists because the naive pattern — constructing http2.Server with only
// IdleTimeout and starting Echo via StartH2CServer — leaves every other
// server-level timeout unset. Slow or half-open clients can hold a
// connection open indefinitely, inflight requests accumulate, HTTP/2
// MaxConcurrentStreams (default 250) saturates, and the MongoDB connection
// pool exhausts — turning a partial slow dependency into a full service
// hang.
//
// To prevent that, this helper:
//
//   - Sets ReadHeaderTimeout, IdleTimeout, and MaxHeaderBytes explicitly on
//     http.Server so slow clients cannot hold connections open forever and
//     idle keep-alive connections get reaped.
//   - Leaves ReadTimeout and WriteTimeout unset on purpose: they cut off
//     streaming responses (SSE, WebSocket, large uploads/downloads).
//   - Wires h2c correctly and handles SIGTERM-driven graceful shutdown so
//     in-flight requests can drain before the process exits.
//
// Per-request deadlines are intentionally NOT enforced here. On Cloud Run
// the platform's request timeout already cancels r.Context() when the
// client connection closes, which is where downstream calls must honor
// ctx.Done(). Adding a middleware-level deadline would only duplicate that
// behavior and force every streaming route to opt out via a skipper.
//
// Defaults are tuned for Cloud Run request/idle limits.

// ServerConfig configures the unified HTTP server exposed by StartServer.
// A zero value on a duration or size field means "apply the default".
type ServerConfig struct {
	// Host is the bind interface. Empty means all interfaces.
	Host string
	// Port is the bind port. Ignored when Listener is set.
	Port int
	// Listener, if non-nil, is used directly and overrides Host/Port.
	// Useful for tests or when binding to an inherited socket.
	Listener net.Listener
	// Handler receives requests. Typically an *echo.Echo. Required.
	Handler http.Handler

	// H2C enables cleartext HTTP/2 (needed for gRPC/h2c behind Cloud Run).
	H2C bool
	// MaxConcurrentStreams caps HTTP/2 concurrent streams per connection.
	// Zero keeps the http2 default (250).
	MaxConcurrentStreams uint32

	// ReadHeaderTimeout bounds the request-header read phase.
	ReadHeaderTimeout time.Duration
	// IdleTimeout bounds keep-alive idle time between requests.
	IdleTimeout time.Duration
	// MaxHeaderBytes bounds the request-header size.
	MaxHeaderBytes int

	// ShutdownTimeout bounds graceful shutdown once ctx is canceled.
	ShutdownTimeout time.Duration
}

const (
	defaultReadHeaderTimeout = 60 * time.Second
	defaultIdleTimeout       = 15 * time.Minute
	defaultMaxHeaderBytes    = 1 << 20
	defaultShutdownTimeout   = 30 * time.Second
)

// StartServer binds an HTTP server, serves until ctx is canceled, and then
// gracefully shuts down. It returns nil on a clean shutdown or the first
// non-ErrServerClosed error from the server.
func StartServer(ctx context.Context, cfg ServerConfig) error {
	if cfg.Handler == nil {
		return errors.New("appx: ServerConfig.Handler is required")
	}
	applyServerDefaults(&cfg)

	handler := cfg.Handler
	if cfg.H2C {
		handler = h2c.NewHandler(handler, &http2.Server{
			IdleTimeout:          cfg.IdleTimeout,
			MaxConcurrentStreams: cfg.MaxConcurrentStreams,
		})
	}

	ln := cfg.Listener
	if ln == nil {
		var err error
		ln, err = net.Listen("tcp", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)))
		if err != nil {
			return err
		}
	}

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	errCh := make(chan error, 1)
	go func() {
		log.Infof("appx: http server listening on %s (h2c=%v)", ln.Addr(), cfg.H2C)
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Infof("appx: shutting down http server (timeout=%s)", cfg.ShutdownTimeout)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return <-errCh
}

func applyServerDefaults(cfg *ServerConfig) {
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = defaultIdleTimeout
	}
	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = defaultMaxHeaderBytes
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = defaultShutdownTimeout
	}
}
