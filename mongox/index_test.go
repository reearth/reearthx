package mongox

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
)

func hasKey(indexes IndexList, key string) bool {
	for _, idx := range indexes {
		for _, k := range idx.Key {
			if k.Key == key {
				return true
			}
		}
	}
	return false
}

func TestClientCollection_Indexes2(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	col := db.Collection("test")
	c := NewCollection(col)

	// first - create initial indexes
	res, err := c.Indexes(ctx, IndexFromKey("c", false), IndexFromKey("d.e,g", false), IndexFromKey("a", true), IndexFromKey("b", true), TTLIndexFromKey("expires_at", 0))
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"c", "d.e,g", "a", "b", "expires_at"}, res.AddedNames())
	assert.ElementsMatch(t, []string{}, res.UpdatedNames())
	assert.ElementsMatch(t, []string{}, res.DeletedNames())

	cur, err := col.Indexes().List(ctx)
	assert.NoError(t, err)

	var indexes IndexList
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.True(t, hasKey(indexes, "_id"))
	assert.True(t, hasKey(indexes, "c"))
	assert.True(t, hasKey(indexes, "d.e"))
	assert.True(t, hasKey(indexes, "g"))
	assert.True(t, hasKey(indexes, "a"))
	assert.True(t, hasKey(indexes, "b"))
	assert.True(t, hasKey(indexes, "expires_at"))

	// second - create same indexes again (should be no-op)
	res, err = c.Indexes(ctx, IndexFromKey("c", false), IndexFromKey("d.e,g", false), IndexFromKey("a", true), IndexFromKey("b", true), TTLIndexFromKey("expires_at", 0))
	assert.NoError(t, err)
	// No changes expected when creating identical indexes
	assert.ElementsMatch(t, []string{}, res.AddedNames())
	assert.ElementsMatch(t, []string{}, res.UpdatedNames())
	assert.ElementsMatch(t, []string{}, res.DeletedNames())

	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	indexes = nil
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.True(t, hasKey(indexes, "_id"))
	assert.True(t, hasKey(indexes, "c"))
	assert.True(t, hasKey(indexes, "d.e"))
	assert.True(t, hasKey(indexes, "g"))
	assert.True(t, hasKey(indexes, "a"))
	assert.True(t, hasKey(indexes, "b"))
	assert.True(t, hasKey(indexes, "expires_at"))

	// third - create additional index
	res, err = c.Indexes(ctx, IndexFromKey("c", false), IndexFromKey("d.e,g", false), IndexFromKey("a", true), IndexFromKey("b", true), TTLIndexFromKey("expires_at", 0), IndexFromKey("new_field", false))
	assert.NoError(t, err)
	// Only new index should be added
	assert.ElementsMatch(t, []string{"new_field"}, res.AddedNames())
	assert.ElementsMatch(t, []string{}, res.UpdatedNames())
	assert.ElementsMatch(t, []string{}, res.DeletedNames())

	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	indexes = nil
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.True(t, hasKey(indexes, "_id"))
	assert.True(t, hasKey(indexes, "c"))
	assert.True(t, hasKey(indexes, "d.e"))
	assert.True(t, hasKey(indexes, "g"))
	assert.True(t, hasKey(indexes, "a"))
	assert.True(t, hasKey(indexes, "b"))
	assert.True(t, hasKey(indexes, "expires_at"))
	assert.True(t, hasKey(indexes, "new_field"))
}
