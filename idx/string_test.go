package idx

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestStringIDFromRef(t *testing.T) {
	assert.Equal(t, lo.ToPtr(StringID[T]("aa")), StringIDFromRef[T](lo.ToPtr("aa")))
	assert.Nil(t, StringIDFromRef[T](nil))
}

func TestStringID_Ref(t *testing.T) {
	id := StringID[T]("a")
	assert.Equal(t, &id, id.Ref())

	empty := StringID[T]("")
	assert.Nil(t, empty.Ref())
}

func TestStringID_CloneRef(t *testing.T) {
	s := lo.ToPtr(StringID[T]("a"))
	res := s.CloneRef()
	assert.Equal(t, s, res)
	assert.NotSame(t, s, res)
	assert.Nil(t, (*StringID[T])(nil).CloneRef())
}

func TestStringID_String(t *testing.T) {
	id := StringID[T]("a")
	assert.Equal(t, "a", (&id).String())

	var empty *StringID[T]
	assert.Equal(t, "", empty.String())
}

func TestStringID_StringRef(t *testing.T) {
	id := StringID[T]("a")
	assert.Equal(t, lo.ToPtr("a"), (&id).StringRef())
	assert.Nil(t, (*StringID[T])(nil).StringRef())
}
