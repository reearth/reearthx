package pgxx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reearth/reearthx/usecasex"
)

// Transactor is a pgx-backed usecasex.Transactor. WithinTransaction begins a
// transaction, stores it in the context (see Executor), runs fn, then commits on
// success or rolls back on error. Serialization failures (see WrapError) are
// retried up to retries additional times.
type Transactor struct {
	pool    *pgxpool.Pool
	retries int
}

var _ usecasex.Transactor = (*Transactor)(nil)

// NewTransactor returns a pgx Transactor. retries is the number of EXTRA attempts
// on serialization failure (0 = a single attempt).
func NewTransactor(pool *pgxpool.Pool, retries int) *Transactor {
	return &Transactor{pool: pool, retries: retries}
}

func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	var err error
	for attempt := 0; ; attempt++ {
		err = t.runOnce(ctx, fn)
		if err == nil || !errors.Is(err, usecasex.ErrTransaction) || attempt >= t.retries {
			return err
		}
	}
}

func (t *Transactor) runOnce(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}
	txCtx := ContextWithTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(context.Background())
		return err
	}
	if err := tx.Commit(context.Background()); err != nil {
		return WrapError(err)
	}
	return nil
}
