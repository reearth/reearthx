package mongox

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/reearth/reearthx/usecasex"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestClientCollection_Paginate(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	seeds := []string{"a", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": len(seeds) - i}
	}))

	got, goterr := c.Paginate(ctx, nil, nil, nil, nil)
	assert.Nil(t, got)
	assert.NoError(t, goterr)

	p := usecasex.CursorPagination{
		First: lo.ToPtr(int64(1)),
	}

	con := &consumer{}
	got, goterr = c.Paginate(ctx, bson.M{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("a").Ref(),
		EndCursor:       usecasex.Cursor("a").Ref(),
		HasNextPage:     true,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"a"}, con.Cursors)

	p = usecasex.CursorPagination{
		First: lo.ToPtr(int64(1)),
		After: usecasex.Cursor("b").Ref(),
	}

	con = &consumer{}
	got, goterr = c.Paginate(ctx, bson.M{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("c").Ref(),
		EndCursor:       usecasex.Cursor("c").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"c"}, con.Cursors)

	p = usecasex.CursorPagination{
		Last: lo.ToPtr(int64(1)),
	}

	con = &consumer{}
	got, goterr = c.Paginate(ctx, bson.M{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("c").Ref(),
		EndCursor:       usecasex.Cursor("c").Ref(),
		HasNextPage:     false,
		HasPreviousPage: true,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"c"}, con.Cursors)

	p = usecasex.CursorPagination{
		Last:   lo.ToPtr(int64(1)),
		Before: usecasex.Cursor("b").Ref(),
	}

	con = &consumer{}
	got, goterr = c.Paginate(ctx, bson.M{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("a").Ref(),
		EndCursor:       usecasex.Cursor("a").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"a"}, con.Cursors)

	op := usecasex.OffsetPagination{
		Offset: int64(1),
		Limit:  int64(2),
	}

	con = &consumer{}
	got, goterr = c.Paginate(ctx, bson.M{}, nil, op.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("b").Ref(),
		EndCursor:       usecasex.Cursor("c").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"b", "c"}, con.Cursors)

	op = usecasex.OffsetPagination{
		Offset: int64(0),
		Limit:  int64(0),
	}

	con = &consumer{}
	got, goterr = c.Paginate(ctx, bson.M{}, &usecasex.Sort{
		Key: "i",
	}, op.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("c").Ref(),
		EndCursor:       usecasex.Cursor("a").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"c", "b", "a"}, con.Cursors)
}

func TestClientCollection_PaginateAggregation(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	seeds := []string{"a", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": len(seeds) - i}
	}))

	got, goterr := c.PaginateAggregation(ctx, nil, nil, nil, nil)
	assert.Nil(t, got)
	assert.NoError(t, goterr)

	p := usecasex.CursorPagination{
		First: lo.ToPtr(int64(1)),
	}

	con := &consumer{}
	got, goterr = c.PaginateAggregation(ctx, []any{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("a").Ref(),
		EndCursor:       usecasex.Cursor("a").Ref(),
		HasNextPage:     true,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"a"}, con.Cursors)

	p = usecasex.CursorPagination{
		First: lo.ToPtr(int64(1)),
		After: usecasex.Cursor("b").Ref(),
	}

	con = &consumer{}
	got, goterr = c.PaginateAggregation(ctx, []any{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("c").Ref(),
		EndCursor:       usecasex.Cursor("c").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"c"}, con.Cursors)

	p = usecasex.CursorPagination{
		Last: lo.ToPtr(int64(1)),
	}

	con = &consumer{}
	got, goterr = c.PaginateAggregation(ctx, []any{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("c").Ref(),
		EndCursor:       usecasex.Cursor("c").Ref(),
		HasNextPage:     false,
		HasPreviousPage: true,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"c"}, con.Cursors)

	p = usecasex.CursorPagination{
		Last:   lo.ToPtr(int64(1)),
		Before: usecasex.Cursor("b").Ref(),
	}

	con = &consumer{}
	got, goterr = c.PaginateAggregation(ctx, []any{}, nil, p.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("a").Ref(),
		EndCursor:       usecasex.Cursor("a").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"a"}, con.Cursors)

	op := usecasex.OffsetPagination{
		Offset: int64(1),
		Limit:  int64(2),
	}

	con = &consumer{}
	got, goterr = c.PaginateAggregation(ctx, []any{}, nil, op.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("b").Ref(),
		EndCursor:       usecasex.Cursor("c").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"b", "c"}, con.Cursors)

	op = usecasex.OffsetPagination{
		Offset: int64(0),
		Limit:  int64(0),
	}

	con = &consumer{}
	got, goterr = c.PaginateAggregation(ctx, []any{}, &usecasex.Sort{
		Key: "i",
	}, op.Wrap(), con)
	assert.Equal(t, &usecasex.PageInfo{
		TotalCount:      3,
		StartCursor:     usecasex.Cursor("c").Ref(),
		EndCursor:       usecasex.Cursor("a").Ref(),
		HasNextPage:     false,
		HasPreviousPage: false,
	}, got)
	assert.NoError(t, goterr)
	assert.Equal(t, []usecasex.Cursor{"c", "b", "a"}, con.Cursors)
}

