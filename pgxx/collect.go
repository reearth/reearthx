package pgxx

// OrderByIDs reorders items to match the order of ids, matching each item by
// key(item). An id with no matching item gets the zero value of T (nil for
// pointer types). This is the FindByIDs contract repositories share: fetch rows
// with `WHERE id = ANY($1)` (arbitrary order), then restore the requested order
// with nil placeholders for missing ids.
//
// Filter items (e.g. by workspace) before calling this; unmatched ids become nil.
func OrderByIDs[K comparable, T any](ids []K, items []T, key func(T) K) []T {
	byID := make(map[K]T, len(items))
	for _, it := range items {
		byID[key(it)] = it
	}
	out := make([]T, len(ids))
	for i, id := range ids {
		out[i] = byID[id] // zero value (nil) when absent
	}
	return out
}
