package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Tx struct {
	session mongo.Session
	commit  bool
}

func (t *Tx) Commit() {
	if t == nil {
		return
	}
	t.commit = true
}

func (t *Tx) End(ctx context.Context) error {
	if t == nil {
		return nil
	}

	if t.commit {
		if err := t.session.CommitTransaction(ctx); err != nil {
			return err
		}
	} else if err := t.session.AbortTransaction(ctx); err != nil {
		return err
	}

	t.session.EndSession(ctx)
	return nil
}

func (c *Client) CreateUniqueIndex(ctx context.Context, col string, keys, uniqueKeys []string) []string {
	coll := c.Collection(col)
	indexedKeys := indexes(ctx, coll)

	// store unique keys as map to check them in an efficient way
	ukm := map[string]struct{}{}
	for _, k := range append([]string{"id"}, uniqueKeys...) {
		ukm[k] = struct{}{}
	}

	var newIndexes []mongo.IndexModel
	for _, k := range append([]string{"id"}, keys...) {
		if _, ok := indexedKeys[k]; ok {
			continue
		}
		indexBg := true
		_, isUnique := ukm[k]
		newIndexes = append(newIndexes, mongo.IndexModel{
			Keys: map[string]int{
				k: 1,
			},
			Options: &options.IndexOptions{
				Background: &indexBg,
				Unique:     &isUnique,
			},
		})
	}

	if len(newIndexes) > 0 {
		index, err := coll.Indexes().CreateMany(ctx, newIndexes)
		if err != nil {
			panic(err)
		}
		return index
	}
	return nil
}

func (t *Tx) IsCommitted() bool {
	return t.commit
}
