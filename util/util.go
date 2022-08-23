package util

import (
	"net/url"

	"github.com/samber/lo"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// IsZero
//
// Deprecated: use lo.IsEmpty instead.
func IsZero[T comparable](v T) bool {
	return lo.IsEmpty(v)
}

// IsNotZero
//
// Deprecated: use lo.IsNotEmpty instead.
func IsNotZero[T comparable](v T) bool {
	return lo.IsNotEmpty(v)
}

// Deref
//
// Deprecated: use lo.FromPtr instead.
func Deref[T any](r *T) T {
	return lo.FromPtr(r)
}

// DerefOr
//
// Deprecated: use lo.FromPtrOrust instead.
func DerefOr[T any](ref *T, def T) T {
	return lo.FromPtrOr(ref, def)
}

// Unwrap
//
// Deprecated: use lo.Must instead.
func Unwrap[T any](t T, err error) T {
	return lo.Must(t, err)
}

func ToPtrIfNotEmpty[T comparable](t T) *T {
	if lo.IsEmpty(t) {
		return nil
	}
	return &t
}

func OrError[T comparable](t T, err error) (r T, _ error) {
	if lo.IsEmpty(t) && err != nil {
		return r, err
	}
	return t, nil
}

func CloneRef[T any](r *T) *T {
	if r == nil {
		return nil
	}
	r2 := *r
	return &r2
}

func CopyURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	v := CloneRef(u)
	v.User = CloneRef(u.User)
	return v
}

// DR discards right
func DR[A, B any](a A, _ B) A {
	return a
}

func Try(tries ...func() error) error {
	for _, f := range tries {
		if f == nil {
			continue
		}
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func SortedEntries[K constraints.Ordered, V any](m map[K]V) []lo.Entry[K, V] {
	entries := lo.Entries(m)
	slices.SortStableFunc(entries, func(a, b lo.Entry[K, V]) bool {
		return a.Key < b.Key
	})
	return entries
}
