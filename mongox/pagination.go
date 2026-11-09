package mongox

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	facetItemsKey = "items"
	facetCountKey = "count"
)

func (c *Collection) Paginate(ctx context.Context, rawFilter any, s *usecasex.Sort, p *usecasex.Pagination, consumer Consumer, opts ...*options.FindOptions) (*usecasex.PageInfo, error) {
	if p == nil || (p.Cursor == nil && p.Offset == nil) {
		return nil, nil
	}

	pFilter, err := c.pageFilter(ctx, *p, s)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, err)
	}

	filter := rawFilter
	if pFilter != nil {
		filter = And(rawFilter, "", pFilter)
	}

	return c.paginate(ctx, rawFilter, s, p, filter, consumer, opts)
}

// PaginateAggregation returns one page of the results of pipeline.
//
// When the pipeline does more than filter (e.g. $lookup/$set before a $match),
// the total and the page's _ids come from a single run of the pipeline and the
// full documents are then loaded by re-running the pipeline on the page's _ids
// only, so expensive stages are not executed twice over the whole result set.
// Pipelines that only filter, or that reorder, group or reshape rows, are paged
// with a direct query plus a count instead.
func (c *Collection) PaginateAggregation(ctx context.Context, pipeline []any, s *usecasex.Sort, p *usecasex.Pagination, consumer Consumer, opts ...*options.AggregateOptions) (*usecasex.PageInfo, error) {
	if p == nil || p.Cursor == nil && p.Offset == nil {
		return nil, nil
	}

	pFilter, pOpts, err := c.aggregateFilter(ctx, *p, s)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, err)
	}

	aOpts := append([]*options.AggregateOptions{pOpts}, opts...)

	if reloadablePipeline(pipeline) {
		res, ok, err := c.paginateAggregationByReload(ctx, pipeline, pFilter, p, aOpts)
		if err != nil {
			return nil, err
		}
		if ok {
			return res.finish(p, consumer)
		}
	}

	res, err := c.paginateAggregationDirect(ctx, pipeline, pFilter, p, aOpts)
	if err != nil {
		return nil, err
	}
	return res.finish(p, consumer)
}

type aggregationPage struct {
	items                  []bson.Raw
	startCursor, endCursor *usecasex.Cursor
	hasMore                bool
	count                  int64
}

func (r aggregationPage) finish(p *usecasex.Pagination, consumer Consumer) (*usecasex.PageInfo, error) {
	if p.Cursor != nil && p.Cursor.Last != nil {
		reverse(r.items)
		r.startCursor, r.endCursor = r.endCursor, r.startCursor
	}

	for _, item := range r.items {
		if err := consumer.Consume(item); err != nil {
			return nil, err
		}
	}

	hasNextPage, hasPreviousPage := pageInfo(p, r.hasMore)
	return usecasex.NewPageInfo(r.count, r.startCursor, r.endCursor, hasNextPage, hasPreviousPage), nil
}

