package pgxx

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

// IsSerializationError reports whether err is a Postgres serialization failure
// (40001) or deadlock (40P01) — the cases worth retrying.
func IsSerializationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40001" || pgErr.Code == "40P01"
	}
	return false
}

// IsUniqueViolation reports whether err is a Postgres unique_violation (23505).
// Repositories use it to translate a duplicate-key error into a domain
// "already exists" result.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// WrapError maps a Postgres serialization failure to usecasex.ErrTransaction.
// Other errors (and nil) pass through unchanged.
//
// Call it in the REPOSITORY layer on errors returned from queries, because the
// serialization failure surfaces on the failing statement. The wrap only causes
// a retry when the Transactor is configured to retry — a pgxx.Client created
// with WithTxRetry(n>0), or usecasex.NewTransactor(t, n>0). With a default
// single-attempt Client the wrap affects error classification only, not retry.
func WrapError(err error) error {
	if err == nil {
		return nil
	}
	if IsSerializationError(err) {
		return errors.Join(usecasex.ErrTransaction, err)
	}
	return err
}

// MapError translates common pgx errors into reearthx domain errors for repos:
// no rows -> rerror.ErrNotFound, serialization failure -> usecasex.ErrTransaction
// (retried only when the Transactor is configured to retry; see WrapError),
// everything else passes through. nil -> nil.
//
// It is opt-in: call it only where a missing row should be a not-found error.
// Finds that return (nil, nil) on no rows should test pgx.ErrNoRows themselves
// and skip MapError for that case.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return rerror.ErrNotFound
	}
	return WrapError(err)
}
