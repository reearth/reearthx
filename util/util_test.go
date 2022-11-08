package util

import (
	"errors"
	"net/url"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestToPtrIfNotEmpty(t *testing.T) {
	assert.Nil(t, ToPtrIfNotEmpty(""))
	assert.NotEqual(t, lo.ToPtr(""), ToPtrIfNotEmpty(""))
	assert.Equal(t, lo.ToPtr("a"), ToPtrIfNotEmpty("a"))
}

func TestOrError(t *testing.T) {
	a := &struct{}{}
	err := errors.New("err")
	got, goterr := OrError(a, err)
	assert.Same(t, a, got)
	assert.Same(t, err, goterr)
	got, goterr = OrError[*struct{}](nil, err)
	assert.Nil(t, got)
	assert.Same(t, err, goterr)
}

func TestMapError(t *testing.T) {
	f := func() (string, error) { return "a", errors.New("a") }
	got, goterr := MapError(f())(func(e error) error { return errors.New("b" + e.Error()) })
	assert.Equal(t, "a", got)
	assert.Equal(t, errors.New("ba"), goterr)
}

func TestCopyRef(t *testing.T) {
	target := lo.ToPtr(1)
	got := CloneRef(target)
	assert.Equal(t, target, got)
	assert.NotSame(t, target, got)
	assert.Nil(t, CloneRef[int](nil))
}

func TestCopyURL(t *testing.T) {
	u, _ := url.Parse("http://localhost")
	u2, _ := url.Parse("http://aaa:bbb@localhost")

	tests := []struct {
		name string
		args *url.URL
	}{
		{
			name: "normal",
			args: u,
		},
		{
			name: "userinfo",
			args: u2,
		},
		{
			name: "nil",
			args: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CopyURL(tt.args)
			assert.Equal(t, tt.args, got)
			if got != nil {
				assert.NotSame(t, tt.args, got)
				if got.User != nil {
					assert.NotSame(t, tt.args.User, got.User)
				}
			}
		})
	}
}

func TestDR(t *testing.T) {
	f := func() (string, error) {
		return "a", nil
	}
	assert.Equal(t, "a", DR(f()))
}

func TestTry(t *testing.T) {
	err := errors.New("try")
	assert.Same(t, err, Try(func() error { return err }, func() error { panic("should not called") }))
	assert.NoError(t, Try(func() error { return nil }, func() error { return nil }))
}

func TestSortedEntries(t *testing.T) {
	for i := 0; i < 10; i++ {
		assert.Equal(t, []lo.Entry[string, string]{
			{Key: "a", Value: "1"},
			{Key: "b", Value: "2"},
		}, SortedEntries(map[string]string{
			"b": "2",
			"a": "1",
		}))
	}
}
