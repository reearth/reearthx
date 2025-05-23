package util

import (
	"slices"

	"github.com/samber/lo"
)

type List[T comparable] []T

type Identifiable[ID comparable] interface {
	ID() ID
}

type IDLister[ID comparable] interface {
	LayerCount() int
	Layers() []ID
}

type Converter[S any, T any] func(*S) *T

type ConverterValue[S any, T any] func(S) *T

func (l List[T]) Has(elements ...T) bool {
	return Any(elements, func(e T) bool {
		return slices.Contains(l, e)
	})
}

func (l List[T]) At(i int) *T {
	if len(l) == 0 || i < 0 || len(l) <= i {
		return nil
	}
	e := l[i]
	return &e
}

func (l List[T]) Index(e T) int {
	return slices.Index(l, e)
}

func (l List[T]) Len() int {
	return len(l)
}

func (l List[T]) Copy() List[T] {
	if l == nil {
		return nil
	}
	return slices.Clone(l)
}

func (l List[T]) Ref() *List[T] {
	if l == nil {
		return nil
	}
	return &l
}

func (l List[T]) Refs() []*T {
	return Map(l, func(e T) *T {
		return &e
	})
}

func (l List[T]) Delete(elements ...T) List[T] {
	if l == nil {
		return nil
	}
	m := l.Copy()
	for _, e := range elements {
		if j := l.Index(e); j >= 0 {
			m = slices.Delete[[]T](m, j, j+1)
		}
	}
	return m
}

func (l List[T]) DeleteAt(i int) List[T] {
	if l == nil {
		return nil
	}
	m := l.Copy()
	return slices.Delete(m, i, i+1)
}

func (l List[T]) Add(elements ...T) List[T] {
	res := l.Copy()
	for _, e := range elements {
		res = append(res, e)
	}
	return res
}

func (l List[T]) AddUniq(elements ...T) List[T] {
	res := append(List[T]{}, l...)
	for _, id := range elements {
		if !res.Has(id) {
			res = append(res, id)
		}
	}
	return res
}

func (l List[T]) Insert(i int, elements ...T) List[T] {
	if i < 0 || len(l) < i {
		return l.Add(elements...)
	}
	return slices.Insert(l, i, elements...)
}

func (l List[T]) Move(e T, to int) List[T] {
	return l.MoveAt(l.Index(e), to)
}

func (l List[T]) MoveAt(from, to int) List[T] {
	if from < 0 || from == to || len(l) <= from {
		return l.Copy()
	}
	e := l[from]
	if from < to {
		to--
	}
	m := l.DeleteAt(from)
	if to < 0 {
		return m
	}
	return m.Insert(to, e)
}

func (l List[T]) Reverse() List[T] {
	//nolint:staticcheck
	return lo.Reverse(l.Copy())
}

func (l List[T]) Concat(m []T) List[T] {
	return append(l, m...)
}

func (l List[T]) Intersect(m []T) List[T] {
	if l == nil {
		return nil
	}
	return lo.Intersect(m, l)
}

func Last[T any](list []*T) *T {
	if len(list) == 0 {
		return nil
	}
	return list[len(list)-1]
}

func ExtractIDs[ID comparable, T Identifiable[ID]](list []*T) []ID {
	if len(list) == 0 {
		return nil
	}
	ids := make([]ID, 0, len(list))
	for _, item := range Deref(list, false) {
		ids = append(ids, item.ID())
	}
	return ids
}

func Pick[ID comparable, T Identifiable[ID]](list []*T, idList IDLister[ID]) []*T {
	if idList == nil || idList.LayerCount() == 0 {
		return nil
	}

	layers := make([]*T, 0, idList.LayerCount())
	for _, lid := range idList.Layers() {
		if l := Find(list, lid); l != nil {
			layers = append(layers, l)
		}
	}
	return layers
}

func Find[ID comparable, T Identifiable[ID]](list []*T, lid ID) *T {
	for _, item := range list {
		if item == nil {
			continue
		}
		if (*item).ID() == lid {
			return item
		}
	}
	return nil
}

func Deref[T any](list []*T, skipNil bool) []T {
	if !skipNil && list == nil {
		return nil
	}
	res := make([]T, 0, len(list))
	for _, item := range list {
		if item == nil {
			if !skipNil {
				var zeroValue T
				res = append(res, zeroValue)
			}
			continue
		}
		res = append(res, *item)
	}
	return res
}

func MapAdd[ID comparable, T Identifiable[ID]](m map[ID]*T, items ...*T) map[ID]*T {
	if m == nil {
		m = map[ID]*T{}
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		m[(*item).ID()] = item
	}
	return m
}

func ListMap[ID comparable, T Identifiable[ID]](list []*T) map[ID]*T {
	m := make(map[ID]*T, len(list))
	MapAdd(m, list...)
	return m
}

func MapWithIDFunc[ID comparable, T any](list []*T, idFunc func(*T) ID, checkNil bool) map[ID]*T {
	if checkNil && list == nil {
		return nil
	}
	m := make(map[ID]*T, len(list))
	for _, item := range list {
		if item != nil {
			id := idFunc(item)
			m[id] = item
		}
	}
	return m
}

