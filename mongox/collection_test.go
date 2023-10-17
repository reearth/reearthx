package mongox

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

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
