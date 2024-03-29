package mongox

import (
	"context"
	"errors"
	"fmt"

	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *Collection) Paginate(ctx context.Context, rawFilter any, s *usecasex.Sort, p *usecasex.Pagination, consumer Consumer, opts ...*options.FindOptions) (*usecasex.PageInfo, error) {
	if p == nil || p.Cursor == nil && p.Offset == nil {
		return nil, nil
	}

	pFilter, pOpts, err := c.findFilter(ctx, *p, s)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, err)
	}
	filter := rawFilter
	if pFilter != nil {
		filter = And(rawFilter, "", pFilter)
	}

	cursor, err := c.collection.Find(ctx, filter, append([]*options.FindOptions{pOpts}, opts...)...)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to find: %w", err))
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	count, err := c.collection.CountDocuments(ctx, rawFilter)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to count: %w", err))
	}

	startCursor, endCursor, hasMore, err := consume(ctx, cursor, limit(*p), consumer)
	if err != nil {
		return nil, err
	}

	hasNextPage, hasPreviousPage := pageInfo(p, hasMore)

	return usecasex.NewPageInfo(count, startCursor, endCursor, hasNextPage, hasPreviousPage), nil
}

func (c *Collection) PaginateAggregation(ctx context.Context, pipeline []any, s *usecasex.Sort, p *usecasex.Pagination, consumer Consumer, opts ...*options.AggregateOptions) (*usecasex.PageInfo, error) {
	if p == nil || p.Cursor == nil && p.Offset == nil {
		return nil, nil
	}

	pFilter, pOpts, err := c.aggregateFilter(ctx, *p, s)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, err)
	}

	pPipeline := append(pipeline, pFilter...)

	cursor, err := c.collection.Aggregate(ctx, pPipeline, append([]*options.AggregateOptions{pOpts}, opts...)...)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to find: %w", err))
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	count, err := c.CountAggregation(ctx, pipeline)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to count: %w", err))
	}

	startCursor, endCursor, hasMore, err := consume(ctx, cursor, limit(*p), consumer)
	if err != nil {
		return nil, err
	}

	hasNextPage, hasPreviousPage := pageInfo(p, hasMore)

	return usecasex.NewPageInfo(count, startCursor, endCursor, hasNextPage, hasPreviousPage), nil
}

func pageInfo(p *usecasex.Pagination, hasMore bool) (bool, bool) {
	// ref: https://facebook.github.io/relay/graphql/connections.htm#sec-undefined.PageInfo.Fields
	// If first is set, false can be returned unless it can be efficiently determined whether or not a previous page exists.
	// If last is set, false can be returned unless it can be efficiently determined whether or not a next page exists.
	// Returning absolutely false because the existing implementation cannot determine it efficiently.
	hasNextPage := (p.Cursor != nil && p.Cursor.First != nil || p.Offset != nil) && hasMore
	hasPreviousPage := (p.Cursor != nil && p.Cursor.Last != nil) && hasMore
	return hasNextPage, hasPreviousPage
}

func consume(ctx context.Context, cursor *mongo.Cursor, limit int64, consumer Consumer) (*usecasex.Cursor, *usecasex.Cursor, bool, error) {
	i := int64(0)
	var startCursor, endCursor *usecasex.Cursor
	for cursor.Next(ctx) {
		if i < limit-1 {
			cur, err := getCursor(cursor.Current)
			if err != nil {
				return nil, nil, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to get cursor: %w", err))
			}

			if startCursor == nil {
				startCursor = cur
			}
			endCursor = cur

			if err := consumer.Consume(cursor.Current); err != nil {
				return nil, nil, false, err
			}
		}

		i++
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to read cursor: %w", err))
	}
	return startCursor, endCursor, i == limit, nil
}

func (c *Collection) findFilter(ctx context.Context, p usecasex.Pagination, s *usecasex.Sort) (any, *options.FindOptions, error) {
	if p.Cursor == nil && p.Offset == nil {
		return nil, nil, errors.New("invalid pagination")
	}

	opts := findOptionsFromPagination(p, s)
	if p.Offset != nil {
		return nil, opts, nil
	}
	f, err := c.pageFilter(ctx, p, s)
	return f, opts, err
}

