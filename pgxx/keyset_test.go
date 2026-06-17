package pgxx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ksRow struct{ ID string }

func ksCursor(r ksRow) usecasex.Cursor { return usecasex.Cursor(r.ID) }

func setupKeyset(t *testing.T) (context.Context, pgxx.DBTX) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE ks (id text PRIMARY KEY)`)
	require.NoError(t, err)
	for _, id := range []string{"a", "b", "c", "d", "e"} {
		_, e := pool.Exec(ctx, `INSERT INTO ks (id) VALUES ($1)`, id)
		require.NoError(t, e)
	}
	return ctx, pgxx.NewClient(pool).DB(ctx)
}

func keyset(t *testing.T, ctx context.Context, db pgxx.DBTX, p *usecasex.CursorPagination) ([]ksRow, *usecasex.PageInfo) {
	rows, info, err := pgxx.KeysetPaginate[ksRow](ctx, db, "ks", "id", "", nil, "id", p, pgx.RowToStructByPos[ksRow], ksCursor)
	require.NoError(t, err)
	return rows, info
}

func ids(rows []ksRow) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.ID
	}
	return out
}

func TestKeysetPaginate_Forward(t *testing.T) {
	ctx, db := setupKeyset(t)
	n := int64(2)
	rows, info := keyset(t, ctx, db, &usecasex.CursorPagination{First: &n})
	assert.Equal(t, []string{"a", "b"}, ids(rows))
	assert.Equal(t, int64(5), info.TotalCount)
	assert.True(t, info.HasNextPage)
	assert.False(t, info.HasPreviousPage)
	assert.Equal(t, usecasex.Cursor("a"), *info.StartCursor)
	assert.Equal(t, usecasex.Cursor("b"), *info.EndCursor)
}

func TestKeysetPaginate_ForwardAfter(t *testing.T) {
	ctx, db := setupKeyset(t)
	n, after := int64(2), usecasex.Cursor("b")
	rows, info := keyset(t, ctx, db, &usecasex.CursorPagination{First: &n, After: &after})
	assert.Equal(t, []string{"c", "d"}, ids(rows))
	assert.True(t, info.HasNextPage)     // "e" remains
	assert.True(t, info.HasPreviousPage) // After set
}

func TestKeysetPaginate_Backward(t *testing.T) {
	ctx, db := setupKeyset(t)
	n := int64(2)
	rows, info := keyset(t, ctx, db, &usecasex.CursorPagination{Last: &n})
	assert.Equal(t, []string{"d", "e"}, ids(rows)) // ascending order restored
	assert.True(t, info.HasPreviousPage)           // more before "d"
	assert.False(t, info.HasNextPage)
}

func TestKeysetPaginate_BackwardBefore(t *testing.T) {
	ctx, db := setupKeyset(t)
	n, before := int64(2), usecasex.Cursor("d")
	rows, info := keyset(t, ctx, db, &usecasex.CursorPagination{Last: &n, Before: &before})
	assert.Equal(t, []string{"b", "c"}, ids(rows))
	assert.True(t, info.HasNextPage)     // Before set
	assert.True(t, info.HasPreviousPage) // "a" remains before "b"
}
