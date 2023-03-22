package util

import "github.com/samber/lo"

type DiffResult[T any] struct {
	Added      []T
	Updated    []DiffUpdate[T]
	NotUpdated []DiffUpdate[T]
	Deleted    []T
}

func (d DiffResult[T]) UpdatedPrev() []T {
	return lo.Map(d.Updated, func(u DiffUpdate[T], _ int) T { return u.Prev })
}

func (d DiffResult[T]) UpdatedNext() []T {
	return lo.Map(d.Updated, func(u DiffUpdate[T], _ int) T { return u.Next })
}

func (d DiffResult[T]) NotUpdatedPrev() []T {
	return lo.Map(d.NotUpdated, func(u DiffUpdate[T], _ int) T { return u.Prev })
}

func (d DiffResult[T]) NotUpdatedNext() []T {
	return lo.Map(d.NotUpdated, func(u DiffUpdate[T], _ int) T { return u.Next })
}

type DiffUpdate[T any] struct {
	Prev T
	Next T
}

func Diff[T any](prev, next []T, keyEqual, isUpdated func(T, T) bool) (r DiffResult[T]) {
	for _, pe := range prev {
		ne, ok := lo.Find(next, func(e T) bool {
			return keyEqual(pe, e)
		})
		if !ok {
			r.Deleted = append(r.Deleted, pe)
		} else if isUpdated(pe, ne) {
			r.Updated = append(r.Updated, DiffUpdate[T]{
				Prev: pe,
				Next: ne,
			})
		} else {
			r.NotUpdated = append(r.NotUpdated, DiffUpdate[T]{
				Prev: pe,
				Next: ne,
			})
		}
	}

	r.Added = lo.Filter(next, func(ne T, _ int) bool {
		return !lo.SomeBy(prev, func(pe T) bool {
			return keyEqual(pe, ne)
		})
	})
	if len(r.Added) == 0 {
		r.Added = nil
	}

	return
}
