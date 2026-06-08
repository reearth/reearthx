package pgxx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reearth/reearthx/usecasex"
	"go.uber.org/atomic"
)

// Transaction is a pgx-backed implementation of usecasex.Transaction.
type Transaction struct {
	pool *pgxpool.Pool
}

var _ usecasex.Transaction = (*Transaction)(nil)

func NewTransaction(pool *pgxpool.Pool) *Transaction {
	return &Transaction{pool: pool}
}

// Begin opens a pgx transaction and returns a Tx whose Context() carries it, so
// repositories using Executor(ctx, pool) run on the transaction's connection.
func (t *Transaction) Begin(ctx context.Context) (usecasex.Tx, error) {
	pgtx, err := t.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: pgtx, ctx: ContextWithTx(ctx, pgtx)}, nil
}

type Tx struct {
	tx        pgx.Tx
	ctx       context.Context
	committed atomic.Bool
}

var _ usecasex.Tx = (*Tx)(nil)

func (t *Tx) Context() context.Context { return t.ctx }

func (t *Tx) Commit() { t.committed.Store(true) }

func (t *Tx) IsCommitted() bool { return t.committed.Load() }

// End commits if Commit() was called, otherwise rolls back. A rollback after a
// successful commit is a no-op in pgx, so the post-commit path is safe.
func (t *Tx) End(ctx context.Context) error {
	if t.committed.Load() {
		return t.tx.Commit(ctx)
	}
	return t.tx.Rollback(ctx)
}
