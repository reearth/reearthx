package pgxx_test

import (
	"context"
	"io"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvisoryLock(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	const key int64 = 0x52454143

	unlock, err := pgxx.AdvisoryLock(ctx, pool, key)
	require.NoError(t, err)

	// While held, a non-blocking try from another session must fail.
	u2, ok, err := pgxx.TryAdvisoryLock(ctx, pool, key)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, u2)

	// After release, the lock is acquirable again.
	require.NoError(t, unlock(ctx))
	u3, ok, err := pgxx.TryAdvisoryLock(ctx, pool, key)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, u3(ctx))
}

// Unlock must release the lock even when the caller's context is already done.
func TestAdvisoryLock_UnlockWithCancelledContext(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	const key int64 = 0x52454144

	unlock, err := pgxx.AdvisoryLock(context.Background(), pool, key)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, unlock(ctx))

	u, ok, err := pgxx.TryAdvisoryLock(context.Background(), pool, key)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, u(context.Background()))
}

// Unlock must return within the unlock timeout when the network to Postgres
// stops delivering packets.
func TestAdvisoryLock_UnlockIsBounded(t *testing.T) {
	uri := os.Getenv(pgxtest.Env)
	if uri == "" {
		t.Skipf("pgxtest: %s not set; skipping Postgres integration test", pgxtest.Env)
	}
	cfg, err := pgxpool.ParseConfig(uri)
	require.NoError(t, err)

	var blackhole atomic.Bool
	proxy := startBlackholeProxy(t, net.JoinHostPort(cfg.ConnConfig.Host, strconv.Itoa(int(cfg.ConnConfig.Port))), &blackhole)
	host, port, err := net.SplitHostPort(proxy)
	require.NoError(t, err)
	p, err := strconv.Atoi(port)
	require.NoError(t, err)
	cfg.ConnConfig.Host = host
	cfg.ConnConfig.Port = uint16(p)

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	defer pgxx.SetUnlockTimeout(300 * time.Millisecond)()

	unlock, err := pgxx.AdvisoryLock(context.Background(), pool, 0x52454145)
	require.NoError(t, err)

	blackhole.Store(true)
	start := time.Now()
	assert.Error(t, unlock(context.Background()))
	assert.Less(t, time.Since(start), 3*time.Second)
}

// startBlackholeProxy forwards TCP to target until blackhole is set, after
// which client bytes are silently dropped.
func startBlackholeProxy(t *testing.T, target string, blackhole *atomic.Bool) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			client, err := ln.Accept()
			if err != nil {
				return
			}
			server, err := net.Dial("tcp", target)
			if err != nil {
				_ = client.Close()
				continue
			}
			t.Cleanup(func() { _ = client.Close(); _ = server.Close() })
			go func() { _, _ = io.Copy(client, server) }()
			go func() {
				buf := make([]byte, 32*1024)
				for {
					n, err := client.Read(buf)
					if err != nil {
						return
					}
					if blackhole.Load() {
						continue
					}
					if _, err := server.Write(buf[:n]); err != nil {
						return
					}
				}
			}()
		}
	}()
	return ln.Addr().String()
}
