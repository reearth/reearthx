package usecasex

import "github.com/reearth/reearthx/util"

// CursorPagination is a struct for Relay-Style Cursor Pagination
// ref: https://www.apollographql.com/docs/react/features/pagination/#relay-style-cursor-pagination
type CursorPagination struct {
	Before *Cursor `json:"before"`
	After  *Cursor `json:"after"`
	First  *int64  `json:"first"`
	Last   *int64  `json:"last"`
}

func (p *CursorPagination) Clone() *CursorPagination {
	if p == nil {
		return nil
	}

	return &CursorPagination{
		Before: util.CloneRef(p.Before),
		After:  util.CloneRef(p.After),
		First:  util.CloneRef(p.First),
		Last:   util.CloneRef(p.Last),
	}
}

func (p CursorPagination) Wrap() *Pagination {
	return &Pagination{
		Cursor: &p,
	}
}

type OffsetPagination struct {
	Offset int64 `json:"offset"`
	Limit  int64 `json:"limit"`
}

func (p OffsetPagination) Wrap() *Pagination {
	return &Pagination{
		Offset: &p,
	}
}

type Pagination struct {
	Cursor *CursorPagination
	Offset *OffsetPagination
}

func (p *Pagination) Clone() *Pagination {
	if p == nil {
		return nil
	}

	return &Pagination{
		Cursor: p.Cursor.Clone(),
		Offset: util.CloneRef(p.Offset),
	}
}
