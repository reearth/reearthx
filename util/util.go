package util

import (
	"net/url"
	"slices"

	"github.com/samber/lo"
	"golang.org/x/exp/constraints"
)

func ToPtrIfNotEmpty[T comparable](t T) *T {
	if lo.IsEmpty(t) {
		return nil
	}
	return &t
}

func OrError[T comparable](t T, err error) (r T, _ error) {
	return MapError(t, err)(nil)
}

func MapError[T comparable](t T, err error) func(m func(error) error) (T, error) {
	return func(m func(error) error) (T, error) {
		if err != nil {
			if m == nil {
				return t, err
			}
			return t, m(err)
		}
		return t, nil
	}
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
	slices.SortStableFunc(entries, func(a, b lo.Entry[K, V]) int {
		switch {
		case a.Key < b.Key:
			return -1
		case a.Key > b.Key:
			return 1
		default:
			return 0
		}
	})
	return entries
}
