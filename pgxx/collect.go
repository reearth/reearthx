package pgxx

// OrderByIDs reorders items to match the order of ids, matching each item by
// key(item). Ids with no matching item are omitted (not padded with a zero
// value), so the result contains only present items and never a nil element.
// This is the FindByIDs contract repositories share, and it matches Mongo, whose
// callers iterate the result without nil checks: fetch rows with
// `WHERE id = ANY($1)` (arbitrary order), restore the requested order, and drop
// ids that returned no row.
//
// Filter items (e.g. by workspace) before calling this; unmatched ids are dropped.
func OrderByIDs[K comparable, T any](ids []K, items []T, key func(T) K) []T {
	byID := make(map[K]T, len(items))
	for _, it := range items {
		byID[key(it)] = it
	}
	out := make([]T, 0, len(items))
	for _, id := range ids {
		if it, ok := byID[id]; ok {
			out = append(out, it)
		}
	}
	return out
}
