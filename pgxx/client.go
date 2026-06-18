package pgxx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

// Client is a pgx-backed database handle — the Postgres analogue of mongox.Client.
// It resolves the executor for the current context and runs functions within a
// transaction. It implements usecasex.Transactor, so it can back
// repo.Container.Transaction directly.
type Client struct {
	pool    *pgxpool.Pool
	retries int
}

var _ usecasex.Transactor = (*Client)(nil)

// Option configures a Client.
type Option func(*Client)

// WithTxRetry makes WithinTransaction retry up to n additional times when the
// callback returns an error matching usecasex.ErrTransaction (e.g. a Postgres
// serialization failure mapped via WrapError/MapError). Each retry re-begins a
// fresh transaction. Default is 0 (single attempt). Nested calls (an ambient tx
// already in ctx) never retry — only the outermost call owns the transaction.
func WithTxRetry(n int) Option {
	return func(c *Client) { c.retries = n }
}

func NewClient(pool *pgxpool.Pool, opts ...Option) *Client {
	c := &Client{pool: pool}
	for _, o := range opts {
		o(c)
	}
	return c
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
// that transaction — nested calls compose, with no nested BEGIN (and no retry).
// When configured with WithTxRetry, the outermost call retries on
// usecasex.ErrTransaction with a fresh transaction.
func (c *Client) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}
	var err error
	for attempt := 0; ; attempt++ {
		err = c.runOnce(ctx, fn)
		if err == nil || !errors.Is(err, usecasex.ErrTransaction) || attempt >= c.retries {
			return err
		}
	}
}

func (c *Client) runOnce(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	// Roll back on any early return or panic; a no-op after a successful Commit
	// (the tx is already closed). Without it, a panic in fn would skip the
	// rollback and leak the pooled connection, eventually exhausting the pool.
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(ContextWithTx(ctx, tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return rerror.ErrInternalByWithContext(ctx, err)
	}
	return nil
}
