package util

import (
	"errors"
	"net/url"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	err := errors.New("ERR")
	Must(nil)
	assert.PanicsWithValue(t, err, func() {
		Must(err)
	})
}

func TestIsZero(t *testing.T) {
	assert.True(t, IsZero(0))
	assert.False(t, IsZero(-1))
	assert.True(t, IsZero(struct {
		A int
		B string
	}{}))
	assert.False(t, IsZero(struct {
		A int
		B string
	}{A: 1}))
	assert.True(t, IsZero((*(struct{}))(nil)))
	assert.False(t, IsZero((*(struct{}))(&struct{}{})))
}

func TestIsNotZero(t *testing.T) {
	assert.False(t, IsNotZero(0))
	assert.True(t, IsNotZero(-1))
	assert.False(t, IsNotZero(struct {
		A int
		B string
	}{}))
	assert.True(t, IsNotZero(struct {
		A int
		B string
	}{A: 1}))
	assert.False(t, IsNotZero((*(struct{}))(nil)))
	assert.True(t, IsNotZero((*(struct{}))(&struct{}{})))
}

func TestDeref(t *testing.T) {
	assert.Equal(t, struct{ A int }{}, Deref((*(struct{ A int }))(nil)))
	assert.Equal(t, struct{ A int }{A: 1}, Deref((*(struct{ A int }))(&struct{ A int }{A: 1})))
}

func TestDerefOr(t *testing.T) {
	assert.Equal(t, "b", DerefOr((*string)(nil), "b"))
	assert.Equal(t, "a", DerefOr(lo.ToPtr("a"), ""))
}

func TestUnwrap(t *testing.T) {
	err := errors.New("hoge")
	res := lo.ToPtr(1)
	assert.PanicsWithValue(t, err, func() { _ = Unwrap(1, err) })
	assert.Same(t, res, Unwrap(res, nil))
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
