package mongox

import (
	"context"
	"errors"
	"fmt"

	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *ClientCollection) Paginate(ctx context.Context, rawFilter any, sort *string, p *usecasex.Pagination, consumer Consumer) (*usecasex.PageInfo, error) {
	if p == nil || p.Cursor == nil && p.Offset == nil {
		return nil, nil
	}

	filter, findOptions, err := c.paginationFilter(ctx, *p, sort, rawFilter)
	if err != nil {
		return nil, rerror.ErrInternalBy(err)
	}

	limit := int(*findOptions.Limit)
	count, err := c.client.CountDocuments(ctx, rawFilter)
	if err != nil {
		return nil, rerror.ErrInternalBy(fmt.Errorf("failed to count: %w", err))
	}

	cursor, err := c.client.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, rerror.ErrInternalBy(fmt.Errorf("failed to find: %w", err))
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	i := 0
	var startCursor, endCursor *usecasex.Cursor
	for cursor.Next(ctx) {
		if i < limit-1 {
			cur, err := getCursor(cursor.Current)
			if err != nil {
				return nil, rerror.ErrInternalBy(fmt.Errorf("failed to get cursor: %w", err))
			}

			if startCursor == nil {
				startCursor = cur
			}
			endCursor = cur

			if err := consumer.Consume(cursor.Current); err != nil {
				return nil, err
			}
		}

		i++
	}

	if err := cursor.Err(); err != nil {
		return nil, rerror.ErrInternalBy(fmt.Errorf("failed to read cursor: %w", err))
	}

	// ref: https://facebook.github.io/relay/graphql/connections.htm#sec-undefined.PageInfo.Fields
	// If first is set, false can be returned unless it can be efficiently determined whether or not a previous page exists.
	// If last is set, false can be returned unless it can be efficiently determined whether or not a next page exists.
	// Returning absolutely false because the existing implementation cannot determine it efficiently.
	hasMore := i == limit
	hasNextPage := (p.Cursor != nil && p.Cursor.First != nil || p.Offset != nil) && hasMore
	hasPreviousPage := (p.Cursor != nil && p.Cursor.Last != nil) && hasMore

	return usecasex.NewPageInfo(count, startCursor, endCursor, hasNextPage, hasPreviousPage), nil
}

func (c *ClientCollection) paginationFilter(ctx context.Context, p usecasex.Pagination, sortKey *string, filter any) (any, *options.FindOptions, error) {
	opts := findOptionsFromPagination(p, sortKey)

	if p.Offset != nil {
		return filter, opts, nil
	}

	if p.Cursor == nil {
		return nil, nil, errors.New("invalid cursor")
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
		return nil, nil, errors.New("neither first nor last are set")
	}

	var paginationFilter bson.M
	if cur != nil {
		if sortKey == nil || *sortKey == "" {
			paginationFilter = bson.M{idKey: bson.M{op: *cur}}
		} else {
			var cursorDoc bson.M
			if err := c.client.FindOne(ctx, bson.M{idKey: *cur}).Decode(&cursorDoc); err != nil {
				return nil, nil, fmt.Errorf("failed to find cursor element")
			}

			if cursorDoc[*sortKey] == nil {
				return nil, nil, fmt.Errorf("invalied sort key")
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
	}

	return And(filter, "", paginationFilter), opts, nil
}

func findOptionsFromPagination(p usecasex.Pagination, sort *string) *options.FindOptions {
	const defaultLimit = 20
	o := options.Find()

	if p.Offset != nil {
		o = o.SetSkip(p.Offset.Offset).SetLimit(p.Offset.Limit)
	} else if p.Cursor != nil {
		var limit int64
		if p.Cursor.First != nil {
			limit = int64(*p.Cursor.First)
		} else if p.Cursor.Last != nil {
			limit = int64(*p.Cursor.Last)
		}
		o = o.SetLimit(limit)
	}

	if o.Limit == nil || *o.Limit <= 0 {
		o = o.SetLimit(defaultLimit + 1)
	} else {
		// Read one more element so that we can see whether there's a further one
		o = o.SetLimit(*o.Limit + 1)
	}

	// sort
	sortDirection := 1
	if p.Cursor != nil && p.Cursor.Last != nil {
		sortDirection = -1
	}

	var sortOptions bson.D
	if sort != nil && *sort != "" && *sort != idKey {
		sortOptions = append(sortOptions, bson.E{Key: *sort, Value: sortDirection})
	}
	sortOptions = append(sortOptions, bson.E{Key: idKey, Value: sortDirection})

	o = o.SetCollation(&options.Collation{Strength: 1, Locale: "en"}).SetSort(sortOptions)

	return o
}
