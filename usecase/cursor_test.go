package usecase

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestCursorFromRef(t *testing.T) {
	assert.Equal(t, lo.ToPtr(Cursor("a")), CursorFromRef(lo.ToPtr("a")))
	assert.Nil(t, CursorFromRef(nil))
}

func TestCursor_Ref(t *testing.T) {
	assert.Equal(t, lo.ToPtr(Cursor("a")), Cursor("a").Ref())
}

func TestCursor_CopyRef(t *testing.T) {
	c := lo.ToPtr(Cursor("a"))
	got := c.CopyRef()
	assert.Equal(t, c, got)
	assert.NotSame(t, c, got)
	assert.Nil(t, (*Cursor)(nil).CopyRef())
}

func TestCursor_StringRef(t *testing.T) {
	c := lo.ToPtr(Cursor("a"))
	assert.Equal(t, lo.ToPtr("a"), c.StringRef())
	assert.Nil(t, (*Cursor)(nil).StringRef())
}
