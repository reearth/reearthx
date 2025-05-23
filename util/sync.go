package util

import (
	"slices"
	"sync"
)

type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}

func SyncMapFrom[K comparable, V any](entries map[K]V) *SyncMap[K, V] {
	m := &SyncMap[K, V]{}
	m.StoreAll(entries)
	return m
}

func (m *SyncMap[K, V]) Load(key K) (vv V, _ bool) {
	if m == nil {
		return
	}

	v, ok := m.m.Load(key)
	if ok {
		vv = v.(V)
	}
	return vv, ok
}

func (m *SyncMap[K, V]) LoadAll(keys ...K) (r []V) {
	if m == nil {
		return
	}

	m.Range(func(k K, v V) bool {
		if slices.Contains(keys, k) {
			r = append(r, v)
		}
		return true
	})
	return r
}

func (m *SyncMap[K, V]) LoadOr(key K, o V) (res V) {
	if m == nil {
		return
	}

	v, ok := m.m.Load(key)
	if ok {
		return v.(V)
	}
	return o
}

func (m *SyncMap[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

func (m *SyncMap[K, V]) StoreAll(entries map[K]V) {
	for k, v := range entries {
		m.m.Store(k, v)
	}
}

func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (vv V, _ bool) {
	v, ok := m.m.LoadOrStore(key, value)
	if ok {
		vv = v.(V)
	}
	return vv, ok
}

func (m *SyncMap[K, V]) LoadAndDelete(key K) (vv V, ok bool) {
	v, ok := m.m.LoadAndDelete(key)
	if ok {
		vv = v.(V)
	}
	return vv, ok
}

func (m *SyncMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *SyncMap[K, V]) DeleteAll(key ...K) {
	for _, k := range key {
		m.Delete(k)
	}
}

// Range calls f sequentially for each key and value present in the map. If f returns false, range stops the iteration.
func (m *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (m *SyncMap[K, V]) Unsync() map[K]V {
	res := make(map[K]V, m.Len())
	m.Range(func(k K, v V) bool {
		res[k] = v
		return true
	})
	return res
}

func (m *SyncMap[K, V]) Find(f func(key K, value V) bool) (v V) {
	m.Range(func(key K, value V) bool {
		if f(key, value) {
			v = value
			return false
		}
		return true
	})
	return
}

func (m *SyncMap[K, V]) FindAll(f func(key K, value V) bool) (v []V) {
	m.Range(func(key K, value V) bool {
		if f(key, value) {
			v = append(v, value)
		}
		return true
	})
	return
}

func (m *SyncMap[K, V]) CountAll(f func(key K, value V) bool) (i int) {
	m.Range(func(key K, value V) bool {
		if f(key, value) {
			i++
		}
		return true
	})
	return
}

func (m *SyncMap[K, V]) Clone() *SyncMap[K, V] {
	if m == nil {
		return nil
	}
	n := &SyncMap[K, V]{}
	m.Range(func(key K, value V) bool {
		n.Store(key, value)
		return true
	})
	return n
}

func (m *SyncMap[K, V]) Map(f func(K, V) V) *SyncMap[K, V] {
	n := &SyncMap[K, V]{}
	m.Range(func(key K, value V) bool {
		n.Store(key, f(key, value))
		return true
	})
	return n
}

func (m *SyncMap[K, V]) Merge(n *SyncMap[K, V]) {
	n.Range(func(key K, value V) bool {
		m.Store(key, value)
		return true
	})
}

func (m *SyncMap[K, V]) Keys() (l []K) {
	m.Range(func(key K, _ V) bool {
		l = append(l, key)
		return true
	})
	return l
}

func (m *SyncMap[K, V]) Values() (l []V) {
	m.Range(func(_ K, value V) bool {
		l = append(l, value)
		return true
	})
	return l
}

func (m *SyncMap[K, V]) Len() (i int) {
	m.m.Range(func(_ any, _ any) bool {
		i++
		return true
	})
	return
}

type LockMap[T comparable] struct {
	m SyncMap[T, *sync.Mutex]
}

func (m *LockMap[T]) Lock(k T) func() {
	nl := &sync.Mutex{}
	l, ok := m.m.LoadOrStore(k, nl)
	if ok {
		l.Lock()
	} else {
		nl.Lock()
	}
	return func() {
		m.Unlock(k)
	}
}

func (m *LockMap[T]) Unlock(k T) {
	if l, ok := m.m.LoadAndDelete(k); ok {
		l.Unlock()
	}
}
