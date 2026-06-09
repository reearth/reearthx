package pgxx_test

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupItems(t *testing.T) (context.Context, *pgxx.Transactor, func(context.Context) int, func(context.Context, string) error) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE items (id text PRIMARY KEY)`)
	require.NoError(t, err)
	count := func(ctx context.Context) int {
		var n int
		require.NoError(t, pgxx.Executor(ctx, pool).QueryRow(ctx, `SELECT count(*) FROM items`).Scan(&n))
		return n
	}
	insert := func(ctx context.Context, id string) error {
		_, err := pgxx.Executor(ctx, pool).Exec(ctx, `INSERT INTO items (id) VALUES ($1)`, id)
		return err
	}
	return ctx, pgxx.NewTransactor(pool, 0), count, insert
}

func TestTransactor_Commits(t *testing.T) {
	ctx, tr, count, insert := setupItems(t)
	err := tr.WithinTransaction(ctx, func(ctx context.Context) error { return insert(ctx, "a") })
	require.NoError(t, err)
	assert.Equal(t, 1, count(ctx))
}

func TestTransactor_RollsBackOnError(t *testing.T) {
	ctx, tr, count, insert := setupItems(t)
	sentinel := assert.AnError
	err := tr.WithinTransaction(ctx, func(ctx context.Context) error {
		if e := insert(ctx, "a"); e != nil {
			return e
		}
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, 0, count(ctx))
}
