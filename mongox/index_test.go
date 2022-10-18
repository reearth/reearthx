package mongox

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestClientCollection_Indexes(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	col := db.Collection("test")
	c := NewClientCollection(col)

	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"d": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"c": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"a": 1,
		},
	})
	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"f": 1,
		},
	})

	_, err := c.Indexes(ctx, []string{"d", "e", "f"}, []string{"a", "b", "c"})
	assert.NoError(t, err)

	cur, err := col.Indexes().List(ctx)
	assert.NoError(t, err)
	var indexes []IndexDocument
	assert.NoError(t, cur.All(ctx, &indexes))
	if assert.Equal(t, 7, len(indexes)) {
		assert.Equal(t, []IndexDocument{
			{Name: indexes[0].Name, Key: map[string]int{"_id": 1}, Unique: false},
			{Name: indexes[1].Name, Key: map[string]int{"d": 1}, Unique: false},
			{Name: indexes[2].Name, Key: map[string]int{"e": 1}, Unique: false},
			{Name: indexes[3].Name, Key: map[string]int{"f": 1}, Unique: false},
			{Name: indexes[4].Name, Key: map[string]int{"a": 1}, Unique: true},
			{Name: indexes[5].Name, Key: map[string]int{"b": 1}, Unique: true},
			{Name: indexes[6].Name, Key: map[string]int{"c": 1}, Unique: true},
		}, indexes)
	}
}
