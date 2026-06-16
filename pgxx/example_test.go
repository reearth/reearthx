package pgxx_test

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reearth/reearthx/pgxx"
)

// Example shows the golden-path shape of a Postgres repository built on pgxx:
// hold a *pgxx.Client, resolve the executor with DB(ctx) (which transparently
// joins any ambient transaction), map errors with MapError, and wrap
// multi-write operations in WithinTransaction so they commit/roll back together.
func Example() {
	var pool *pgxpool.Pool // obtained from pgxpool.New(ctx, dsn)

	// Construct once at boot. WithTxRetry opts into serialization-failure retries.
	client := pgxx.NewClient(pool, pgxx.WithTxRetry(2))

	// A read in a repository method — DB(ctx) returns the ambient tx or the pool:
	_ = func(ctx context.Context, id string) error {
		var name string
		err := client.DB(ctx).QueryRow(ctx, `SELECT name FROM things WHERE id = $1`, id).Scan(&name)
		if err != nil {
			return pgxx.MapError(err) // pgx.ErrNoRows -> rerror.ErrNotFound
		}
		return nil
	}

	// An atomic multi-write across repositories that share this Client and ctx:
	_ = client.WithinTransaction(context.Background(), func(ctx context.Context) error {
		// repoA.Save(ctx, ...) and repoB.Save(ctx, ...) both run on the same tx.
		return nil
	})
}
