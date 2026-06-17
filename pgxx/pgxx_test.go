package pgxx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reearth/reearthx/pgxx"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_ReturnsPoolWhenNoTx(t *testing.T) {
	var pool *pgxpool.Pool // nil pool is fine; we only check identity, not use
	got := pgxx.Executor(context.Background(), pool)
	assert.Equal(t, pgxx.DBTX(pool), got)
}

func TestExecutor_ReturnsTxFromContext(t *testing.T) {
	ctx := pgxx.ContextWithTx(context.Background(), fakeTx{})
	got := pgxx.Executor(ctx, (*pgxpool.Pool)(nil))
	_, isFake := got.(fakeTx)
	assert.True(t, isFake, "Executor must return the tx stored in context")
}

// fakeTx embeds pgx.Tx so it satisfies the interface without implementing every method.
type fakeTx struct{ pgx.Tx }
