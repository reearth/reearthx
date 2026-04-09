package mongox

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const idKey = "id"

var (
	defaultFindOptions = []options.Lister[options.FindOptions]{
		options.Find().SetAllowDiskUse(true),
	}

	defaultAggregateOptions = []options.Lister[options.AggregateOptions]{
		options.Aggregate().SetAllowDiskUse(true),
	}
)

type Collection struct {
	collection *mongo.Collection
}

func NewCollection(c *mongo.Collection) *Collection {
	return &Collection{collection: c}
}

func (c *Collection) Client() *mongo.Collection {
	return c.collection
}

func (c *Collection) Find(ctx context.Context, filter any, consumer Consumer, opts ...options.Lister[options.FindOptions]) error {
	cursor, err := c.collection.Find(ctx, filter, append(defaultFindOptions, opts...)...)
	if errors.Is(err, mongo.ErrNilDocument) || errors.Is(err, mongo.ErrNoDocuments) {
		return rerror.ErrNotFound
	}
	if err != nil {
		return wrapError(ctx, err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	for {
		c := cursor.Next(ctx)
		if err := cursor.Err(); err != nil && !errors.Is(err, io.EOF) {
			return wrapError(ctx, err)
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

func (c *Collection) FindOne(ctx context.Context, filter any, consumer Consumer, opts ...options.Lister[options.FindOneOptions]) error {
	raw, err := c.collection.FindOne(ctx, filter, opts...).Raw()
	if err != nil {
		if errors.Is(err, mongo.ErrNilDocument) || errors.Is(err, mongo.ErrNoDocuments) {
			return rerror.ErrNotFound
		}
		return wrapError(ctx, err)
	}
	if err := consumer.Consume(raw); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func (c *Collection) Count(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	count, err := c.collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, wrapError(ctx, err)
	}
	return count, nil
}

func (c *Collection) Aggregate(ctx context.Context, pipeline []any, consumer Consumer, opts ...options.Lister[options.AggregateOptions]) error {
	cursor, err := c.collection.Aggregate(ctx, pipeline, append(defaultAggregateOptions, opts...)...)
	if errors.Is(err, mongo.ErrNilDocument) || errors.Is(err, mongo.ErrNoDocuments) {
		return rerror.ErrNotFound
	}
	if err != nil {
		return wrapError(ctx, err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	for {
		c := cursor.Next(ctx)
		if err := cursor.Err(); err != nil && !errors.Is(err, io.EOF) {
			return wrapError(ctx, err)
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

func (c *Collection) AggregateOne(ctx context.Context, pipeline []any, consumer Consumer, opts ...options.Lister[options.AggregateOptions]) error {
	p := append(pipeline, bson.M{"$limit": 1})
	cursor, err := c.collection.Aggregate(ctx, p, append(defaultAggregateOptions, opts...)...)
	if errors.Is(err, mongo.ErrNilDocument) || errors.Is(err, mongo.ErrNoDocuments) {
		return rerror.ErrNotFound
	}
	if err != nil {
		return wrapError(ctx, err)
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	ok := cursor.Next(ctx)
	if err := cursor.Err(); err != nil && !errors.Is(err, io.EOF) {
		return wrapError(ctx, err)
	}

	if !ok {
		return rerror.ErrNotFound
	}

	if err := consumer.Consume(cursor.Current); err != nil {
		return err
	}

	return nil
}

func (c *Collection) CountAggregation(ctx context.Context, pipeline []any) (int64, error) {
	var result struct {
		Count int64 `bson:"count"`
	}
	p := append(pipeline, bson.M{"$count": "count"})
	cursor, err := c.collection.Aggregate(ctx, p)
	defer func() {
		_ = cursor.Close(ctx)
	}()
	if err != nil {
		return 0, wrapError(ctx, err)
	}
	if !cursor.Next(ctx) {
		return 0, nil
	}
	if err := cursor.Decode(&result); err != nil {
		return 0, wrapError(ctx, err)
	}
	return result.Count, nil
}

func (c *Collection) RemoveAll(ctx context.Context, f any) error {
	_, err := c.collection.DeleteMany(ctx, f)
	if err != nil {
		return wrapError(ctx, err)
	}
	return nil
}

func (c *Collection) RemoveOne(ctx context.Context, f any) error {
	res, err := c.collection.DeleteOne(ctx, f)
	if err != nil {
		return wrapError(ctx, err)
	}
	if res != nil && res.DeletedCount == 0 {
		return rerror.ErrNotFound
	}
	return nil
}

func (c *Collection) CreateOne(ctx context.Context, id string, doc any) error {
	if id == "" {
		return errors.New("id is empty")
	}

	count, err := c.collection.CountDocuments(ctx, bson.M{idKey: id})
	if err != nil {
		return wrapError(ctx, err)
	}

	if count > 0 {
		return rerror.ErrAlreadyExists
	}

	_, err = c.collection.UpdateOne(
		ctx,
		bson.M{idKey: id},
		bson.M{"$setOnInsert": doc},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return wrapError(ctx, err)
	}
	return nil
}

func (c *Collection) SaveOne(ctx context.Context, id string, replacement any) error {
	_, err := c.collection.ReplaceOne(
		ctx,
		bson.M{idKey: id},
		replacement,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		return wrapError(ctx, err)
	}
	return nil
}

func (c *Collection) SetOne(ctx context.Context, id string, replacement any) error {
	_, err := c.collection.UpdateOne(
		ctx,
		bson.M{idKey: id},
		bson.M{"$set": replacement},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return wrapError(ctx, err)
	}
	return nil
}

func (c *Collection) SaveAll(ctx context.Context, ids []string, updates []any) error {
	if len(ids) == 0 || len(updates) == 0 {
		return nil
	}
	if len(ids) != len(updates) {
		return wrapError(ctx, errors.New("invalid save args"))
	}

	writeModels := make([]mongo.WriteModel, 0, len(updates))
	for i, u := range updates {
		id := ids[i]
		writeModels = append(
			writeModels,
			mongo.NewReplaceOneModel().SetFilter(bson.M{idKey: id}).SetReplacement(u).SetUpsert(true),
		)
	}

	_, err := c.collection.BulkWrite(ctx, writeModels)
	if err != nil {
		return wrapError(ctx, err)
	}
	return nil
}

func (c *Collection) UpdateMany(ctx context.Context, filter, update any) error {
	_, err := c.collection.UpdateMany(ctx, filter, bson.M{
		"$set": update,
	})
	if err != nil {
		return wrapError(ctx, err)
	}
	return nil
}

type Update struct {
	Filter       any
	Update       any
	ArrayFilters []any
}

func (c *Collection) UpdateManyMany(ctx context.Context, updates []Update) error {
	writeModels := make([]mongo.WriteModel, 0, len(updates))
	for _, w := range updates {
		wm := mongo.NewUpdateManyModel().SetFilter(w.Filter).SetUpdate(bson.M{
			"$set": w.Update,
		})
		if len(w.ArrayFilters) > 0 {
			wm.SetArrayFilters(w.ArrayFilters)
		}
		writeModels = append(writeModels, wm)
	}

	_, err := c.collection.BulkWrite(ctx, writeModels)
	if err != nil {
		return wrapError(ctx, err)
	}
	return nil
}

func getCursor(raw bson.Raw) (*usecasex.Cursor, error) {
	val, err := raw.LookupErr(idKey)
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

func wrapError(ctx context.Context, err error) error {
	if IsTransactionError(err) {
		log.Errorfc(ctx, "transaction error: %v", err)
		return usecasex.ErrTransaction
	}
	return rerror.ErrInternalByWithContext(ctx, err)
}
