package idx

import (
	"sort"

	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
)

type List[T Type] []ID[T]

type RefList[T Type] []*ID[T]

func ListFrom[T Type](ids []string) (List[T], error) {
	got, err := util.TryMap(ids, fromNID)
	if err != nil {
		return nil, err
	}
	return nidsTo[T](got), nil
}

func MustList[T Type](ids []string) List[T] {
	got, err := ListFrom[T](ids)
	if err != nil {
		lo.Must[any](nil, err)
	}
	return got
}

func (l List[T]) list() util.List[nid] {
	nids := newNIDs(l)
	values := make([]nid, len(nids))
	for i, n := range nids {
		values[i] = *n
	}
	return util.List[nid](values)
}

func (l List[T]) Has(ids ...ID[T]) bool {
	nids := newNIDs(ids)
	values := convertToUtilList(nids)
	return l.list().Has(values...)
}

func (l List[T]) At(i int) *ID[T] {
	if l == nil || i < 0 || i >= len(l) {
		return nil
	}
	return &l[i]
}

func (l List[T]) Index(id ID[T]) int {
	if l == nil || id.nid == nil {
		return -1
	}
	for i, lid := range l {
		if lid.nid != nil && lid.nid.Compare(id.nid) == 0 {
			return i
		}
	}
	return -1
}

func (l List[T]) Len() int {
	return len(l)
}

func (l List[T]) Ref() *List[T] {
	if l == nil {
		return nil
	}
	return &l
}

func (l List[T]) Refs() RefList[T] {
	if l == nil {
		return nil
	}
	refs := make([]*ID[T], len(l))
	for i := range l {
		refs[i] = &l[i]
	}
	return refs
}

func (l List[T]) Delete(ids ...ID[T]) List[T] {
	if l == nil {
		return nil
	}
	result := make(List[T], 0, len(l))
	for _, item := range l {
		keep := true
		for _, id := range ids {
			if id.nid != nil && item.nid != nil && item.nid.Compare(id.nid) == 0 {
				keep = false
				break
			}
		}
		if keep {
			result = append(result, item)
		}
	}
	return result
}

func (l List[T]) DeleteAt(i int) List[T] {
	if l == nil {
		return nil
	}

	// Convert util.List[nid] to []*nid
	convertedList := convertToNidSlice(l.list().DeleteAt(i))
	return nidsTo[T](convertedList)
}
func (l List[T]) Add(ids ...ID[T]) List[T] {
	if l == nil {
		return append(List[T]{}, ids...)
	}
	result := make(List[T], len(l)+len(ids))
	copy(result, l)
	copy(result[len(l):], ids)
	return result
}

func (l List[T]) AddUniq(ids ...ID[T]) List[T] {
	if l == nil {
		return append(List[T]{}, ids...)
	}
	result := l.Clone()
	for _, id := range ids {
		if !result.Has(id) {
			result = append(result, id)
		}
	}
	return result
}

func (l List[T]) Insert(i int, ids ...ID[T]) List[T] {
	nids := newNIDs(ids)
	values := make([]nid, len(nids))
	for i, n := range nids {
		values[i] = *n
	}
	inserted := l.list().Insert(i, values...)
	pointers := make([]*nid, len(inserted))
	for i, v := range inserted {
		pointers[i] = &v
	}
	return nidsTo[T](pointers)
}

func (l List[T]) Move(e ID[T], to int) List[T] {
	if l == nil {
		return nil
	}

	// Convert util.List[nid] to []*nid
	convertedList := convertToNidSlice(l.list().Move(*newNID(e), to))
	return nidsTo[T](convertedList)
}

func (l List[T]) MoveAt(from, to int) List[T] {
	if l == nil {
		return nil
	}

	// Convert util.List[nid] to []*nid
	convertedList := convertToNidSlice(l.list().MoveAt(from, to))
	return nidsTo[T](convertedList)
}

func (l List[T]) Reverse() List[T] {
	if l == nil {
		return nil
	}
	result := make(List[T], len(l))
	for i := range l {
		result[i] = l[len(l)-1-i]
	}
	return result
}

func (l List[T]) Concat(m List[T]) List[T] {
	if l == nil {
		if m == nil {
			return nil
		}
		return m.Clone()
	}
	result := make(List[T], len(l)+len(m))
	copy(result, l)
	copy(result[len(l):], m)
	return result
}

func (l List[T]) Intersect(m List[T]) List[T] {
	if l == nil {
		return nil
	}
	result := make(List[T], 0)
	for _, item := range l {
		if m.Has(item) {
			result = append(result, item)
		}
	}
	return result
}

func (l List[T]) Strings() []string {
	if l == nil {
		return nil
	}
	result := make([]string, len(l))
	for i, id := range l {
		result[i] = id.String()
	}
	return result
}

func (l List[T]) Clone() List[T] {
	if l == nil {
		return nil
	}
	result := make(List[T], len(l))
	for i, id := range l {
		cloned := id
		if id.nid != nil {
			cloned.nid = id.nid.Clone()
		}
		result[i] = cloned
	}
	return result
}

func (l List[T]) Sort() List[T] {
	if l == nil {
		return nil
	}
	result := l.Clone()
	sort.Sort(result)
	return result
}

func (l RefList[T]) Deref() List[T] {
	if l == nil {
		return nil
	}
	result := make(List[T], 0, len(l))
	for _, id := range l {
		if id != nil && !id.IsNil() {
			result = append(result, *id)
		}
	}
	return result
}

func (l List[T]) Less(i, j int) bool {
	if l[i].nid == nil || l[j].nid == nil {
		return false
	}
	return l[i].nid.Compare(l[j].nid) < 0
}

func (l List[T]) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l List[T]) Set() *Set[T] {
	if l == nil {
		return nil
	}
	s := &Set[T]{
		m: make(map[string]ID[T]),
	}
	s.Add(l...)
	return s
}

// Helper function to convert util.List[nid] to []*nid
func convertToNidSlice(list util.List[nid]) []*nid {
	var result []*nid
	for _, item := range list {
		result = append(result, &item)
	}
	return result
}

func convertToUtilList(list []*nid) util.List[nid] {
	var result util.List[nid]
	for _, item := range list {
		result = append(result, *item)
	}
	return result
}
