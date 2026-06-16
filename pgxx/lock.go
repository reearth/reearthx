package pgxx

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Unlock releases an advisory lock and returns the held connection to the pool.
type Unlock func(context.Context) error

// AdvisoryLock acquires a session-level Postgres advisory lock on key, blocking
// until it is available. It holds a dedicated pooled connection for the lock's
// lifetime; call the returned Unlock to release the lock and the connection.
func AdvisoryLock(ctx context.Context, pool *pgxpool.Pool, key int64) (Unlock, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, WrapError(err)
	}
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, key); err != nil {
		conn.Release()
		return nil, WrapError(err)
	}
	return unlocker(conn, key), nil
}

// TryAdvisoryLock attempts to acquire the advisory lock without blocking. If the
// lock is already held (by any session), it returns acquired=false and a nil
// Unlock. On success it returns an Unlock as in AdvisoryLock.
func TryAdvisoryLock(ctx context.Context, pool *pgxpool.Pool, key int64) (Unlock, bool, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, false, WrapError(err)
	}
	var ok bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, key).Scan(&ok); err != nil {
		conn.Release()
		return nil, false, WrapError(err)
	}
	if !ok {
		conn.Release()
		return nil, false, nil
	}
	return unlocker(conn, key), true, nil
}

func unlocker(conn *pgxpool.Conn, key int64) Unlock {
	var once sync.Once
	return func(ctx context.Context) error {
		var err error
		once.Do(func() {
			defer conn.Release()
			_, e := conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, key)
			err = WrapError(e)
		})
		return err // second and later calls are no-ops returning nil
	}
}
