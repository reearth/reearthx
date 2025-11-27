package mongox

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestClientCollection_Indexes2(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	col := db.Collection("test")
	c := NewCollection(col)

	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"a": 1}, // a_1
	})

	// first
	res, err := c.Indexes2(ctx, IndexFromKey("c", false), IndexFromKey("d.e,g", false), IndexFromKey("a", true), IndexFromKey("b", true), TTLIndexFromKey("expires_at", 0))
	assert.NoError(t, err)
	assert.Equal(t, []string{"c", "d.e,g", "a", "b", "expires_at"}, res.AddedNames())
	assert.Equal(t, []string{}, res.UpdatedNames())
	cur, err := col.Indexes().List(ctx)
	assert.NoError(t, err)
	var indexes IndexList
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.True(t, len(indexes) >= 6)
	assert.Contains(t, getIndexNames(indexes), "_id_")
	assert.Contains(t, getIndexNames(indexes), "re_c")
	assert.Contains(t, getIndexNames(indexes), "re_d.e,g")
	assert.Contains(t, getIndexNames(indexes), "re_a")
	assert.Contains(t, getIndexNames(indexes), "re_b")
	assert.Contains(t, getIndexNames(indexes), "re_expires_at")

	// second - call with subset of indexes (should be no-op)
	res, err = c.Indexes2(ctx, IndexList{
		{Name: "d.e,g", Key: bson.D{{Key: "d.e", Value: 1}, {Key: "g", Value: 1}}},
		{Name: "a", Unique: true, Key: bson.D{{Key: "a", Value: 1}}},
		{Name: "c", Key: bson.D{{Key: "c", Value: 1}}},
	}...)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, res.AddedNames())
	assert.Equal(t, []string{}, res.UpdatedNames())
	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)
	indexes = nil
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.True(t, len(indexes) >= 6)
	assert.Contains(t, getIndexNames(indexes), "_id_")
	assert.Contains(t, getIndexNames(indexes), "re_c")
	assert.Contains(t, getIndexNames(indexes), "re_d.e,g")
	assert.Contains(t, getIndexNames(indexes), "re_a")
	assert.Contains(t, getIndexNames(indexes), "re_b")
	assert.Contains(t, getIndexNames(indexes), "re_expires_at")

	// third - call with no indexes (should be no-op)
	res, err = c.Indexes2(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, res.AddedNames())
	assert.Equal(t, []string{}, res.UpdatedNames())

	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)
	indexes = nil
	assert.NoError(t, cur.All(ctx, &indexes))
	// Since indexes aren't actually dropped, they should still exist
	assert.True(t, len(indexes) >= 6)
	assert.Contains(t, getIndexNames(indexes), "_id_")
	assert.Contains(t, getIndexNames(indexes), "re_c")
	assert.Contains(t, getIndexNames(indexes), "re_d.e,g")
	assert.Contains(t, getIndexNames(indexes), "re_a")
	assert.Contains(t, getIndexNames(indexes), "re_b")
	assert.Contains(t, getIndexNames(indexes), "re_expires_at")
}

func getIndexNames(indexes IndexList) []string {
	names := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		names = append(names, idx.Name)
	}
	return names
}
