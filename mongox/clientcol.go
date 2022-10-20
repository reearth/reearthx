package mongox

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientCollection struct {
	client *mongo.Collection
}

func NewClientCollection(c *mongo.Collection) *ClientCollection {
	return &ClientCollection{client: c}
}

func (c *ClientCollection) Client() *mongo.Collection {
	return c.client
}

func (c *ClientCollection) Find(ctx context.Context, filter any, consumer Consumer) error {
	cursor, err := c.client.Find(ctx, filter)
	if errors.Is(err, mongo.ErrNilDocument) || errors.Is(err, mongo.ErrNoDocuments) {
		return rerror.ErrNotFound
	}
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	for {
		c := cursor.Next(ctx)
		if err := cursor.Err(); err != nil && !errors.Is(err, io.EOF) {
			return rerror.ErrInternalBy(err)
		}

		if !c {
			if err := consumer.Consume(nil); err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			break
		}

		if err := consumer.Consume(cursor.Current); err != nil {
			return err
		}
	}
	return nil
}

func (c *ClientCollection) FindOne(ctx context.Context, filter any, consumer Consumer) error {
	raw, err := c.client.FindOne(ctx, filter).DecodeBytes()
	if err != nil {
		if errors.Is(err, mongo.ErrNilDocument) || errors.Is(err, mongo.ErrNoDocuments) {
			return rerror.ErrNotFound
		}
		return rerror.ErrInternalBy(err)
	}
	if err := consumer.Consume(raw); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func (c *ClientCollection) Count(ctx context.Context, filter any) (int64, error) {
	count, err := c.client.CountDocuments(ctx, filter)
	if err != nil {
		return 0, rerror.ErrInternalBy(err)
	}
	return count, nil
}

func (c *ClientCollection) RemoveAll(ctx context.Context, f any) error {
	_, err := c.client.DeleteMany(ctx, f)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (c *ClientCollection) RemoveOne(ctx context.Context, f any) error {
	res, err := c.client.DeleteOne(ctx, f)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	if res != nil && res.DeletedCount == 0 {
		return rerror.ErrNotFound
	}
	return nil
}

func (c *ClientCollection) SaveOne(ctx context.Context, id string, replacement any) error {
	return c.ReplaceOne(ctx, bson.M{"id": id}, replacement)
}

func (c *ClientCollection) ReplaceOne(ctx context.Context, filter any, replacement any) error {
	_, err := c.client.ReplaceOne(
		ctx,
		filter,
		replacement,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (c *ClientCollection) SetOne(ctx context.Context, id string, replacement any) error {
	_, err := c.client.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": replacement},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (c *ClientCollection) SaveAll(ctx context.Context, ids []string, updates []any) error {
	if len(ids) == 0 || len(updates) == 0 {
		return nil
	}
	if len(ids) != len(updates) {
		return rerror.ErrInternalBy(errors.New("invalid save args"))
	}

	writeModels := make([]mongo.WriteModel, 0, len(updates))
	for i, u := range updates {
		id := ids[i]
		writeModels = append(
			writeModels,
			mongo.NewReplaceOneModel().SetFilter(bson.M{"id": id}).SetReplacement(u).SetUpsert(true),
		)
	}

	_, err := c.client.BulkWrite(ctx, writeModels)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (c *ClientCollection) UpdateMany(ctx context.Context, filter, update any) error {
	_, err := c.client.UpdateMany(ctx, filter, bson.M{
		"$set": update,
	})
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

type Update struct {
	Filter       any
	Update       any
	ArrayFilters []any
}

func (c *ClientCollection) UpdateManyMany(ctx context.Context, updates []Update) error {
	writeModels := make([]mongo.WriteModel, 0, len(updates))
	for _, w := range updates {
		wm := mongo.NewUpdateManyModel().SetFilter(w.Filter).SetUpdate(bson.M{
			"$set": w.Update,
		})
		if len(w.ArrayFilters) > 0 {
			wm.SetArrayFilters(options.ArrayFilters{
				Filters: w.ArrayFilters,
			})
		}
		writeModels = append(writeModels, wm)
	}

	_, err := c.client.BulkWrite(ctx, writeModels)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func getCursor(raw bson.Raw, key string) (*usecasex.Cursor, error) {
	val, err := raw.LookupErr(key)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup cursor: %v", err.Error())
	}
	var s string
	if err := val.Unmarshal(&s); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %v", err.Error())
	}
	c := usecasex.Cursor(s)
	return &c, nil
}

func (c *ClientCollection) Paginate(ctx context.Context, filter any, sort *string, p *usecasex.Pagination, consumer Consumer) (*usecasex.PageInfo, error) {
	if p == nil {
		return nil, nil
	}

	const key = "id"
	pa := PaginationFrom(p)
	findOptions := options.Find().SetCollation(&options.Collation{Strength: 1, Locale: "en"})
	sortOptions, sortKey := pa.SortOptions(sort, key)
	findOptions.Sort = sortOptions

	count, err := c.client.CountDocuments(ctx, filter)
	if err != nil {
		return nil, rerror.ErrInternalBy(fmt.Errorf("failed to count documents: %v", err.Error()))
	}

	filter, limit, err := c.paginationFilter(ctx, pa, sortKey, key, filter)
	if err != nil {
		return nil, rerror.ErrInternalBy(err)
	}

	// Read one more element so that we can see whether there's a further one
	limit++
	findOptions.Limit = &limit

	cursor, err := c.client.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, rerror.ErrInternalBy(fmt.Errorf("failed to find: %v", err.Error()))
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	results := make([]bson.Raw, 0, limit)
	for cursor.Next(ctx) {
		raw := make(bson.Raw, len(cursor.Current))
		copy(raw, cursor.Current)
		results = append(results, raw)
	}
	if err := cursor.Err(); err != nil {
		return nil, rerror.ErrInternalBy(fmt.Errorf("failed to read cursor: %v", err.Error()))
	}

	hasMore := false
	if len(results) == int(limit) {
		hasMore = true
		// Remove the extra one reading.
		results = results[:len(results)-1]
	}

	for _, result := range results {
		if err := consumer.Consume(result); err != nil {
			return nil, err
		}
	}

	var startCursor, endCursor *usecasex.Cursor
	if len(results) > 0 {
		sc, err := getCursor(results[0], key)
		if err != nil {
			return nil, rerror.ErrInternalBy(fmt.Errorf("failed to get start cursor: %v", err.Error()))
		}
		startCursor = sc
		ec, err := getCursor(results[len(results)-1], key)
		if err != nil {
			return nil, rerror.ErrInternalBy(fmt.Errorf("failed to get end cursor: %v", err.Error()))
		}
		endCursor = ec
	}

	// ref: https://facebook.github.io/relay/graphql/connections.htm#sec-undefined.PageInfo.Fields
	// If first is set, false can be returned unless it can be efficiently determined whether or not a previous page exists.
	// If last is set, false can be returned unless it can be efficiently determined whether or not a next page exists.
	// Returning absolutely false because the existing implementation cannot determine it efficiently.
	var hasNextPage, hasPreviousPage bool
	switch {
	case p.First != nil:
		hasNextPage = hasMore
	case p.Last != nil:
		hasPreviousPage = hasMore
	}

	return usecasex.NewPageInfo(int(count), startCursor, endCursor, hasNextPage, hasPreviousPage), nil
}

func (c *ClientCollection) BeginTransaction() (usecasex.Tx, error) {
	return NewClientWithDatabase(c.client.Database()).BeginTransaction()
}

func (c *ClientCollection) paginationFilter(ctx context.Context, p *pagination, sortKey, key string, filter any) (any, int64, error) {
	limit, op, cur, err := p.Parameters()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to parse pagination parameters: %w", err)
	}

	var paginationFilter bson.M

	if cur != nil {
		if sortKey == "" {
			paginationFilter = bson.M{key: bson.M{op: *cur}}
		} else {
			var curObj bson.M
			if err := c.client.FindOne(ctx, bson.M{key: *cur}).Decode(&curObj); err != nil {
				return nil, 0, fmt.Errorf("failed to find cursor element")
			}
			if curObj[sortKey] == nil {
				return nil, 0, fmt.Errorf("invalied sort key")
			}
			paginationFilter = bson.M{
				"$or": []bson.M{
					{sortKey: bson.M{op: curObj[sortKey]}},
					{
						sortKey: curObj[sortKey],
						key:     bson.M{op: *cur},
					},
				},
			}
		}
	}

	return And(
		filter,
		"",
		paginationFilter,
	), limit, nil
}
