package pgxtest_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyFS(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()

	fsys := fstest.MapFS{
		"0001_init.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE things (id text PRIMARY KEY);`)},
		"atlas.sum":     &fstest.MapFile{Data: []byte("ignored\n")}, // non-.sql must be skipped
	}
	require.NoError(t, pgxtest.ApplyFS(ctx, pool, fsys))

	var n int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM things`).Scan(&n))
	assert.Equal(t, 0, n)
}
