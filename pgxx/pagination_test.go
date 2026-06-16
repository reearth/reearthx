package pgxx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pagItem struct {
	ID string
}

func TestCountAndList(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE items (id text PRIMARY KEY)`)
	require.NoError(t, err)
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		_, err := pool.Exec(ctx, `INSERT INTO items (id) VALUES ($1)`, id)
		require.NoError(t, err)
	}
	db := pgxx.NewClient(pool).DB(ctx)

	rows, total, err := pgxx.CountAndList[pagItem](ctx, db,
		`SELECT count(*) FROM items`,
		`SELECT * FROM items ORDER BY id`,
		nil, 2, 2, pgx.RowToStructByPos[pagItem])
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	require.Len(t, rows, 2)
	assert.Equal(t, "c", rows[0].ID)
	assert.Equal(t, "d", rows[1].ID)

	all, err := pgxx.List[pagItem](ctx, db, `SELECT * FROM items ORDER BY id`, nil, pgx.RowToStructByPos[pagItem])
	require.NoError(t, err)
	assert.Len(t, all, 5)
}
