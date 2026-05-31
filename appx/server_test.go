package appx

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

func TestApplyServerDefaults(t *testing.T) {
	cfg := ServerConfig{}
	applyServerDefaults(&cfg)
	assert.Equal(t, defaultReadHeaderTimeout, cfg.ReadHeaderTimeout)
	assert.Equal(t, defaultIdleTimeout, cfg.IdleTimeout)
	assert.Equal(t, defaultMaxHeaderBytes, cfg.MaxHeaderBytes)
	assert.Equal(t, defaultShutdownTimeout, cfg.ShutdownTimeout)

	cfg = ServerConfig{
		ReadHeaderTimeout: time.Second,
		IdleTimeout:       2 * time.Second,
		MaxHeaderBytes:    4096,
		ShutdownTimeout:   4 * time.Second,
	}
	applyServerDefaults(&cfg)
	assert.Equal(t, time.Second, cfg.ReadHeaderTimeout)
	assert.Equal(t, 2*time.Second, cfg.IdleTimeout)
	assert.Equal(t, 4096, cfg.MaxHeaderBytes)
	assert.Equal(t, 4*time.Second, cfg.ShutdownTimeout)
}

func TestStartServer_ServesAndShutsDown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "pong")
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- StartServer(ctx, ServerConfig{
			Listener: ln,
			Handler:  mux,
		})
	}()

	url := fmt.Sprintf("http://%s/ping", ln.Addr())
	requireEventually(t, 2*time.Second, func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	})

	cancel()
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("StartServer did not return after ctx cancel")
	}
}

func TestStartServer_H2C(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/proto", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Proto)
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- StartServer(ctx, ServerConfig{
			Listener: ln,
			Handler:  mux,
			H2C:      true,
		})
	}()
	defer func() {
		cancel()
		<-errCh
	}()

	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}

	url := fmt.Sprintf("http://%s/proto", ln.Addr())
	requireEventually(t, 2*time.Second, func() error {
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "HTTP/2.0" {
			return fmt.Errorf("proto=%q, want HTTP/2.0", body)
		}
		return nil
	})
}

func TestStartServer_RequiresHandler(t *testing.T) {
	err := StartServer(context.Background(), ServerConfig{})
	assert.Error(t, err)
}

func requireEventually(t *testing.T, within time.Duration, fn func() error) {
	t.Helper()
	deadline := time.Now().Add(within)
	var lastErr error
	for time.Now().Before(deadline) {
		if lastErr = fn(); lastErr == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s: %v", within, lastErr)
}
