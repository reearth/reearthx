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
	c := NewClientCollection(initDB(t).Collection("test"))

	// seeds
	seeds := []string{"a", "b", "c"}
	_, _ = c.Client().InsertMany(ctx, lo.Map(seeds, func(s string, i int) any {
		return bson.M{"id": s, "i": len(seeds) - i}
	}))

	// nil
	got, goterr := c.Paginate(ctx, nil, nil, nil, nil)
	assert.Nil(t, got)
	assert.NoError(t, goterr)

	// cursor: first
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

	// cursor: first, after
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

	// cursor: last
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

	// cursor: last, before
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

	// cursor: offset
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

	// cursor: offset, sort
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

type consumer struct {
	Cursors []usecasex.Cursor
}

func (c *consumer) Consume(b bson.Raw) error {
	c.Cursors = append(c.Cursors, lo.FromPtr(lo.Must(getCursor(b))))
	return nil
}