// paginateAggregationDirect appends the page stages to pipeline and counts the
// pipeline separately. A pipeline of $match stages only lets the page's $sort
// use an index.
func (c *Collection) paginateAggregationDirect(ctx context.Context, pipeline, pFilter []any, p *usecasex.Pagination, opts []*options.AggregateOptions) (aggregationPage, error) {
	cursor, err := c.collection.Aggregate(ctx, append(slices.Clone(pipeline), pFilter...), opts...)
	if err != nil {
		return aggregationPage{}, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to find: %w", err))
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	count, err := c.CountAggregation(ctx, pipeline)
	if err != nil {
		return aggregationPage{}, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to count: %w", err))
	}

	items, startCursor, endCursor, hasMore, err := consume(ctx, cursor, limit(*p))
	if err != nil {
		return aggregationPage{}, err
	}
	return aggregationPage{items: items, startCursor: startCursor, endCursor: endCursor, hasMore: hasMore, count: count}, nil
}

// paginateAggregationByReload gets the total and the page's _ids from one $facet
// run, then loads the page's documents by _id. The facet holds at most limit
// small {_id, id} entries, far below the 16MB document limit. It reports false
// if the result rows cannot be identified by _id (missing or repeated _id),
// in which case the caller falls back to the direct query.
func (c *Collection) paginateAggregationByReload(ctx context.Context, pipeline, pFilter []any, p *usecasex.Pagination, opts []*options.AggregateOptions) (aggregationPage, bool, error) {
	facetPipeline := append(slices.Clone(pipeline), bson.M{"$facet": bson.M{
		facetItemsKey: append(slices.Clone(pFilter), bson.M{"$project": bson.M{"_id": 1, idKey: 1}}),
		facetCountKey: []any{bson.M{"$count": "count"}},
	}})
	cursor, err := c.collection.Aggregate(ctx, facetPipeline, opts...)
	if err != nil {
		return aggregationPage{}, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to find: %w", err))
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var facet struct {
		Items []bson.Raw `bson:"items"`
		Count []struct {
			Count int64 `bson:"count"`
		} `bson:"count"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&facet); err != nil {
			return aggregationPage{}, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to decode: %w", err))
		}
	}
	if err := cursor.Err(); err != nil {
		return aggregationPage{}, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to read cursor: %w", err))
	}

	var count int64
	if len(facet.Count) > 0 {
		count = facet.Count[0].Count
	}

	keys, startCursor, endCursor, hasMore, err := consumeRaw(ctx, facet.Items, limit(*p))
	if err != nil {
		return aggregationPage{}, false, err
	}

	ids, ok := uniqueIDs(keys)
	if !ok {
		return aggregationPage{}, false, nil
	}

	items, err := c.aggregateByIDs(ctx, pipeline, ids, opts)
	if err != nil {
		return aggregationPage{}, false, err
	}
	return aggregationPage{items: items, startCursor: startCursor, endCursor: endCursor, hasMore: hasMore, count: count}, true, nil
}

// uniqueIDs returns the _id of every row, or false if a row has no _id or two
// rows share one (e.g. after $unwind), since such rows cannot be reloaded by _id.
func uniqueIDs(rows []bson.Raw) ([]bson.RawValue, bool) {
	ids := make([]bson.RawValue, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		v, err := r.LookupErr("_id")
		if err != nil {
			return nil, false
		}
		k := rawValueKey(v)
		if _, dup := seen[k]; dup {
			return nil, false
		}
		seen[k] = struct{}{}
		ids = append(ids, v)
	}
	return ids, true
}

// rowShapingStages change the order or the number of rows, or replace _id, so
// running them on a subset selected by _id would not reproduce the same rows.
var rowShapingStages = map[string]struct{}{
	"$group": {}, "$sort": {}, "$skip": {}, "$limit": {}, "$sample": {},
	"$bucket": {}, "$bucketAuto": {}, "$sortByCount": {}, "$facet": {},
	"$setWindowFields": {}, "$densify": {}, "$fill": {}, "$unionWith": {},
	"$replaceRoot": {}, "$replaceWith": {}, "$geoNear": {}, "$search": {},
	"$searchMeta": {}, "$vectorSearch": {}, "$documents": {},
}

// reloadablePipeline reports whether the page can be reloaded by _id. A
// pipeline of $match stages only is not worth it: the direct query is cheap
// and its $sort can use an index.
func reloadablePipeline(pipeline []any) bool {
	onlyMatch := true
	for _, stage := range pipeline {
		name, ok := stageName(stage)
		if !ok {
			return false
		}
		if _, shaping := rowShapingStages[name]; shaping {
			return false
		}
		if name != "$match" {
			onlyMatch = false
		}
	}
	return !onlyMatch
}

// stageName returns the operator of a single-key pipeline stage.
func stageName(stage any) (string, bool) {
	var keys []string
	switch st := stage.(type) {
	case bson.M:
		for k := range st {
			keys = append(keys, k)
		}
	case map[string]any:
		for k := range st {
			keys = append(keys, k)
		}
	case bson.D:
		for _, e := range st {
			keys = append(keys, e.Key)
		}
	default:
		return "", false
	}
	if len(keys) != 1 {
		return "", false
	}
	return keys[0], true
}

// aggregateByIDs runs pipeline over the documents whose _id is in ids and
// returns the rows in the order of ids. The selection runs before the
// pipeline; reloadablePipeline ensures the pipeline does not depend on its
// input's order or size.
func (c *Collection) aggregateByIDs(ctx context.Context, pipeline []any, ids []bson.RawValue, opts []*options.AggregateOptions) ([]bson.Raw, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	pPipeline := append([]any{bson.M{"$match": bson.M{"_id": bson.M{"$in": ids}}}}, pipeline...)
	cursor, err := c.collection.Aggregate(ctx, pPipeline, opts...)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to find: %w", err))
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	byID := make(map[string]bson.Raw, len(ids))
	for cursor.Next(ctx) {
		var item bson.Raw
		if err := cursor.Decode(&item); err != nil {
			return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to decode item: %w", err))
		}
		v, err := item.LookupErr("_id")
		if err != nil {
			return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to lookup _id: %w", err))
		}
		byID[rawValueKey(v)] = item
	}
	if err := cursor.Err(); err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to read cursor: %w", err))
	}

	items := make([]bson.Raw, 0, len(ids))
	for _, v := range ids {
		// A document changed or deleted between the two runs is skipped.
		if item, ok := byID[rawValueKey(v)]; ok {
			items = append(items, item)
		}
	}
	return items, nil
}

// rawValueKey returns a map key that identifies a BSON value by type and bytes.
func rawValueKey(v bson.RawValue) string {
	return string(append([]byte{byte(v.Type)}, v.Value...))
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

func consume(ctx context.Context, cursor *mongo.Cursor, limit int64) ([]bson.Raw, *usecasex.Cursor, *usecasex.Cursor, bool, error) {
	var raws []bson.Raw
	for cursor.Next(ctx) {
		var item bson.Raw
		if err := cursor.Decode(&item); err != nil {
			return nil, nil, nil, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to decode item: %w", err))
		}
		raws = append(raws, item)
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, nil, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to read cursor: %w", err))
	}

	return consumeRaw(ctx, raws, limit)
}

// consumeRaw keeps the first limit-1 documents — the extra one is only read to
// learn whether a further page exists.
func consumeRaw(ctx context.Context, raws []bson.Raw, limit int64) ([]bson.Raw, *usecasex.Cursor, *usecasex.Cursor, bool, error) {
	var startCursor, endCursor *usecasex.Cursor
	var items []bson.Raw

	for i, item := range raws {
		if int64(i) >= limit-1 {
			break
		}

		cur, err := getCursor(item)
		if err != nil {
			return nil, nil, nil, false, rerror.ErrInternalByWithContext(ctx, fmt.Errorf("failed to get cursor: %w", err))
		}

		if startCursor == nil {
			startCursor = cur
		}
		endCursor = cur

		items = append(items, item)
	}

	return items, startCursor, endCursor, int64(len(raws)) == limit, nil
}

func reverse(items []bson.Raw) {
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
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

func aggregateOptionsFromPagination(_ usecasex.Pagination, _ *usecasex.Sort) *options.AggregateOptions {
	collation := options.Collation{
		Locale:   "en",
		Strength: 2,
	}
	return options.Aggregate().SetAllowDiskUse(true).SetCollation(&collation)
}

func (c *Collection) pageFilter(ctx context.Context, p usecasex.Pagination, s *usecasex.Sort) (bson.M, error) {
	if p.Cursor == nil {
		return nil, nil
	}

	var filter bson.M
	sortKey := idKey
	sortOrder := 1

	if s != nil && s.Key != "" {
		sortKey = s.Key
		if s.Reverted {
			sortOrder = -1
		}
	}

	var cursor *usecasex.Cursor
	var op string

	if p.Cursor.After != nil {
		cursor = p.Cursor.After
		op = "$gt"
	} else if p.Cursor.Before != nil {
		cursor = p.Cursor.Before
		op = "$lt"
	}

	if cursor != nil {
		cursorDoc, err := c.getCursorDocument(ctx, *cursor)
		if err != nil {
			return nil, err
		}

		filter = bson.M{
			"$or": []bson.M{
				{sortKey: bson.M{op: cursorDoc[sortKey]}},
				{
					sortKey: cursorDoc[sortKey],
					idKey:   bson.M{op: cursorDoc[idKey]},
				},
			},
		}

		if sortOrder == -1 {
			if op == "$gt" {
				op = "$lt"
			} else {
				op = "$gt"
			}
			filter = bson.M{
				"$or": []bson.M{
					{sortKey: bson.M{op: cursorDoc[sortKey]}},
					{
						sortKey: cursorDoc[sortKey],
						idKey:   bson.M{op: cursorDoc[idKey]},
					},
				},
			}
		}
	}

	return filter, nil
}

func (c *Collection) getCursorDocument(ctx context.Context, cursor usecasex.Cursor) (bson.M, error) {
	var cursorDoc bson.M
	err := c.collection.FindOne(ctx, bson.M{idKey: cursor}).Decode(&cursorDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to find cursor element: %w", err)
	}
	return cursorDoc, nil
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

func (c *Collection) PaginateProject(ctx context.Context, rawFilter any, s *usecasex.Sort, p *usecasex.Pagination, consumer Consumer, opts ...*options.FindOptions) (*usecasex.PageInfo, error) {
	if p == nil || (p.Cursor == nil && p.Offset == nil) {
		return nil, nil
	}

	pFilter, err := c.pageFilter(ctx, *p, s)
	if err != nil {
		return nil, rerror.ErrInternalByWithContext(ctx, err)
	}

	filter := rawFilter
	if pFilter != nil {
		filter = AddCondition(rawFilter, "", pFilter)
	}

	return c.paginate(ctx, rawFilter, s, p, filter, consumer, opts)

}

func (c *Collection) paginate(ctx context.Context, rawFilter any, s *usecasex.Sort, p *usecasex.Pagination, filter any, consumer Consumer, opts []*options.FindOptions) (*usecasex.PageInfo, error) {

	sortKey := idKey
	sortOrder := 1
	if s != nil && s.Key != "" {
		sortKey = s.Key
		if s.Reverted {
			sortOrder = -1
		}
	}

	if p.Cursor != nil && p.Cursor.Last != nil {
		sortOrder *= -1
	}

	sort := bson.D{{Key: sortKey, Value: sortOrder}}
	if sortKey != idKey {
		sort = append(sort, bson.E{Key: idKey, Value: sortOrder})
	}

	findOpts := options.Find().
		SetSort(sort).
		SetLimit(limit(*p))

	if p.Offset != nil {
		findOpts.SetSkip(p.Offset.Offset)
	}

	cursor, err := c.collection.Find(ctx, filter, append([]*options.FindOptions{findOpts}, opts...)...)
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

	items, startCursor, endCursor, hasMore, err := consume(ctx, cursor, limit(*p))
	if err != nil {
		return nil, err
	}

	if p.Cursor != nil && p.Cursor.Last != nil {
		reverse(items)
		startCursor, endCursor = endCursor, startCursor
	}

	for _, item := range items {
		if err := consumer.Consume(item); err != nil {
			return nil, err
		}
	}

	hasNextPage, hasPreviousPage := pageInfo(p, hasMore)

	return usecasex.NewPageInfo(count, startCursor, endCursor, hasNextPage, hasPreviousPage), nil
}