// A last page holding more than one item tells whether start and end cursors
// are swapped back after the items are reversed; a single-item page can't.
func TestClientCollection_PaginateLastCursors(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	seeds := []string{"a", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": len(seeds) - i}
	}))

	paginators := []struct {
		name     string
		paginate func(*usecasex.Pagination, Consumer) (*usecasex.PageInfo, error)
	}{
		{
			name: "Paginate",
			paginate: func(p *usecasex.Pagination, con Consumer) (*usecasex.PageInfo, error) {
				return c.Paginate(ctx, bson.M{}, nil, p, con)
			},
		},
		{
			name: "PaginateAggregation",
			paginate: func(p *usecasex.Pagination, con Consumer) (*usecasex.PageInfo, error) {
				return c.PaginateAggregation(ctx, []any{}, nil, p, con)
			},
		},
	}

	tests := []struct {
		name        string
		pagination  usecasex.CursorPagination
		wantInfo    *usecasex.PageInfo
		wantCursors []usecasex.Cursor
	}{
		{
			name:       "last 2",
			pagination: usecasex.CursorPagination{Last: new(int64(2))},
			wantInfo: &usecasex.PageInfo{
				TotalCount:      3,
				StartCursor:     usecasex.Cursor("b").Ref(),
				EndCursor:       usecasex.Cursor("c").Ref(),
				HasNextPage:     false,
				HasPreviousPage: true,
			},
			wantCursors: []usecasex.Cursor{"b", "c"},
		},
		{
			name: "last 2 before c",
			pagination: usecasex.CursorPagination{
				Last:   new(int64(2)),
				Before: usecasex.Cursor("c").Ref(),
			},
			wantInfo: &usecasex.PageInfo{
				TotalCount:      3,
				StartCursor:     usecasex.Cursor("a").Ref(),
				EndCursor:       usecasex.Cursor("b").Ref(),
				HasNextPage:     false,
				HasPreviousPage: false,
			},
			wantCursors: []usecasex.Cursor{"a", "b"},
		},
	}

	for _, pg := range paginators {
		for _, tt := range tests {
			t.Run(pg.name+"/"+tt.name, func(t *testing.T) {
				con := &consumer{}
				got, err := pg.paginate(tt.pagination.Wrap(), con)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantInfo, got)
				assert.Equal(t, tt.wantCursors, con.Cursors)
			})
		}
	}
}

