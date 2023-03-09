package mongox

import (
	"context"
	"strings"

	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/slices"
)

// Indexes creates or deletes indexes by keys declaratively
func (c *Collection) Indexes(ctx context.Context, keys, uniqueKeys []string) ([]string, []string, error) {
	cur, err := c.client.Indexes().List(ctx)
	if err != nil {
		return nil, nil, err
	}

	var indexes indexList
	if err := cur.All(ctx, &indexes); err != nil {
		return nil, nil, err
	}
	indexes = indexes.RemoveDefaultIndex()

	existingUniqueIndexes, existingIndexes := indexes.PartionByUnique()
	maintainedIndexes, oldIndexes := existingIndexes.PartionByHasKeys(keys...)
	maintainedUniqueIndexes, oldUniqueIndexes := existingUniqueIndexes.PartionByHasKeys(uniqueKeys...)
	newIndexKeys := lo.Filter(keys, func(k string, _ int) bool { return !maintainedIndexes.HasKey(k) })
	newUniqueIndexKeys := lo.Filter(uniqueKeys, func(k string, _ int) bool { return !maintainedUniqueIndexes.HasKey(k) })

	if err := c.dropIndexes(ctx, append(oldIndexes.Names(), oldUniqueIndexes.Names()...)); err != nil {
		return nil, nil, err
	}

	newIndexes := append(
		// unique
		lo.Map(newUniqueIndexKeys, func(k string, _ int) mongo.IndexModel {
			return newIndexDocument(k, true).Model()
		}),
		// normal
		lo.Map(newIndexKeys, func(k string, _ int) mongo.IndexModel {
			return newIndexDocument(k, false).Model()
		})...,
	)

	if len(newIndexes) > 0 {
		if _, err := c.client.Indexes().CreateMany(ctx, newIndexes); err != nil {
			return nil, nil, err
		}
	}

	added := append(newUniqueIndexKeys, newIndexKeys...)
	deleted := append(oldUniqueIndexes.Keys(), oldIndexes.Keys()...)
	return added, deleted, nil
}

func (c *Collection) dropIndexes(ctx context.Context, indexes []string) error {
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

type indexDocument struct {
	Name   string
	Key    bson.D
	Unique bool
}

func newIndexDocument(key string, unique bool) indexDocument {
	return indexDocument{
		Key:    toKeyBSON(key),
		Unique: unique,
	}
}

func toKeyBSON(key string) bson.D {
	return lo.Map(
		strings.Split(key, ","),
		func(k string, _ int) bson.E {
			return bson.E{
				Key:   strings.TrimSpace(k),
				Value: 1,
			}
		},
	)
}

func (i indexDocument) Model() mongo.IndexModel {
	var o *options.IndexOptions
	if i.Unique {
		o = options.Index().SetUnique(i.Unique)
	}
	return mongo.IndexModel{
		Keys:    i.Key,
		Options: o,
	}
}

func (i indexDocument) HasKey(keys ...string) bool {
	for _, k := range keys {
		if slices.EqualFunc(i.Key, toKeyBSON(k), func(a, b bson.E) bool {
			return a.Key == b.Key
		}) {
			return true
		}
	}
	return false
}

func (i indexDocument) Keys() string {
	return strings.Join(
		lo.Map(i.Key, func(k bson.E, _ int) string {
			return k.Key
		}),
		",",
	)
}

type indexList []indexDocument

func (l indexList) Names() []string {
	return lo.Map(l, func(i indexDocument, _ int) string { return i.Name })
}

func (l indexList) Keys() []string {
	return lo.Map(l, func(i indexDocument, _ int) string { return i.Keys() })
}

func (l indexList) HasKey(key string) bool {
	for _, i := range l {
		if i.HasKey(key) {
			return true
		}
	}
	return false
}

func (l indexList) PartionByHasKeys(keys ...string) (f, t indexList) {
	groups := lo.GroupBy(l, func(i indexDocument) bool {
		return i.HasKey(keys...)
	})
	return groups[true], groups[false]
}

func (l indexList) PartionByUnique() (normal, unique indexList) {
	groups := lo.GroupBy(l, func(i indexDocument) bool { return i.Unique })
	return groups[true], groups[false]
}

func (l indexList) RemoveDefaultIndex() indexList {
	return lo.Filter(l, func(i indexDocument, _ int) bool { return i.Name != "_id_" })
}
