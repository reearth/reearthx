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
	return util.List[nid](newNIDs(l))
}

func (l List[T]) Has(ids ...ID[T]) bool {
	return l.list().Has(newNIDs(ids)...)
}

func (l List[T]) At(i int) *ID[T] {
	return refNIDTo[T](l.list().At(i))
}

func (l List[T]) Index(id ID[T]) int {
	return l.list().Index(newNID(id))
}

func (l List[T]) Len() int {
	return l.list().Len()
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

	return refNIDsTo[T](lo.ToSlicePtr(newNIDs(l)))
}

func (l List[T]) Delete(ids ...ID[T]) List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](l.list().Delete(newNIDs(ids)...))
}

func (l List[T]) DeleteAt(i int) List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](l.list().DeleteAt(i))
}

func (l List[T]) Add(ids ...ID[T]) List[T] {
	return nidsTo[T](l.list().Add(newNIDs(ids)...))
}

func (l List[T]) AddUniq(ids ...ID[T]) List[T] {
	return nidsTo[T](l.list().AddUniq(newNIDs(ids)...))
}

func (l List[T]) Insert(i int, ids ...ID[T]) List[T] {
	return nidsTo[T](l.list().Insert(i, newNIDs(ids)...))
}

func (l List[T]) Move(e ID[T], to int) List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](l.list().Move(newNID(e), to))
}

func (l List[T]) MoveAt(from, to int) List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](l.list().MoveAt(from, to))
}

func (l List[T]) Reverse() List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](l.list().Reverse())
}

func (l List[T]) Concat(m List[T]) List[T] {
	return nidsTo[T](l.list().Concat(newNIDs(m)))
}

func (l List[T]) Intersect(m List[T]) List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](l.list().Intersect(newNIDs(m)))
}

func (l List[T]) Strings() []string {
	if l == nil {
		return nil
	}

	return util.Map(newNIDs(l), func(id nid) string {
		return id.String()
	})
}

func (l List[T]) Clone() List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](util.Map(newNIDs(l), func(id nid) nid {
		return id.Clone()
	}))
}

func (l List[T]) Sort() List[T] {
	sort.Sort(l)
	return l
}

func (l RefList[T]) Deref() List[T] {
	if l == nil {
		return nil
	}

	return nidsTo[T](util.FilterMap(newRefNIDs(l), func(id *nid) *nid {
		if id != nil && !(*id).IsNil() {
			return id
		}
		return nil
	}))
}

func (l List[T]) Less(i, j int) bool {
	return l[i].Compare(l[j]) < 0
}

func (l List[T]) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l List[T]) Set() *Set[T] {
	if l == nil {
		return nil
	}
	return NewSet[T](l...)
}
