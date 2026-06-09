// Package pgxx provides reusable PostgreSQL building blocks for reearth services:
// a pgx-backed usecasex.Transactor, an executor-from-context helper that lets
// repositories transparently participate in a transaction, and error helpers.
package pgxx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the minimal query surface shared by *pgxpool.Pool and pgx.Tx.
// It matches the interface sqlc generates for the pgx/v5 driver, so a value of
// this type is assignable to a generated package's DBTX parameter.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type txKey struct{}

// ContextWithTx returns a copy of ctx carrying tx. Used by Transaction.Begin;
// exported so tests (and advanced callers) can inject a transaction.
func ContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// Executor returns the active transaction stored in ctx if present, otherwise
// the supplied db (typically a *pgxpool.Pool). Repositories build their sqlc
// Queries with Executor(ctx, pool) so writes inside a usecasex.Transactor run
// on the transaction's connection automatically.
func Executor(ctx context.Context, db DBTX) DBTX {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return db
}
