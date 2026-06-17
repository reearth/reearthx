package pgxx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- #1 retry contract ---

func TestClient_WithinTransaction_RetriesOnErrTransaction(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	c := pgxx.NewClient(pool, pgxx.WithTxRetry(2))

	calls := 0
	err := c.WithinTransaction(ctx, func(ctx context.Context) error {
		calls++
		if calls == 1 {
			return pgxx.WrapError(&pgconn.PgError{Code: "40001"}) // serialization -> retryable
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 2, calls, "fn should run twice: fail once (retryable) then succeed")
}

func TestClient_WithinTransaction_NoRetryByDefault(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	c := pgxx.NewClient(pool) // default: single attempt

	calls := 0
	err := c.WithinTransaction(ctx, func(ctx context.Context) error {
		calls++
		return pgxx.WrapError(&pgconn.PgError{Code: "40001"})
	})
	assert.ErrorIs(t, err, usecasex.ErrTransaction)
	assert.Equal(t, 1, calls, "default client must not retry")
}

// --- #5 nested WithinTransaction commits via the single outer tx ---

func TestClient_WithinTransaction_NestedCommits(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE nested_items (id text PRIMARY KEY)`)
	require.NoError(t, err)
	c := pgxx.NewClient(pool)
	ins := func(ctx context.Context, id string) error {
		_, e := c.DB(ctx).Exec(ctx, `INSERT INTO nested_items (id) VALUES ($1)`, id)
		return e
	}

	require.NoError(t, c.WithinTransaction(ctx, func(ctx context.Context) error {
		if e := ins(ctx, "a"); e != nil {
			return e
		}
		return c.WithinTransaction(ctx, func(ctx context.Context) error { return ins(ctx, "b") })
	}))

	var n int
	require.NoError(t, c.DB(ctx).QueryRow(ctx, `SELECT count(*) FROM nested_items`).Scan(&n))
	assert.Equal(t, 2, n)
}

// --- #2 double-unlock safety ---

func TestAdvisoryLock_UnlockIsIdempotent(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	unlock, err := pgxx.AdvisoryLock(ctx, pool, 777)
	require.NoError(t, err)
	require.NoError(t, unlock(ctx))
	require.NotPanics(t, func() {
		assert.NoError(t, unlock(ctx)) // second call: no-op, no panic
	})
}

// --- #7 CountAndList placeholder math + caller-slice isolation ---

type ciItem struct {
	ID  string
	Grp string
}

func TestCountAndList_ArgsIsolationAndPaging(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, `CREATE TABLE ci_items (id text PRIMARY KEY, grp text)`)
	require.NoError(t, err)
	for _, id := range []string{"a", "b", "c"} {
		_, e := pool.Exec(ctx, `INSERT INTO ci_items (id, grp) VALUES ($1, 'g')`, id)
		require.NoError(t, e)
	}
	db := pgxx.NewClient(pool).DB(ctx)
	args := []any{"g"}

	rows, total, err := pgxx.CountAndList[ciItem](ctx, db,
		`SELECT count(*) FROM ci_items WHERE grp = $1`,
		`SELECT id, grp FROM ci_items WHERE grp = $1 ORDER BY id`,
		args, 2, 1, pgx.RowToStructByPos[ciItem])
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, rows, 2) // LIMIT 2 OFFSET 1 -> b, c
	assert.Equal(t, "b", rows[0].ID)
	assert.Equal(t, "c", rows[1].ID)
	assert.Equal(t, []any{"g"}, args, "caller args slice must not be mutated")
}

// --- #8 OrderByIDs edges ---

func TestOrderByIDs_Edges(t *testing.T) {
	type item struct{ id, v string }
	key := func(i *item) string { return i.id }
	items := []*item{{"a", "A"}, {"a", "A2"}} // duplicate key: last wins

	got := pgxx.OrderByIDs([]string{"a", "a", "x"}, items, key)
	require.Len(t, got, 3)
	assert.Equal(t, "A2", got[0].v)
	assert.Equal(t, "A2", got[1].v) // duplicate id resolves to the same item
	assert.Nil(t, got[2])           // missing id -> nil

	assert.Empty(t, pgxx.OrderByIDs[string, *item](nil, nil, key))
}

// --- #11 JSONB null / malformed ---

func TestJSONBSlice_NullAndMalformed(t *testing.T) {
	out, err := pgxx.UnmarshalJSONBSlice[kv]([]byte("null"))
	require.NoError(t, err)
	assert.Nil(t, out) // JSON null decodes to a nil slice

	_, err = pgxx.UnmarshalJSONBSlice[kv]([]byte("{not json"))
	assert.Error(t, err)
}

// --- #12 MapError passes unique violations through ---

func TestMapError_UniqueViolationPassesThrough(t *testing.T) {
	uerr := &pgconn.PgError{Code: "23505"}
	got := pgxx.MapError(uerr)
	assert.False(t, errors.Is(got, rerror.ErrNotFound))
	assert.False(t, errors.Is(got, usecasex.ErrTransaction))
	assert.True(t, pgxx.IsUniqueViolation(got))
}