func TestClientCollection_PaginateWithUpdatedAtSort(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	seeds := []struct {
		id        string
		updatedAt int64
	}{
		{"a", 1000},
		{"b", 2000},
		{"c", 3000},
		{"d", 4000},
		{"e", 5000},
	}

	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s struct {
		id        string
		updatedAt int64
	}, i int) any {
		return bson.M{"id": s.id, "updatedAt": s.updatedAt}
	}))

	sortOpt := &usecasex.Sort{Key: "updatedAt", Reverted: false}

	p := usecasex.CursorPagination{
		First: lo.ToPtr(int64(2)),
	}

	con := &consumer{}
	_, err := c.Paginate(ctx, bson.M{}, sortOpt, p.Wrap(), con)
	assert.NoError(t, err)
	assert.Equal(t, []usecasex.Cursor{"a", "b"}, con.Cursors)

	p = usecasex.CursorPagination{
		Last: lo.ToPtr(int64(2)),
	}

	con = &consumer{}
	_, err = c.Paginate(ctx, bson.M{}, sortOpt, p.Wrap(), con)
	assert.NoError(t, err)
	assert.Equal(t, []usecasex.Cursor{"d", "e"}, con.Cursors)

	p = usecasex.CursorPagination{
		First: lo.ToPtr(int64(2)),
		After: usecasex.Cursor("b").Ref(),
	}

	con = &consumer{}
	_, err = c.Paginate(ctx, bson.M{}, sortOpt, p.Wrap(), con)
	assert.NoError(t, err)
	assert.Equal(t, []usecasex.Cursor{"c", "d"}, con.Cursors)

	p = usecasex.CursorPagination{
		Last:   lo.ToPtr(int64(2)),
		Before: usecasex.Cursor("d").Ref(),
	}

	con = &consumer{}
	_, err = c.Paginate(ctx, bson.M{}, sortOpt, p.Wrap(), con)
	assert.NoError(t, err)
	assert.Equal(t, []usecasex.Cursor{"b", "c"}, con.Cursors)

	p = usecasex.CursorPagination{
		Last: lo.ToPtr(int64(3)),
	}

	con = &consumer{}
	_, err = c.Paginate(ctx, bson.M{}, sortOpt, p.Wrap(), con)
	assert.NoError(t, err)
	assert.Equal(t, []usecasex.Cursor{"c", "d", "e"}, con.Cursors)
}

func TestClientCollection_DetailedPagination(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	seeds := []struct {
		id        string
		updatedAt int64
	}{
		{"a", 1000},
		{"b", 2000},
		{"c", 3000},
		{"d", 4000},
		{"e", 5000},
	}

	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s struct {
		id        string
		updatedAt int64
	}, i int) any {
		return bson.M{"id": s.id, "updatedAt": s.updatedAt}
	}))

	testCases := []struct {
		name       string
		sort       *usecasex.Sort
		pagination *usecasex.CursorPagination
		expected   []string
	}{
		{
			name:       "First 2, Ascending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: false},
			pagination: &usecasex.CursorPagination{First: lo.ToPtr(int64(2))},
			expected:   []string{"a", "b"},
		},
		{
			name:       "First 2, Descending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: true},
			pagination: &usecasex.CursorPagination{First: lo.ToPtr(int64(2))},
			expected:   []string{"e", "d"},
		},
		{
			name:       "Last 2, Ascending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: false},
			pagination: &usecasex.CursorPagination{Last: lo.ToPtr(int64(2))},
			expected:   []string{"d", "e"},
		},
		{
			name:       "Last 2, Descending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: true},
			pagination: &usecasex.CursorPagination{Last: lo.ToPtr(int64(2))},
			expected:   []string{"b", "a"},
		},
		{
			name:       "First 2 After 'b', Ascending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: false},
			pagination: &usecasex.CursorPagination{First: lo.ToPtr(int64(2)), After: usecasex.Cursor("b").Ref()},
			expected:   []string{"c", "d"},
		},
		{
			name:       "First 2 After 'd', Descending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: true},
			pagination: &usecasex.CursorPagination{First: lo.ToPtr(int64(2)), After: usecasex.Cursor("d").Ref()},
			expected:   []string{"c", "b"},
		},
		{
			name:       "Last 2 Before 'd', Ascending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: false},
			pagination: &usecasex.CursorPagination{Last: lo.ToPtr(int64(2)), Before: usecasex.Cursor("d").Ref()},
			expected:   []string{"b", "c"},
		},
		{
			name:       "Last 2 Before 'b', Descending",
			sort:       &usecasex.Sort{Key: "updatedAt", Reverted: true},
			pagination: &usecasex.CursorPagination{Last: lo.ToPtr(int64(2)), Before: usecasex.Cursor("b").Ref()},
			expected:   []string{"d", "c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			con := &consumer{}
			_, err := c.Paginate(ctx, bson.M{}, tc.sort, tc.pagination.Wrap(), con)
			assert.NoError(t, err)

			gotIDs := lo.Map(con.Cursors, func(c usecasex.Cursor, _ int) string {
				return string(c)
			})
			assert.Equal(t, tc.expected, gotIDs)
		})
	}
}

