package util

import (
	"errors"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestEnumerate(t *testing.T) {
	assert.Nil(t, Enumerate[int](nil))
	assert.Equal(t, []Element[int]{
		{Index: 0, Element: 3},
		{Index: 1, Element: 2},
		{Index: 2, Element: 1},
	}, Enumerate([]int{3, 2, 1}))
}

func TestMap(t *testing.T) {
	assert.Nil(t, Map[int, bool](nil, nil))
	assert.Equal(t, []bool{true, false, true}, Map([]int{1, 0, 2}, func(i int) bool { return i != 0 }))
}

func TestTryMap(t *testing.T) {
	res, err := TryMap[int, bool](nil, nil)
	assert.Nil(t, res)
	assert.NoError(t, err)

	iteratee := func(i int) (bool, error) {
		if i == 0 {
			return false, errors.New("aaa")
		}
		return true, nil
	}
	res, err = TryMap([]int{1, 2, 3}, iteratee)
	assert.Equal(t, []bool{true, true, true}, res)
	assert.NoError(t, err)

	res, err = TryMap([]int{1, 0, 3}, iteratee)
	assert.Nil(t, res)
	assert.Equal(t, errors.New("aaa"), err)
}

func TestTryFilterMap(t *testing.T) {
	res, err := TryMap[int, bool](nil, nil)
	assert.Nil(t, res)
	assert.NoError(t, err)

	iteratee := func(i int) (bool, bool, error) {
		if i == 0 {
			return false, false, errors.New("aaa")
		}
		if i == 1 {
			return true, false, nil
		}
		return true, true, nil
	}
	res, err = TryFilterMap([]int{1, 2, 3}, iteratee)
	assert.Equal(t, []bool{true, true}, res)
	assert.NoError(t, err)

	res, err = TryFilterMap([]int{1, 0, 3}, iteratee)
	assert.Nil(t, res)
	assert.Equal(t, errors.New("aaa"), err)
}

func TestFilterMap(t *testing.T) {
	assert.Nil(t, FilterMap[int, bool](nil, nil))
	assert.Equal(t, []bool{true, false}, FilterMap([]int{1, 0, 2}, func(i int) *bool {
		if i == 0 {
			return nil
		}
		return lo.ToPtr(i == 1)
	}))
}

func TestFilterMapOk(t *testing.T) {
	assert.Nil(t, FilterMapOk[int, bool](nil, nil))
	assert.Equal(t, []bool{true, false}, FilterMapOk([]int{1, 0, 2}, func(i int) (bool, bool) {
		if i == 0 {
			return false, false
		}
		return i == 1, true
	}))
}

func TestFilterR(t *testing.T) {
	assert.Nil(t, FilterMapR[int, bool](nil, nil))
	assert.Equal(t, []*bool{lo.ToPtr(true), lo.ToPtr(false)}, FilterMapR([]int{1, 0, 2}, func(i int) *bool {
		if i == 0 {
			return nil
		}
		return lo.ToPtr(i == 1)
	}))
}

func TestAll(t *testing.T) {
	assert.True(t, All([]int{1, 2, 3}, func(i int) bool { return i < 4 }))
	assert.False(t, All([]int{1, 2, 3}, func(i int) bool { return i < 3 }))
}

func TestAny(t *testing.T) {
	assert.True(t, Any([]int{1, 2, 3}, func(i int) bool { return i == 1 }))
	assert.False(t, Any([]int{1, 2, 3}, func(i int) bool { return i == 4 }))
}

func TestFilter(t *testing.T) {
	assert.Nil(t, Filter[int](nil, nil))
	assert.Equal(t, []int{1, 2}, Filter([]int{1, 0, 2}, func(i int) bool {
		return i != 0
	}))
}

func TestDerefSlice(t *testing.T) {
	assert.Nil(t, DerefSlice[int](nil))
	assert.Equal(t, []int{1, 0, 2}, DerefSlice([]*int{lo.ToPtr(1), nil, lo.ToPtr(0), lo.ToPtr(2)}))
}

func TestSubset(t *testing.T) {
	tests := []struct {
		name          string
		collection    []string
		subCollection []string
		want          bool
	}{
		{
			name:          "",
			collection:    []string{},
			subCollection: []string{},
			want:          true,
		},
		{
			name:          "",
			collection:    []string{},
			subCollection: nil,
			want:          true,
		},
		{
			name:          "",
			collection:    []string{"v1", "v2", "v3"},
			subCollection: nil,
			want:          true,
		},
		{
			name:          "",
			collection:    []string{"v1", "v2", "v3"},
			subCollection: []string{},
			want:          true,
		},
		{
			name:          "",
			collection:    []string{"v1", "v2", "v3"},
			subCollection: []string{"v1"},
			want:          true,
		},
		{
			name:          "",
			collection:    []string{"v1", "v2", "v3"},
			subCollection: []string{"v1", "v2"},
			want:          true,
		},
		{
			name:          "",
			collection:    []string{"v1", "v2", "v3"},
			subCollection: []string{"v1", "v2", "v3"},
			want:          true,
		},
		{
			name:          "",
			collection:    []string{"v1", "v2", "v3"},
			subCollection: []string{"v4"},
			want:          false,
		},
		{
			name:          "",
			collection:    []string{"v1", "v2", "v3"},
			subCollection: []string{"v1", "v2", "v4"},
			want:          false,
		},
		{
			name:          "",
			collection:    nil,
			subCollection: []string{"v1"},
			want:          false,
		},
		{
			name:          "",
			collection:    []string{},
			subCollection: []string{"v1"},
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Subset(tt.collection, tt.subCollection))
		})
	}
}

func TestHasDuplicates(t *testing.T) {
	tests := []struct {
		name       string
		collection []string
		want       bool
	}{
		{
			name:       "has",
			collection: []string{"123", "321", "123"},
			want:       true,
		},
		{
			name:       "has not",
			collection: []string{"123", "321", "1232"},
			want:       false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, HasDuplicates(tt.collection))
		})
	}
}
