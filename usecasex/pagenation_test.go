package usecasex

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestNewPagination(t *testing.T) {
	b := lo.ToPtr(Cursor(""))
	a := lo.ToPtr(Cursor(""))
	f := lo.ToPtr(0)
	l := lo.ToPtr(0)
	assert.Equal(t, &Pagination{
		Before: b,
		After:  a,
		First:  f,
		Last:   l,
	}, NewPagination(f, l, b, a))
}
