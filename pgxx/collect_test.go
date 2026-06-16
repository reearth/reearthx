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
	assert.Len(t, got, 3)
	assert.Equal(t, "A", got[0].name)
	assert.Nil(t, got[1]) // missing id -> zero value (nil)
	assert.Equal(t, "B", got[2].name)
}

func TestIsUniqueViolation(t *testing.T) {
	assert.True(t, pgxx.IsUniqueViolation(&pgconn.PgError{Code: "23505"}))
	assert.False(t, pgxx.IsUniqueViolation(&pgconn.PgError{Code: "40001"}))
	assert.False(t, pgxx.IsUniqueViolation(nil))
}
