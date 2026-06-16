package pgxx_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
)

func TestMapError(t *testing.T) {
	assert.Nil(t, pgxx.MapError(nil))
	assert.ErrorIs(t, pgxx.MapError(pgx.ErrNoRows), rerror.ErrNotFound)
	assert.ErrorIs(t, pgxx.MapError(&pgconn.PgError{Code: "40001"}), usecasex.ErrTransaction)
	other := errors.New("boom")
	assert.Equal(t, other, pgxx.MapError(other))
}