type consumer struct {
	Cursors []usecasex.Cursor
}

func (c *consumer) Consume(b bson.Raw) error {
	c.Cursors = append(c.Cursors, lo.FromPtr(lo.Must(getCursor(b))))
	return nil
}

func TestPaginate_SortLogic(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	seeds := []struct {
		id        string
		updatedAt int64
	}{
		{"a", 1000},
		{"b", 2000},
		{"c", 3000},
	}

	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s struct {
		id        string
		updatedAt int64
	}, i int) any {
		return bson.M{"id": s.id, "updatedAt": s.updatedAt}
	}))

	cases := []struct {
		name          string
		sortKey       string
		sortOrder     int
		expectedOrder []usecasex.Cursor
	}{
		{
			name:          "Sort by id ascending",
			sortKey:       "id",
			sortOrder:     1,
			expectedOrder: []usecasex.Cursor{"a", "b", "c"},
		},
		{
			name:          "Sort by id descending",
			sortKey:       "id",
			sortOrder:     -1,
			expectedOrder: []usecasex.Cursor{"c", "b", "a"},
		},
		{
			name:          "Sort by updatedAt ascending",
			sortKey:       "updatedAt",
			sortOrder:     1,
			expectedOrder: []usecasex.Cursor{"a", "b", "c"},
		},
		{
			name:          "Sort by updatedAt descending",
			sortKey:       "updatedAt",
			sortOrder:     -1,
			expectedOrder: []usecasex.Cursor{"c", "b", "a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sortOpt := &usecasex.Sort{
				Key:      tc.sortKey,
				Reverted: tc.sortOrder == -1,
			}
			p := usecasex.CursorPagination{
				First: lo.ToPtr(int64(len(seeds))),
			}

			con := &consumer{}
			_, err := c.Paginate(ctx, bson.M{}, sortOpt, p.Wrap(), con)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedOrder, con.Cursors)

		})
	}
}

