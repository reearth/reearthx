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

const idKey = "id"

var findOptions = options.Find().SetAllowDiskUse(true)

type Collection struct {
	client *mongo.Collection
}

func NewCollection(c *mongo.Collection) *Collection {
	return &Collection{client: c}
}

func (c *Collection) Client() *mongo.Collection {
	return c.client
}

func (c *Collection) Find(ctx context.Context, filter any, consumer Consumer) error {
	cursor, err := c.client.Find(ctx, filter, findOptions)
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

func (c *Collection) FindOne(ctx context.Context, filter any, consumer Consumer) error {
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

func (c *Collection) Count(ctx context.Context, filter any) (int64, error) {
	count, err := c.client.CountDocuments(ctx, filter)
	if err != nil {
		return 0, rerror.ErrInternalBy(err)
	}
	return count, nil
}

func (c *Collection) RemoveAll(ctx context.Context, f any) error {
	_, err := c.client.DeleteMany(ctx, f)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (c *Collection) RemoveOne(ctx context.Context, f any) error {
	res, err := c.client.DeleteOne(ctx, f)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	if res != nil && res.DeletedCount == 0 {
		return rerror.ErrNotFound
	}
	return nil
}

func (c *Collection) SaveOne(ctx context.Context, id string, replacement any) error {
	return c.ReplaceOne(ctx, bson.M{idKey: id}, replacement)
}

func (c *Collection) ReplaceOne(ctx context.Context, filter any, replacement any) error {
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

func (c *Collection) SetOne(ctx context.Context, id string, replacement any) error {
	_, err := c.client.UpdateOne(
		ctx,
		bson.M{idKey: id},
		bson.M{"$set": replacement},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (c *Collection) SaveAll(ctx context.Context, ids []string, updates []any) error {
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
			mongo.NewReplaceOneModel().SetFilter(bson.M{idKey: id}).SetReplacement(u).SetUpsert(true),
		)
	}

	_, err := c.client.BulkWrite(ctx, writeModels)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (c *Collection) UpdateMany(ctx context.Context, filter, update any) error {
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

func (c *Collection) UpdateManyMany(ctx context.Context, updates []Update) error {
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

func (c *Collection) BeginTransaction(ctx context.Context) (usecasex.Tx, error) {
	return NewClientWithDatabase(c.client.Database()).BeginTransaction(ctx)
}
