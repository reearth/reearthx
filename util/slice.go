package util

import "github.com/samber/lo"

type Element[T any] struct {
	Index   int
	Element T
}

// Enumerate returns a new slice with each element and its index.
func Enumerate[T any](collection []T) []Element[T] {
	if collection == nil {
		return nil
	}

	return lo.Map(collection, func(e T, i int) Element[T] {
		return Element[T]{
			Index:   i,
			Element: e,
		}
	})
}

// Map is similar to lo.Map, but accepts an iteratee without the index argument.
func Map[T any, V any](collection []T, iteratee func(v T) V) []V {
	if collection == nil {
		return nil
	}

	return lo.Map(collection, func(v T, _ int) V {
		return iteratee(v)
	})
}

// TryMap is similar to Map, but when an error occurs in the iteratee, it terminates the iteration and returns an error.
func TryMap[T any, V any](collection []T, iteratee func(v T) (V, error)) ([]V, error) {
	if collection == nil {
		return nil, nil
	}

	m := make([]V, 0, len(collection))
	for _, e := range collection {
		j, err := iteratee(e)
		if err != nil {
			return nil, err
		}
		m = append(m, j)
	}
	return m, nil
}

// TryFilterMap is similar to TryMap, but if the iteratee returns nil, that element will be omitted from the new slice.
func TryFilterMap[T any, V any](collection []T, iteratee func(v T) (V, bool, error)) ([]V, error) {
	if collection == nil {
		return nil, nil
	}

	m := make([]V, 0, len(collection))
	for _, e := range collection {
		j, ok, err := iteratee(e)
		if err != nil {
			return nil, err
		}
		if ok {
			m = append(m, j)
		}
	}
	return m, nil
}

// Filter is similar to lo.Filter, but accepts an iteratee without the index argument.
func Filter[T any](collection []T, iteratee func(v T) bool) []T {
	if collection == nil {
		return nil
	}

	return lo.Filter(collection, func(v T, _ int) bool {
		return iteratee(v)
	})
}

// FilterMap is similar to Map, but if the iteratee returns nil, that element will be omitted from the new slice.
func FilterMap[T any, V any](collection []T, iteratee func(v T) *V) []V {
	if collection == nil {
		return nil
	}

	m := make([]V, 0, len(collection))
	for _, e := range collection {
		if j := iteratee(e); j != nil {
			m = append(m, *j)
		}
	}
	return m
}

// FilterMapR is similar to FilterMap, but if the return value of the iteratee is not nil,
// it is not dereferenced and is used as the value of the new element.
func FilterMapR[T any, V any](collection []T, iteratee func(v T) *V) []*V {
	if collection == nil {
		return nil
	}

	m := make([]*V, 0, len(collection))
	for _, e := range collection {
		if j := iteratee(e); j != nil {
			m = append(m, j)
		}
	}
	return m
}

// DerefSlice drops nil elements in the slice and return a new slice with dereferenced elements.
func DerefSlice[T any](collection []*T) []T {
	return FilterMap(collection, func(e *T) *T {
		return e
	})
}