func (c *Collection) aggregateFilter(ctx context.Context, p usecasex.Pagination, s *usecasex.Sort) ([]any, *options.AggregateOptions, error) {
	if p.Cursor == nil && p.Offset == nil {
		return nil, nil, errors.New("invalid pagination")
	}

	stages := []any{bson.M{"$sort": sortFilter(p, s)}}

	if p.Offset != nil {
		stages = append(stages, bson.M{"$skip": p.Offset.Offset})
	}

	f, err := c.pageFilter(ctx, p, s)
	if err != nil {
		return nil, nil, err
	}
	if f != nil {
		stages = append(stages, bson.M{"$match": f})
	}
	return append(stages, bson.M{"$limit": limit(p)}), aggregateOptionsFromPagination(p, s), err
}

func findOptionsFromPagination(p usecasex.Pagination, s *usecasex.Sort) *options.FindOptions {
	o := options.Find().SetAllowDiskUse(true).SetLimit(limit(p))

	if p.Offset != nil {
		o = o.SetSkip(p.Offset.Offset)
	}

	return o.SetCollation(&options.Collation{Strength: 1, Locale: "simple"}).SetSort(sortFilter(p, s))
}

func aggregateOptionsFromPagination(_ usecasex.Pagination, _ *usecasex.Sort) *options.AggregateOptions {
	return options.Aggregate().SetAllowDiskUse(true).SetCollation(&options.Collation{Strength: 1, Locale: "simple"})
}

func (c *Collection) pageFilter(ctx context.Context, p usecasex.Pagination, s *usecasex.Sort) (any, error) {
	if p.Cursor == nil {
		return nil, nil
	}

	var op string
	var cur *usecasex.Cursor

	if p.Cursor.First != nil {
		op = "$gt"
		cur = p.Cursor.After
	} else if p.Cursor.Last != nil {
		op = "$lt"
		cur = p.Cursor.Before
	} else {
		return nil, errors.New("neither first nor last are set")
	}
	if cur == nil {
		return nil, nil
	}

	var sortKey *string
	if s != nil {
		sortKey = &s.Key
	}
	var paginationFilter bson.M
	if sortKey == nil || *sortKey == "" {
		paginationFilter = bson.M{idKey: bson.M{op: *cur}}
	} else {
		var cursorDoc bson.M
		if err := c.collection.FindOne(ctx, bson.M{idKey: *cur}).Decode(&cursorDoc); err != nil {
			return nil, fmt.Errorf("failed to find cursor element")
		}

		if cursorDoc[*sortKey] == nil {
			return nil, fmt.Errorf("invalied sort key")
		}

		paginationFilter = bson.M{
			"$or": []bson.M{
				{*sortKey: bson.M{op: cursorDoc[*sortKey]}},
				{
					*sortKey: cursorDoc[*sortKey],
					idKey:    bson.M{op: *cur},
				},
			},
		}
	}

	return paginationFilter, nil
}

func sortFilter(p usecasex.Pagination, s *usecasex.Sort) bson.D {
	var sortOptions bson.D
	if s != nil && s.Key != "" && s.Key != idKey {
		sortOptions = append(sortOptions, bson.E{Key: s.Key, Value: sortDirection(p, s)})
	}
	return append(sortOptions, bson.E{Key: idKey, Value: sortDirection(p, s)})
}

func limit(p usecasex.Pagination) int64 {
	const defaultLimit = 20
	var limit *int64

	if p.Offset != nil {
		limit = &p.Offset.Limit
	} else if p.Cursor != nil {
		if p.Cursor.First != nil {
			limit = p.Cursor.First
		} else if p.Cursor.Last != nil {
			limit = p.Cursor.Last
		}
	}

	if limit != nil && *limit > 0 {
		// Read one more element so that we can see whether there's a further one
		return *limit + 1
	}

	return defaultLimit + 1
}

func sortDirection(p usecasex.Pagination, s *usecasex.Sort) int {
	reverted := false
	if s != nil {
		reverted = s.Reverted
	}

	reverted = reverted || (p.Cursor != nil && p.Cursor.Last != nil)

	if reverted {
		return -1
	}
	return 1
}