func Merge[ID comparable, T Identifiable[ID]](m map[ID]*T, m2 map[ID]*T) map[ID]*T {
	if m == nil {
		return Clone(m2)
	}
	m3 := Clone(m)
	if m2 == nil {
		return m3
	}

	return MapAdd(m3, MapList(m2, false)...)
}

func ListMerge[T comparable](list []T, list2 []T, getClone func(T) T, duplicateSkip bool) []T {
	result := make([]T, 0, len(list)+len(list2))

	for _, item := range list {
		result = append(result, getClone(item))
	}

	for _, item := range list2 {
		if duplicateSkip {
			if !Contains(result, item) {
				result = append(result, getClone(item))
			}
		} else {
			result = append(result, getClone(item))
		}
	}

	return result
}

func MapList[ID comparable, T any](m map[ID]*T, skipNil bool) []*T {
	if m == nil {
		return nil
	}
	list := make([]*T, 0, len(m))
	for _, l := range m {
		if !skipNil || l != nil {
			list = append(list, l)
		}
	}
	return list
}

func Clone[ID comparable, T any](m map[ID]*T) map[ID]*T {
	if m == nil {
		return map[ID]*T{}
	}
	m2 := make(map[ID]*T, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func ListClone[T any](list []T, getClone func(T) T) []T {
	if list == nil {
		return nil
	}
	list2 := make([]T, len(list))
	for i, item := range list {
		list2[i] = getClone(item)
	}
	return list2
}

func Remove[ID comparable, T Identifiable[ID]](list []*T, idsToRemove ...ID) []*T {
	if list == nil {
		return nil
	}
	if len(list) == 0 {
		return []*T{}
	}

	result := make([]*T, 0, len(list))
	for _, item := range list {
		remove := false
		for _, id := range idsToRemove {
			if (*item).ID() == id {
				remove = true
				break
			}
		}
		if !remove {
			result = append(result, item)
		}
	}
	return result
}

func AddUnique[ID comparable, T Identifiable[ID]](list []*T, newList []*T) []*T {
	res := append([]*T{}, list...)

	for _, l := range newList {
		if l == nil {
			continue
		}
		if Find(res, (*l).ID()) != nil {
			continue
		}
		res = append(res, l)
	}

	return res
}

func MapPick[ID comparable, T Identifiable[ID]](m map[ID]*T, idList IDLister[ID]) []*T {
	if idList == nil || idList.LayerCount() == 0 {
		return nil
	}

	layers := make([]*T, 0, idList.LayerCount())
	for _, lid := range idList.Layers() {
		if l := m[lid]; l != nil {
			layers = append(layers, l)
		}
	}
	return layers
}

func ExtractKeys[ID comparable, T any](m map[ID]*T) []ID {
	keys := make([]ID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ToGenericList[S any, T any](list []*S, converter Converter[S, T]) []*T {
	res := make([]*T, 0, len(list))
	for _, l := range list {
		if li := converter(l); li != nil {
			res = append(res, li)
		}
	}
	return res
}

func ToGenericListValue[S any, T any](list []S, converter ConverterValue[S, T]) []*T {
	if len(list) == 0 {
		return nil
	}
	res := make([]*T, 0, len(list))
	for _, l := range list {
		if li := converter(l); li != nil {
			res = append(res, li)
		}
	}
	return res
}

func ListHas[ID comparable, T any](list []*T, getId func(*T) ID, id ID) bool {
	for _, item := range list {
		if getId(item) == id {
			return true
		}
	}
	return false
}

func Get[ID comparable, T any](list []*T, getId func(*T) ID, id ID) *T {
	for _, item := range list {
		if getId(item) == id {
			return item
		}
	}
	return nil
}

func RemoveById[ID comparable, T any](list []*T, getId func(*T) ID, id ID) []*T {
	for index, item := range list {
		if getId(item) == id {
			list = append(list[:index], list[index+1:]...)
			return list
		}
	}
	return list
}

func RemoveByIds[ID comparable, T any](list []*T, getId func(*T) ID, ids ...ID) []*T {
	result := make([]*T, 0, len(list))
	for _, item := range list {
		itemID := getId(item)
		if !Contains(ids, itemID) {
			result = append(result, item)
		}
	}
	return result
}

func Contains[ID comparable](ids []ID, id ID) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}

func Properties[ID comparable, T any](list []*T, getProperty func(*T) ID) []ID {
	if list == nil {
		return nil
	}
	ids := make([]ID, 0, len(list))
	for _, item := range list {
		if item != nil {
			ids = append(ids, getProperty(item))
		}
	}
	return ids
}

func ListFilter[ID comparable, T any](list []T, id ID, getId func(T) ID) []T {
	if len(list) == 0 {
		return nil
	}
	res := make([]T, 0, len(list))
	for _, item := range list {
		if getId(item) == id {
			res = append(res, item)
		}
	}
	return res
}

func IndexOf[ID comparable, T any](list []*T, getId func(*T) ID, id ID) int {
	for index, item := range list {
		if getId(item) == id {
			return index
		}
	}
	return -1
}
