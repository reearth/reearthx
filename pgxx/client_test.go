package pgxx_test

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupItems(t *testing.T) (context.Context, *pgxx.Client, func(context.Context) int, func(context.Context, string) error) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE items (id text PRIMARY KEY)`)
	require.NoError(t, err)
	c := pgxx.NewClient(pool)
	count := func(ctx context.Context) int {
		var n int
		require.NoError(t, c.DB(ctx).QueryRow(ctx, `SELECT count(*) FROM items`).Scan(&n))
		return n
	}
	insert := func(ctx context.Context, id string) error {
		_, err := c.DB(ctx).Exec(ctx, `INSERT INTO items (id) VALUES ($1)`, id)
		return err
	}
	return ctx, c, count, insert
}

func TestClient_WithinTransaction_Commits(t *testing.T) {
	ctx, c, count, insert := setupItems(t)
	err := c.WithinTransaction(ctx, func(ctx context.Context) error { return insert(ctx, "a") })
	require.NoError(t, err)
	assert.Equal(t, 1, count(ctx))
}

func TestClient_WithinTransaction_RollsBackOnError(t *testing.T) {
	ctx, c, count, insert := setupItems(t)
	sentinel := assert.AnError
	err := c.WithinTransaction(ctx, func(ctx context.Context) error {
		if e := insert(ctx, "a"); e != nil {
			return e
		}
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, 0, count(ctx))
}

// Nested WithinTransaction must reuse the ambient tx: a rollback at the outer
// level discards the inner insert too (no independent nested BEGIN).
func TestClient_WithinTransaction_NestedComposes(t *testing.T) {
	ctx, c, count, insert := setupItems(t)
	sentinel := assert.AnError
	err := c.WithinTransaction(ctx, func(ctx context.Context) error {
		if e := insert(ctx, "a"); e != nil {
			return e
		}
		if e := c.WithinTransaction(ctx, func(ctx context.Context) error {
			return insert(ctx, "b")
		}); e != nil {
			return e
		}
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, 0, count(ctx), "nested insert must roll back with the outer tx")
}
