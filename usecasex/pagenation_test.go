package usecasex

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestPagination_Wrap(t *testing.T) {
	cp := CursorPagination{
		Before: lo.ToPtr(Cursor("a")),
		After:  lo.ToPtr(Cursor("b")),
		First:  lo.ToPtr(int64(100)),
		Last:   lo.ToPtr(int64(10)),
	}
	op := OffsetPagination{Offset: 100, Limit: 10}
	assert.Equal(t, &Pagination{Cursor: &cp}, cp.Wrap())
	assert.Equal(t, &Pagination{Offset: &op}, op.Wrap())
}

func TestPagination_Clone(t *testing.T) {
	target := &Pagination{
		Cursor: &CursorPagination{
			Before: lo.ToPtr(Cursor("a")),
			After:  lo.ToPtr(Cursor("b")),
			First:  lo.ToPtr(int64(100)),
			Last:   lo.ToPtr(int64(10)),
		},
		Offset: &OffsetPagination{Offset: 100, Limit: 10},
	}
	got := target.Clone()

	assert.Equal(t, got, target)
	assert.NotSame(t, got, target)
	assert.Nil(t, (*Pagination)(nil).Clone())
}
