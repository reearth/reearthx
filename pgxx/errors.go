package pgxx

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
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

// WrapError maps a Postgres serialization failure to usecasex.ErrTransaction so
// the existing usecasex.DoTransaction retry loop picks it up. Other errors
// (and nil) pass through unchanged.
func WrapError(err error) error {
	if err == nil {
		return nil
	}
	if IsSerializationError(err) {
		return errors.Join(usecasex.ErrTransaction, err)
	}
	return err
}
