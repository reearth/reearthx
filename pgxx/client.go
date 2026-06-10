package pgxx

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

// Client is a pgx-backed database handle — the Postgres analogue of mongox.Client.
// It resolves the executor for the current context and runs functions within a
// transaction. It implements usecasex.Transactor, so it can back
// repo.Container.Transaction directly.
type Client struct {
	pool *pgxpool.Pool
}

var _ usecasex.Transactor = (*Client)(nil)

func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{pool: pool}
}

// Pool returns the underlying connection pool.
func (c *Client) Pool() *pgxpool.Pool { return c.pool }

// DB returns the executor for ctx: the ambient transaction if one is active
// (see WithinTransaction), otherwise the pool. Repositories build their sqlc
// queries with gen.New(client.DB(ctx)) so they transparently join a transaction.
func (c *Client) DB(ctx context.Context) DBTX {
	return Executor(ctx, c.pool)
}

// WithinTransaction runs fn inside a transaction, committing on a nil return and
// rolling back on error. If a transaction is already active in ctx, fn runs on
// that transaction — nested calls compose, with no nested BEGIN.
func (c *Client) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	if err := fn(ContextWithTx(ctx, tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	return nil
}
