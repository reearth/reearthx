package pgxx_test

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvisoryLock(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	const key int64 = 0x52454143

	unlock, err := pgxx.AdvisoryLock(ctx, pool, key)
	require.NoError(t, err)

	// While held, a non-blocking try from another session must fail.
	u2, ok, err := pgxx.TryAdvisoryLock(ctx, pool, key)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, u2)

	// After release, the lock is acquirable again.
	require.NoError(t, unlock(ctx))
	u3, ok, err := pgxx.TryAdvisoryLock(ctx, pool, key)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, u3(ctx))
}
