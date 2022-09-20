package mongox

import (
	"context"

	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexDocument struct {
	Name   string
	Key    map[string]int
	Unique bool
}

func (c *ClientCollection) CreateIndex(ctx context.Context, keys []string, uniqueKeys []string) []string {
	return lo.Must(c.Indexes(ctx, keys, uniqueKeys))
}

func (c *ClientCollection) Indexes(ctx context.Context, keys []string, uniqueKeys []string) ([]string, error) {
	cur, err := c.client.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}

	var indexes []IndexDocument
	if err := cur.All(ctx, &indexes); err != nil {
		return nil, err
	}

	existingIndexes := map[string]string{}
	existingIndexNames := map[string]struct{}{}
	existingUniqueIndexes := map[string]string{}
	existingUniqueIndexNames := map[string]struct{}{}
	for _, i := range indexes {
		if i.Unique {
			existingUniqueIndexNames[i.Name] = struct{}{}
		} else {
			existingIndexNames[i.Name] = struct{}{}
		}
		for k := range i.Key {
			if i.Unique {
				existingUniqueIndexes[k] = i.Name
			} else {
				existingIndexes[k] = i.Name
			}
		}
	}

	oldIndexes := lo.FilterMap(indexes, func(i IndexDocument, _ int) (string, bool) {
		if i.Name == "_id_" { // default index
			return "", false
		}
		_, ok := existingIndexNames[i.Name]
		return i.Name, ok
	})
	oldUniqueIndexes := lo.FilterMap(indexes, func(i IndexDocument, _ int) (string, bool) {
		if i.Name == "_id_" { // default index
			return "", false
		}
		_, ok := existingUniqueIndexNames[i.Name]
		return i.Name, ok
	})

	for _, name := range oldIndexes {
		if _, err := c.client.Indexes().DropOne(ctx, name); err != nil {
			return nil, err
		}
	}
	for _, name := range oldUniqueIndexes {
		if _, err := c.client.Indexes().DropOne(ctx, name); err != nil {
			return nil, err
		}
	}

	newIndexes := append(
		lo.FilterMap(keys, func(k string, _ int) (mongo.IndexModel, bool) {
			if _, ok := existingIndexes[k]; ok {
				return mongo.IndexModel{}, false
			}

			return mongo.IndexModel{
				Keys: map[string]int{
					k: 1,
				},
				Options: options.Index().SetUnique(false),
			}, true
		}),
		lo.FilterMap(uniqueKeys, func(k string, _ int) (mongo.IndexModel, bool) {
			if _, ok := existingUniqueIndexes[k]; ok {
				return mongo.IndexModel{}, false
			}

			return mongo.IndexModel{
				Keys: map[string]int{
					k: 1,
				},
				Options: options.Index().SetUnique(true),
			}, true
		})...,
	)

	if len(newIndexes) > 0 {
		return c.client.Indexes().CreateMany(ctx, newIndexes)
	}
	return nil, nil
}
