package mongo

import (
	"errors"

	"github.com/reearth/mongogit"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/samber/lo"
)

// The mongogit library owns its own index, pagination, and error types. These
// helpers bridge them to the mongox/usecasex/rerror types the rest of the asset
// mongo layer speaks.

func toMongoxIndexes(in []mongogit.Index) []mongox.Index {
	return lo.Map(in, func(i mongogit.Index, _ int) mongox.Index {
		return mongox.Index(i)
	})
}

func toGitSort(s *usecasex.Sort) *mongogit.Sort {
	if s == nil {
		return nil
	}
	return &mongogit.Sort{Key: s.Key, Reverted: s.Reverted}
}

func toGitPagination(p *usecasex.Pagination) *mongogit.Pagination {
	if p == nil {
		return nil
	}
	out := &mongogit.Pagination{}
	if p.Cursor != nil {
		out.Cursor = &mongogit.CursorPagination{
			Before: toGitCursor(p.Cursor.Before),
			After:  toGitCursor(p.Cursor.After),
			First:  p.Cursor.First,
			Last:   p.Cursor.Last,
		}
	}
	if p.Offset != nil {
		out.Offset = &mongogit.OffsetPagination{
			Offset: p.Offset.Offset,
			Limit:  p.Offset.Limit,
		}
	}
	return out
}

func toGitCursor(c *usecasex.Cursor) *mongogit.Cursor {
	if c == nil {
		return nil
	}
	g := mongogit.Cursor(*c)
	return &g
}

func fromGitPageInfo(pi *mongogit.PageInfo) *usecasex.PageInfo {
	if pi == nil {
		return nil
	}
	return usecasex.NewPageInfo(
		pi.TotalCount,
		fromGitCursor(pi.StartCursor),
		fromGitCursor(pi.EndCursor),
		pi.HasNextPage,
		pi.HasPreviousPage,
	)
}

func fromGitCursor(c *mongogit.Cursor) *usecasex.Cursor {
	if c == nil {
		return nil
	}
	u := usecasex.Cursor(*c)
	return &u
}

// mapErr translates the mongogit sentinel errors back to the rerror sentinels
// the asset usecases match on (e.g. errors.Is(err, rerror.ErrNotFound)).
func mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, mongogit.ErrNotFound):
		return rerror.ErrNotFound
	case errors.Is(err, mongogit.ErrInvalidParams):
		return rerror.ErrInvalidParams
	case errors.Is(err, mongogit.ErrAlreadyExists):
		return rerror.ErrAlreadyExists
	default:
		return err
	}
}
