package pgxx_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/reearth/reearthx/pgxx"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
)

func TestIsSerializationError(t *testing.T) {
	assert.True(t, pgxx.IsSerializationError(&pgconn.PgError{Code: "40001"}))
	assert.True(t, pgxx.IsSerializationError(&pgconn.PgError{Code: "40P01"}))
	assert.False(t, pgxx.IsSerializationError(&pgconn.PgError{Code: "23505"}))
	assert.False(t, pgxx.IsSerializationError(errors.New("nope")))
	assert.False(t, pgxx.IsSerializationError(nil))
}

func TestWrapError_SerializationBecomesRetryable(t *testing.T) {
	err := pgxx.WrapError(&pgconn.PgError{Code: "40001"})
	assert.True(t, errors.Is(err, usecasex.ErrTransaction))
}

func TestWrapError_PassesThroughOthers(t *testing.T) {
	orig := errors.New("boom")
	assert.Equal(t, orig, pgxx.WrapError(orig))
	assert.Nil(t, pgxx.WrapError(nil))
}
