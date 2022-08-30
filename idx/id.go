package idx

import (
	"errors"

	"github.com/oklog/ulid"
	"github.com/samber/lo"
)

var ErrInvalidID = errors.New("invalid ID")

type Type interface {
	Type() string
}

type ID[T Type] struct {
	nid
}

func New[T Type]() ID[T] {
	return ID[T]{nid: nid{id: generateID()}}
}

func NewAll[T Type](n int) (l List[T]) {
	if n <= 0 {
		return
	}
	if n == 1 {
		return List[T]{New[T]()}
	}
	return nidsTo[T](lo.Map(generateAllID(n), func(id ulid.ULID, _ int) nid {
		return nid{id: id}
	}))
}

func From[T Type](id string) (ID[T], error) {
	parsedID, e := fromNID(id)
	if e != nil {
		return ID[T]{}, e
	}
	return ID[T]{nid: parsedID}, nil
}

func Must[T Type](id string) ID[T] {
	got, err := From[T](id)
	if err != nil {
		_ = lo.Must[any](nil, err)
	}
	return got
}

func FromRef[T Type](id *string) *ID[T] {
	if id == nil {
		return nil
	}
	nid, err := From[T](*id)
	if err != nil {
		return nil
	}
	return &nid
}

func (id ID[T]) Ref() *ID[T] {
	return &id
}

func (id ID[T]) Clone() ID[T] {
	return ID[T]{nid: id.nid.Clone()}
}

func (id *ID[T]) CloneRef() *ID[T] {
	if id == nil {
		return nil
	}
	i := id.Clone()
	return &i
}

func (ID[T]) Type() string {
	var t T
	return t.Type()
}

// GoString implements fmt.GoStringer interface.
func (id ID[T]) GoString() string {
	return id.Type() + "ID(" + id.String() + ")"
}

func (id ID[T]) Compare(id2 ID[T]) int {
	return id.nid.Compare(id2.nid)
}
