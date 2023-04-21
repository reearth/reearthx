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
	assert.Equal(t, []string{"a_1"}, res.DeletedNames())

	cur, err := col.Indexes().List(ctx)
	assert.NoError(t, err)

	var indexes IndexList
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.Equal(t, IndexList{
		{Name: "_id_", Key: bson.D{{Key: "_id", Value: int32(1)}}, Unique: false},
		{Name: "re_c", Key: bson.D{{Key: "c", Value: int32(1)}}, Unique: false},
		{Name: "re_d.e,g", Key: bson.D{{Key: "d.e", Value: int32(1)}, {Key: "g", Value: int32(1)}}, Unique: false},
		{Name: "re_a", Key: bson.D{{Key: "a", Value: int32(1)}}, Unique: true},
		{Name: "re_b", Key: bson.D{{Key: "b", Value: int32(1)}}, Unique: true},
		{Name: "re_expires_at", Key: bson.D{{Key: "expires_at", Value: int32(1)}}, ExpireAfterSeconds: new(int32)},
	}, indexes)

	// second
	res, err = c.Indexes2(ctx, IndexList{
		{Name: "b", Key: bson.D{{Key: "b", Value: 1}}, Filter: bson.M{"f": true}},
		{Name: "d.e,g", Key: bson.D{{Key: "d.e", Value: 1}, {Key: "g", Value: 1}}},
		{Name: "a", Unique: true, Key: bson.D{{Key: "a", Value: 1}}},
	}...)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, res.AddedNames())
	assert.Equal(t, []string{"b"}, res.UpdatedNames())
	assert.Equal(t, []string{"c", "expires_at"}, res.DeletedNames())

	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	indexes = nil
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.Equal(t, IndexList{
		{Name: "_id_", Key: bson.D{{Key: "_id", Value: int32(1)}}, Unique: false},
		{Name: "re_d.e,g", Key: bson.D{{Key: "d.e", Value: int32(1)}, {Key: "g", Value: int32(1)}}, Unique: false},
		{Name: "re_a", Key: bson.D{{Key: "a", Value: int32(1)}}, Unique: true},
		{Name: "re_b", Key: bson.D{{Key: "b", Value: int32(1)}}, Unique: false, Filter: bson.M{"f": true}},
	}, indexes)

	// third
	res, err = c.Indexes2(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, res.AddedNames())
	assert.Equal(t, []string{}, res.UpdatedNames())
	assert.Equal(t, []string{"d.e,g", "a", "b"}, res.DeletedNames())

	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	indexes = nil
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.Equal(t, IndexList{
		{Name: "_id_", Key: bson.D{{Key: "_id", Value: int32(1)}}, Unique: false},
	}, indexes)
}
