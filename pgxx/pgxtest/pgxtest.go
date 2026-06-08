// Package pgxtest provides an env-gated Postgres connection for tests, mirroring
// reearthx/mongox/mongotest: tests skip when no DB URI is configured, and each
// call creates an isolated, uniquely-named database that is dropped on cleanup.
package pgxtest

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Env is the environment variable holding an admin Postgres URI
// (e.g. postgres://user:pass@localhost:5432/postgres?sslmode=disable).
var Env = "REEARTH_DB_PG"

// Connect returns a factory that yields an isolated *pgxpool.Pool per call.
// It t.Skip()s when Env is unset, matching mongotest semantics.
func Connect(t *testing.T) func(*testing.T) *pgxpool.Pool {
	t.Helper()

	adminURI := os.Getenv(Env)
	if adminURI == "" {
		t.Skipf("pgxtest: %s not set; skipping Postgres integration test", Env)
		return nil
	}

	ctx := context.Background()
	admin, err := pgxpool.New(ctx, adminURI)
	if err != nil {
		t.Fatalf("pgxtest: connect admin: %v", err)
	}

	return func(t *testing.T) *pgxpool.Pool {
		t.Helper()

		dbName := "reearth_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		if _, err := admin.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
			t.Fatalf("pgxtest: create database: %v", err)
		}
		t.Cleanup(func() {
			_, _ = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName+" WITH (FORCE)")
		})

		pool, err := pgxpool.New(ctx, replaceDBName(adminURI, dbName))
		if err != nil {
			t.Fatalf("pgxtest: connect test db: %v", err)
		}
		t.Cleanup(pool.Close)
		return pool
	}
}

// replaceDBName swaps the path component (database name) of a Postgres URI.
func replaceDBName(uri, dbName string) string {
	q := ""
	if i := strings.IndexByte(uri, '?'); i >= 0 {
		q = uri[i:]
		uri = uri[:i]
	}
	if i := strings.LastIndexByte(uri, '/'); i >= 0 {
		uri = uri[:i]
	}
	return uri + "/" + dbName + q
}
