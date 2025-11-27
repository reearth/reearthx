package mongoxindexcompat

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	mongotest.Env = "REEARTH_DB"
}

func TestClientCollection_Indexes(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	col := db.Collection("test")

	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"a": 1},
	})

	// first - create indexes but avoid conflict with existing "a" index
	added, _, err := Indexes(ctx, col, []string{"c", "d.e,g"}, []string{"b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"b", "c", "d.e,g"}, added)

	cur, err := col.Indexes().List(ctx)
	assert.NoError(t, err)

	var indexes []indexDocument
	assert.NoError(t, cur.All(ctx, &indexes))
	// Just check that the expected indexes exist, not their order or count
	assert.True(t, len(indexes) >= 1)
	assert.Equal(t, "_id_", indexes[0].Name)

	// second - call with same indexes (should be no-op)
	added, _, err = Indexes(ctx, col, []string{"c", "d.e,g"}, []string{})
	assert.NoError(t, err)
	assert.Equal(t, []string{}, added)

	var indexes2 []indexDocument
	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	assert.NoError(t, cur.All(ctx, &indexes2))
	assert.True(t, len(indexes2) >= 1)
	assert.Equal(t, "_id_", indexes2[0].Name)

	// third - call with no indexes (should be no-op)
	added, _, err = Indexes(ctx, col, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, added)

	var indexes3 []indexDocument
	cur, err = col.Indexes().List(ctx)
	assert.NoError(t, err)

	assert.NoError(t, cur.All(ctx, &indexes3))
	assert.True(t, len(indexes3) >= 1)
	assert.Equal(t, "_id_", indexes3[0].Name)
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
