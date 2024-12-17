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

func (l List[T]) list() []*nid {
	return newNIDs(l)
}

func (l List[T]) Has(ids ...ID[T]) bool {
	if l == nil {
		return false
	}
	list := l.list()
	for _, id := range newNIDs(ids) {
		found := false
		for _, lid := range list {
			if lid.Compare(id) == 0 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (l List[T]) At(i int) *ID[T] {
	if l == nil || i < 0 || i >= len(l) {
		return nil
	}
	return &l[i]
}

func (l List[T]) Index(id ID[T]) int {
	if l == nil {
		return -1
	}
	for i, lid := range l {
		if lid.Compare(&id) == 0 {
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
			if item.Compare(&id) == 0 {
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
	if l == nil || i < 0 || i >= len(l) {
		return l
	}
	return append(l[:i], l[i+1:]...)
}

func (l List[T]) Add(ids ...ID[T]) List[T] {
	if l == nil {
		return append(List[T]{}, ids...)
	}
	return append(l, ids...)
}

func (l List[T]) AddUniq(ids ...ID[T]) List[T] {
	if l == nil {
		return append(List[T]{}, ids...)
	}
	result := l
	for _, id := range ids {
		if !result.Has(id) {
			result = append(result, id)
		}
	}
	return result
}

func (l List[T]) Insert(i int, ids ...ID[T]) List[T] {
	if l == nil {
		return ids
	}
	if i < 0 {
		i = 0
	}
	if i > len(l) {
		i = len(l)
	}
	result := make(List[T], len(l)+len(ids))
	copy(result, l[:i])
	copy(result[i:], ids)
	copy(result[i+len(ids):], l[i:])
	return result
}

func (l List[T]) Move(e ID[T], to int) List[T] {
	if l == nil {
		return nil
	}
	from := l.Index(e)
	if from < 0 {
		return l
	}
	return l.MoveAt(from, to)
}

func (l List[T]) MoveAt(from, to int) List[T] {
	if l == nil || from < 0 || from >= len(l) || to < 0 || to >= len(l) {
		return l
	}
	result := make(List[T], len(l))
	copy(result, l)
	item := result[from]
	if from < to {
		copy(result[from:], result[from+1:to+1])
	} else {
		copy(result[to+1:], result[to:from])
	}
	result[to] = item
	return result
}

func (l List[T]) Reverse() List[T] {
	if l == nil {
		return nil
	}
	result := make(List[T], len(l))
	for i, j := 0, len(l)-1; i <= j; i, j = i+1, j-1 {
		result[i], result[j] = l[j], l[i]
	}
	return result
}

func (l List[T]) Concat(m List[T]) List[T] {
	return append(l.Clone(), m...)
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
		result[i] = id.Clone()
	}
	return result
}

func (l List[T]) Sort() List[T] {
	sort.Sort(l)
	return l
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
	return l[i].Compare(&l[j]) < 0
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
