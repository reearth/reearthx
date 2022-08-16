package util

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

type Entry[K constraints.Ordered, V any] struct {
	Key   K
	Value V
}

func StableEntries[K constraints.Ordered, V any](m map[K]V) []Entry[K, V] {
	entries := Entries(m)
	slices.SortFunc(entries, func(a, b Entry[K, V]) bool {
		return a.Key < b.Key
	})
	return entries
}

func Entries[K constraints.Ordered, V any](m map[K]V) []Entry[K, V] {
	entries := make([]Entry[K, V], 0, len(m))
	for k, v := range m {
		entries = append(entries, Entry[K, V]{
			Key:   k,
			Value: v,
		})
	}
	return entries
}

func FromEntries[K constraints.Ordered, V any](entries []Entry[K, V]) map[K]V {
	m := make(map[K]V, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m
}
