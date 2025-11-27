package mongox

import (
	"context"

	"github.com/reearth/reearthx/mongox/mongoxindexcompat"
	"github.com/reearth/reearthx/util"
)

// Indexes creates indexes by keys declaratively
func (c *Collection) Indexes(ctx context.Context, keys, uniqueKeys []string) ([]string, []string, error) {
	return mongoxindexcompat.Indexes(ctx, c.collection, keys, uniqueKeys)
}

// Indexes creates indexes declaratively
func (c *Collection) Indexes2(ctx context.Context, inputs ...Index) (IndexResult, error) {
	inputIndexes := IndexList(inputs).AddNamePrefix().Normalize()
	indexes, err := c.findIndexes(ctx)
	if err != nil {
		return IndexResult{}, err
	}

	diff := util.Diff(
		indexes,
		inputIndexes,
		func(a, b Index) bool { return a.Name == b.Name },
		func(a, b Index) bool { return !a.Equal(b) },
	)

	createdIndexes := append(diff.UpdatedNext(), diff.Added...)

	if err := c.createIndexes(ctx, createdIndexes); err != nil {
		return IndexResult{}, err
	}

	return IndexResult(diff), nil
}

func (c *Collection) findIndexes(ctx context.Context) (IndexList, error) {
	cur, err := c.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}

	var indexes IndexList
	if err := cur.All(ctx, &indexes); err != nil {
		return nil, err
	}
	indexes = indexes.RemoveDefaultIndex()

	return indexes, nil
}

func (c *Collection) createIndexes(ctx context.Context, indexes IndexList) error {
	if len(indexes) == 0 {
		return nil
	}
	_, err := c.collection.Indexes().CreateMany(ctx, indexes.Models())
	return err
}