func TestClientCollection_PaginateAggregation_Pipelines(t *testing.T) {
	ctx := context.Background()
	initDB := mongotest.Connect(t)
	c := NewCollection(initDB(t).Collection("test"))

	_, err := c.Client().InsertMany(ctx, []any{
		bson.M{"id": "a", "tags": bson.A{"x", "y"}},
		bson.M{"id": "b", "tags": bson.A{"x"}},
		bson.M{"id": "c", "tags": bson.A{"y"}},
		bson.M{"id": "d", "tags": bson.A{}},
	})
	assert.NoError(t, err)

	first2 := usecasex.CursorPagination{First: lo.ToPtr(int64(2))}.Wrap()
	after := usecasex.CursorPagination{First: lo.ToPtr(int64(2)), After: usecasex.Cursor("a").Ref()}.Wrap()
	last1 := usecasex.CursorPagination{Last: lo.ToPtr(int64(1))}.Wrap()

	tests := []struct {
		name      string
		pipeline  []any
		p         *usecasex.Pagination
		wantCount int64
		want      []usecasex.Cursor
		check     func(t *testing.T, raws []bson.Raw)
	}{
		{
			// reload path: computed fields survive the reload
			name: "set then match",
			pipeline: []any{
				bson.M{"$set": bson.M{"n": bson.M{"$size": "$tags"}}},
				bson.M{"$match": bson.M{"n": bson.M{"$gt": 0}}},
			},
			p:         first2,
			wantCount: 3,
			want:      []usecasex.Cursor{"a", "b"},
			check: func(t *testing.T, raws []bson.Raw) {
				for _, r := range raws {
					_, err := r.LookupErr("n")
					assert.NoError(t, err)
				}
			},
		},
		{
			name: "set then match, after cursor",
			pipeline: []any{
				bson.M{"$set": bson.M{"n": bson.M{"$size": "$tags"}}},
				bson.M{"$match": bson.M{"n": bson.M{"$gt": 0}}},
			},
			p:         after,
			wantCount: 3,
			want:      []usecasex.Cursor{"b", "c"},
		},
		{
			name: "set then match, last",
			pipeline: []any{
				bson.M{"$set": bson.M{"n": bson.M{"$size": "$tags"}}},
				bson.M{"$match": bson.M{"n": bson.M{"$gt": 0}}},
			},
			p:         last1,
			wantCount: 3,
			want:      []usecasex.Cursor{"c"},
		},
		{
			// rows share an _id: falls back to the direct query
			name:      "unwind",
			pipeline:  []any{bson.M{"$unwind": "$tags"}},
			p:         first2,
			wantCount: 4,
			want:      []usecasex.Cursor{"a", "a"},
			check: func(t *testing.T, raws []bson.Raw) {
				tags := lo.Map(raws, func(r bson.Raw, _ int) string { return r.Lookup("tags").StringValue() })
				assert.ElementsMatch(t, []string{"x", "y"}, tags)
			},
		},
		{
			// rows have no _id: falls back to the direct query
			name:      "project without _id",
			pipeline:  []any{bson.M{"$project": bson.M{"_id": 0, "id": 1}}},
			p:         first2,
			wantCount: 4,
			want:      []usecasex.Cursor{"a", "b"},
		},
		{
			// the caller's own $limit must apply to the whole input
			name:      "limit",
			pipeline:  []any{bson.M{"$sort": bson.M{"id": -1}}, bson.M{"$limit": 2}},
			p:         first2,
			wantCount: 2,
			want:      []usecasex.Cursor{"c", "d"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			con := &rawConsumer{}
			got, err := c.PaginateAggregation(ctx, tc.pipeline, nil, tc.p, con)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCount, got.TotalCount)
			assert.Equal(t, tc.want, lo.Map(con.raws, func(r bson.Raw, _ int) usecasex.Cursor {
				return lo.FromPtr(lo.Must(getCursor(r)))
			}))
			if tc.check != nil {
				tc.check(t, con.raws)
			}
		})
	}
}

type rawConsumer struct {
	raws []bson.Raw
}

func (c *rawConsumer) Consume(b bson.Raw) error {
	c.raws = append(c.raws, b)
	return nil
}

func TestReloadablePipeline(t *testing.T) {
	assert.False(t, reloadablePipeline(nil))
	assert.False(t, reloadablePipeline([]any{bson.M{"$match": bson.M{}}}))
	assert.True(t, reloadablePipeline([]any{bson.M{"$match": bson.M{}}, bson.M{"$lookup": bson.M{}}}))
	assert.True(t, reloadablePipeline([]any{bson.D{{Key: "$set", Value: bson.M{}}}}))
	assert.False(t, reloadablePipeline([]any{bson.M{"$set": bson.M{}}, bson.M{"$sort": bson.M{}}}))
	assert.False(t, reloadablePipeline([]any{bson.M{"$set": bson.M{}}, bson.M{"$group": bson.M{}}}))
	assert.False(t, reloadablePipeline([]any{"not a stage"}))
}
