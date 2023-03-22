package mongox

import (
	"context"

	"github.com/reearth/reearthx/mongox/mongoxindexcompat"
	"github.com/reearth/reearthx/util"
)

// Indexes creates and deletes indexes by keys declaratively
func (c *Collection) Indexes(ctx context.Context, keys, uniqueKeys []string) ([]string, []string, error) {
	return mongoxindexcompat.Indexes(ctx, c.client, keys, uniqueKeys)
}

// Indexes creates and deletes indexes declaratively
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

	oldIndexNames := append(
		IndexList(diff.Deleted).Names(),
		IndexList(diff.UpdatedPrev()).Names()...,
	)
	createdIndexes := append(diff.UpdatedNext(), diff.Added...)

	if err := c.dropIndexes(ctx, oldIndexNames); err != nil {
		return IndexResult{}, err
	}

	if err := c.createIndexes(ctx, createdIndexes); err != nil {
		return IndexResult{}, err
	}

	return IndexResult(diff), nil
}

func (c *Collection) findIndexes(ctx context.Context) (IndexList, error) {
	cur, err := c.client.Indexes().List(ctx)
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
	_, err := c.client.Indexes().CreateMany(ctx, indexes.Models())
	return err
}

func (c *Collection) dropIndexes(ctx context.Context, indexes []string) error {
	if len(indexes) == 0 {
		return nil
	}
	for _, name := range indexes {
		if name == "_id_" {
			continue // cannot drop _id index
		}
		if _, err := c.client.Indexes().DropOne(ctx, name); err != nil {
			return err
		}
	}
	return nil
}
