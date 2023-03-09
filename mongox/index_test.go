package mongox

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestClientCollection_Indexes(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	col := db.Collection("test")
	c := NewCollection(col)

	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"a": 1},
	})

	// first
	added, deleted, err := c.Indexes(ctx, []string{"c", "d.e,g"}, []string{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c", "d.e,g"}, added)
	assert.Equal(t, []string{"a"}, deleted)

	cur, err := col.Indexes().List(ctx)
	assert.NoError(t, err)

	var indexes []indexDocument
	assert.NoError(t, cur.All(ctx, &indexes))
	assert.Equal(t, []indexDocument{
		{Name: indexes[0].Name, Key: bson.D{{Key: "_id", Value: int32(1)}}, Unique: false},
		{Name: indexes[1].Name, Key: bson.D{{Key: "a", Value: int32(1)}}, Unique: true},
		{Name: indexes[2].Name, Key: bson.D{{Key: "b", Value: int32(1)}}, Unique: true},
		{Name: indexes[3].Name, Key: bson.D{{Key: "c", Value: int32(1)}}, Unique: false},
		{Name: indexes[4].Name, Key: bson.D{{Key: "d.e", Value: int32(1)}, {Key: "g", Value: int32(1)}}, Unique: false},
	}, indexes)

	// second
	added, deleted, err = c.Indexes(ctx, []string{"b", "d.e,g"}, []string{"a"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"b"}, added)
	assert.Equal(t, []string{"b", "c"}, deleted)

	var indexes2 []indexDocument
	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	assert.NoError(t, cur.All(ctx, &indexes2))
	assert.Equal(t, []indexDocument{
		{Name: indexes2[0].Name, Key: bson.D{{Key: "_id", Value: int32(1)}}, Unique: false},
		{Name: indexes2[1].Name, Key: bson.D{{Key: "a", Value: int32(1)}}, Unique: true},
		{Name: indexes2[2].Name, Key: bson.D{{Key: "d.e", Value: int32(1)}, {Key: "g", Value: int32(1)}}, Unique: false},
		{Name: indexes2[3].Name, Key: bson.D{{Key: "b", Value: int32(1)}}, Unique: false},
	}, indexes2)

	// thrid
	added, deleted, err = c.Indexes(ctx, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, added)
	assert.Equal(t, []string{"a", "d.e,g", "b"}, deleted)

	var indexes3 []indexDocument
	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	assert.NoError(t, cur.All(ctx, &indexes3))
	assert.Equal(t, []indexDocument{
		{Name: indexes3[0].Name, Key: bson.D{{Key: "_id", Value: int32(1)}}, Unique: false},
	}, indexes3)
}

func TestToKeyBSON(t *testing.T) {
	assert.Equal(
		t,
		bson.D{{Key: "a", Value: 1}},
		toKeyBSON("a"),
	)
	assert.Equal(
		t,
		bson.D{{Key: "b.a", Value: 1}, {Key: "b", Value: 1}},
		toKeyBSON("b.a,b"),
	)
}

func TestIndexDocument_HasKeys(t *testing.T) {
	assert.True(t,
		indexDocument{
			Key: bson.D{{Key: "a.b", Value: 1}, {Key: "c", Value: 1}},
		}.HasKey("a.b,c"))
	assert.False(t,
		indexDocument{
			Key: bson.D{{Key: "a.b", Value: 1}, {Key: "c", Value: 1}},
		}.HasKey("c,a.b"))
	assert.False(t,
		indexDocument{
			Key: bson.D{{Key: "b.a", Value: 1}},
		}.HasKey("a.b"))
}

func TestIndexList_PartionByHasKeys(t *testing.T) {
	i1 := indexDocument{Name: "1", Key: bson.D{{Key: "a", Value: int32(1)}}}
	i2 := indexDocument{Name: "2", Key: bson.D{{Key: "b", Value: int32(1)}}}
	i3 := indexDocument{Name: "3", Key: bson.D{{Key: "c", Value: int32(1)}, {Key: "d", Value: 1}}}

	tr, fa := indexList{i1, i2, i3}.PartionByHasKeys("a", "c,d", "e")
	assert.Equal(t, indexList{i1, i3}, tr)
	assert.Equal(t, indexList{i2}, fa)
}

func TestIndexList_PartionByUnique(t *testing.T) {
	i1 := indexDocument{Name: "1", Unique: false}
	i2 := indexDocument{Name: "2", Unique: true}
	i3 := indexDocument{Name: "3", Unique: false}

	tr, fa := indexList{i1, i2, i3}.PartionByUnique()
	assert.Equal(t, indexList{i2}, tr)
	assert.Equal(t, indexList{i1, i3}, fa)
}
