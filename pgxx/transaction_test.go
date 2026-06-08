package pgxx_test

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type queryFn struct {
	count  func(context.Context) int
	insert func(context.Context, string) error
}

func setupScratch(t *testing.T) (context.Context, usecasex.Transaction, queryFn) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE items (id text PRIMARY KEY)`)
	require.NoError(t, err)

	count := func(ctx context.Context) int {
		var n int
		err := pgxx.Executor(ctx, pool).QueryRow(ctx, `SELECT count(*) FROM items`).Scan(&n)
		require.NoError(t, err)
		return n
	}
	insert := func(ctx context.Context, id string) error {
		_, err := pgxx.Executor(ctx, pool).Exec(ctx, `INSERT INTO items (id) VALUES ($1)`, id)
		return err
	}
	return ctx, pgxx.NewTransaction(pool), queryFn{count: count, insert: insert}
}

func TestTransaction_CommitsOnCommit(t *testing.T) {
	ctx, tr, q := setupScratch(t)
	tx, err := tr.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, q.insert(tx.Context(), "a"))
	tx.Commit()
	require.NoError(t, tx.End(tx.Context()))
	assert.Equal(t, 1, q.count(ctx))
}

func TestTransaction_RollsBackWithoutCommit(t *testing.T) {
	ctx, tr, q := setupScratch(t)
	tx, err := tr.Begin(ctx)
	require.NoError(t, err)
	require.NoError(t, q.insert(tx.Context(), "a"))
	require.NoError(t, tx.End(tx.Context()))
	assert.Equal(t, 0, q.count(ctx))
}

func TestTransaction_DoTransactionCommits(t *testing.T) {
	ctx, tr, q := setupScratch(t)
	err := usecasex.DoTransaction(ctx, tr, 1, func(ctx context.Context) error {
		return q.insert(ctx, "x")
	})
	require.NoError(t, err)
	assert.Equal(t, 1, q.count(ctx))
}
