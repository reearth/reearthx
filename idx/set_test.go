package idx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_NewSet(t *testing.T) {
	a := New[T]()
	assert.Equal(t, &Set[T]{
		l: nil,
		m: map[string]ID[T]{},
	}, NewSet[T]())
	assert.Equal(t, &Set[T]{
		l: List[T]{a},
		m: map[string]ID[T]{
			a.String(): a,
		},
	}, NewSet(a))
}

func TestSet_Has(t *testing.T) {
	a := New[T]()
	b := New[T]()
	assert.False(t, (*Set[T])(nil).Has(a, b))
	assert.True(t, NewSet(a).Has(a))
	assert.False(t, NewSet(a).Has(b))
}

func TestSet_List(t *testing.T) {
	a := New[T]()
	b := New[T]()
	assert.Nil(t, (*Set[T])(nil).List())
	assert.Empty(t, NewSet[T]().List())
	assert.Equal(t, List[T]{a, b}, NewSet(a, b).List())
}

func TestSet_Clone(t *testing.T) {
	a := New[T]()
	b := New[T]()
	s := NewSet(a, b)
	assert.Nil(t, (*Set[T])(nil).Clone())
	assert.Equal(t, &Set[T]{m: map[string]ID[T]{}}, NewSet[T]().Clone())
	assert.Equal(t, s, s.Clone())
	assert.NotSame(t, s, s.Clone())
}

func TestSet_Add(t *testing.T) {
	a := New[T]()
	b := New[T]()
	s := NewSet(a)
	(*Set[T])(nil).Add(a, b)
	s.Add(a, b)
	expected := NewSet(a, b)
	assert.Equal(t, expected.List(), s.List())
	assert.Equal(t, expected, s)
}

func TestSet_Merge(t *testing.T) {
	a := New[T]()
	b := New[T]()
	s := NewSet(a)
	u := NewSet(a, b)
	(*Set[T])(nil).Merge(u)
	s.Merge(u)
	expected := NewSet(a, b)
	assert.Equal(t, expected.List(), s.List())
	assert.Equal(t, expected, s)
}

func TestSet_Concat(t *testing.T) {
	a := New[T]()
	b := New[T]()
	s := NewSet(a)
	u := NewSet(a, b)
	assert.Nil(t, (*Set[T])(nil).Concat(u))
	expected := NewSet(a, b)
	result := s.Concat(u)
	assert.Equal(t, expected.List(), result.List())
	assert.Equal(t, expected, result)
	assert.Equal(t, NewSet(a), s)
}

func TestSet_Delete(t *testing.T) {
	a := New[T]()
	b := New[T]()
	c := New[T]()
	s := NewSet(a, b, c)
	(*Set[T])(nil).Delete(a, b)
	s.Delete(a, b)
	expected := NewSet(c)
	assert.Equal(t, expected.List(), s.List())
	assert.Equal(t, expected, s)
}
