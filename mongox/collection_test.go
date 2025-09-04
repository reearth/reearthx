package mongox

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestClientCollection_Count(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	// seeds
	seeds := []string{"a", "A", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": len(seeds) - i}
	}))

	got, err := c.Count(ctx,  bson.M{"id": "a"})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), got)

	got, err = c.Count(ctx,  bson.M{"id": "a"}, options.Count().SetCollation(&options.Collation{
		Locale:   "en",
		Strength: 2,
	}))
	assert.NoError(t, err)
	assert.Equal(t, int64(2), got)

	got, err = c.Count(ctx, bson.M{"id": "x"})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), got)

	got, err = c.Count(ctx, bson.M{})
	assert.NoError(t, err)
	assert.Equal(t, int64(4), got)

	got, err = c.Count(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), got)
}

func TestClientCollection_CountAggregation(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	// seeds
	seeds := []string{"a", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": len(seeds) - i}
	}))

	got, err := c.CountAggregation(ctx, []any{
		bson.M{"$match": bson.M{"id": "a"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), got)

	got, err = c.CountAggregation(ctx, []any{
		bson.M{"$match": bson.M{"id": "x"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), got)

	got, err = c.CountAggregation(ctx, []any{})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), got)

	got, err = c.CountAggregation(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), got)
}

func TestClientCollection_Aggregate(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	// seeds
	seeds := []string{"a", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": i}
	}))

	cons := &SliceConsumer[bson.M]{}
	err := c.Aggregate(ctx, []any{
		bson.M{"$match": bson.M{"id": "a"}},
	}, cons)
	assert.NoError(t, err)
	assert.Equal(t, []bson.M{{"_id": cons.Result[0]["_id"], "id": "a", "i": int32(0)}}, cons.Result)

	// empty result
	cons = &SliceConsumer[bson.M]{}
	err = c.Aggregate(ctx, []any{
		bson.M{"$match": bson.M{"id": "zzz"}},
	}, cons)
	assert.NoError(t, err)
	assert.Nil(t, cons.Result)
}

func TestClientCollection_AggregateOne(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	// seeds
	seeds := []string{"a", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": i}
	}))

	// found
	cons := &SliceConsumer[bson.M]{}
	err := c.AggregateOne(ctx, []any{
		bson.M{"$match": bson.M{"id": "a"}},
	}, cons)
	assert.NoError(t, err)
	assert.Equal(t, []bson.M{{"_id": cons.Result[0]["_id"], "id": "a", "i": int32(0)}}, cons.Result)

	// not found
	cons = &SliceConsumer[bson.M]{}
	err = c.AggregateOne(ctx, []any{
		bson.M{"$match": bson.M{"id": "zzz"}},
	}, cons)
	assert.ErrorIs(t, err, rerror.ErrNotFound)
}
