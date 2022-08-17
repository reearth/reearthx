package usecasex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPageInfo(t *testing.T) {
	assert.Equal(t, &PageInfo{
		TotalCount:      100,
		StartCursor:     Cursor("a").Ref(),
		EndCursor:       Cursor("b").Ref(),
		HasNextPage:     true,
		HasPreviousPage: true,
	}, NewPageInfo(100, Cursor("a").Ref(), Cursor("b").Ref(), true, true))
}

func TestEmptyPageInfo(t *testing.T) {
	assert.Equal(t, &PageInfo{}, EmptyPageInfo())
}

func TestPageInfo_OrEmpty(t *testing.T) {
	p := &PageInfo{
		TotalCount:      100,
		StartCursor:     Cursor("a").Ref(),
		EndCursor:       Cursor("b").Ref(),
		HasNextPage:     true,
		HasPreviousPage: true,
	}
	assert.Same(t, p, p.OrEmpty())
	assert.Equal(t, &PageInfo{}, (*PageInfo)(nil).OrEmpty())
}

func TestPageInfo_Clone(t *testing.T) {
	p := &PageInfo{
		TotalCount:      100,
		StartCursor:     Cursor("a").Ref(),
		EndCursor:       Cursor("b").Ref(),
		HasNextPage:     true,
		HasPreviousPage: true,
	}
	got := p.Clone()
	assert.Equal(t, p, got)
	assert.NotSame(t, p, got)
	assert.Nil(t, (*PageInfo)(nil).Clone())
}
