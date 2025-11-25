package mongox

import (
	"context"
)

func (c *Collection) createIndexes(ctx context.Context, indexes IndexList) error {
	if len(indexes) == 0 {
		return nil
	}
	_, err := c.collection.Indexes().CreateMany(ctx, indexes.Models())
	return err
}
