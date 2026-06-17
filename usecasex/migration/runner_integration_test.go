package migration_test

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/pgxx/pgxtest"
	"github.com/reearth/reearthx/usecasex/migration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memConfig is an in-memory ConfigRepo for exercising Runner against a real DB.
type memConfig struct{ cur migration.Key }

func (m *memConfig) Begin(context.Context) error                    { return nil }
func (m *memConfig) End(context.Context) error                      { return nil }
func (m *memConfig) Current(context.Context) (migration.Key, error) { return m.cur, nil }
func (m *memConfig) Save(_ context.Context, k migration.Key) error  { m.cur = k; return nil }

// Runner driven by a real pgxx.Client (the deployed Postgres combination):
// migrations run inside real transactions, the version is tracked, and a
// re-run is idempotent.
func TestRunner_RealClient_AppliesAndIdempotent(t *testing.T) {
	pool := pgxtest.Connect(t)(t)
	ctx := context.Background()
	client := pgxx.NewClient(pool)
	cfg := &memConfig{}

	runs := 0
	migs := migration.Migrations[*pgxx.Client]{
		1: func(ctx context.Context, c *pgxx.Client) error {
			runs++
			_, err := c.DB(ctx).Exec(ctx, `CREATE TABLE mig_real (id integer)`)
			return err
		},
	}
	r := migration.NewRunner[*pgxx.Client](client, client, cfg, migs)

	require.NoError(t, r.Migrate(ctx))
	assert.Equal(t, 1, runs)
	assert.Equal(t, migration.Key(1), cfg.cur)

	require.NoError(t, r.Migrate(ctx)) // already at version 1 -> no-op
	assert.Equal(t, 1, runs)
}
