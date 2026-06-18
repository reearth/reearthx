package pgxx_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/reearth/reearthx/pgxx"
	"github.com/stretchr/testify/assert"
)

func TestOrderByIDs(t *testing.T) {
	type item struct {
		id   string
		name string
	}
	items := []*item{{"b", "B"}, {"a", "A"}} // arbitrary fetch order
	key := func(i *item) string { return i.id }

	got := pgxx.OrderByIDs([]string{"a", "missing", "b"}, items, key)
	// Missing ids are omitted (never a nil element), matching Mongo's FindByIDs.
	assert.Len(t, got, 2)
	assert.Equal(t, "A", got[0].name)
	assert.Equal(t, "B", got[1].name)
}

func TestIsUniqueViolation(t *testing.T) {
	assert.True(t, pgxx.IsUniqueViolation(&pgconn.PgError{Code: "23505"}))
	assert.False(t, pgxx.IsUniqueViolation(&pgconn.PgError{Code: "40001"}))
	assert.False(t, pgxx.IsUniqueViolation(nil))
}
