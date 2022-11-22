package accountid

import (
	"strings"

	"github.com/reearth/reearthx/idx"
)

type ID[T idx.Type] struct {
	domain string
	id     idx.ID[T]
}

func New[T idx.Type](id idx.ID[T], domain string) ID[T] {
	return ID[T]{
		domain: domain,
		id:     id,
	}
}

func Generate[T idx.Type](domain string) ID[T] {
	return ID[T]{
		domain: domain,
		id:     idx.New[T](),
	}
}

func Parse[T idx.Type](uid string) (i ID[T], err error) {
	rid, domain, _ := strings.Cut(uid, "@")

	var id idx.ID[T]
	id, err = idx.From[T](rid)
	if err != nil {
		return
	}

	return ID[T]{
		domain: domain,
		id:     id,
	}, nil
}

func Must[T idx.Type](id string) ID[T] {
	p, err := Parse[T](id)
	if err != nil {
		panic(err)
	}
	return p
}

func (i ID[T]) Domain() string {
	return i.domain
}

func (i ID[T]) HasDomain() bool {
	return i.domain != ""
}

func (i ID[T]) ID() idx.ID[T] {
	return i.id
}

func (i ID[T]) IsEmpty() bool {
	return i.id.IsEmpty()
}

func (i ID[T]) Compare(i2 ID[T]) int {
	return i.id.Compare(i2.id)
}

func (i ID[T]) String() string {
	sb := strings.Builder{}
	_, _ = sb.WriteString(i.id.String())
	if i.domain != "" {
		_, _ = sb.WriteString("@")
		_, _ = sb.WriteString(i.domain)
	}
	return sb.String()
}

func (i ID[T]) GoString() string {
	sb := strings.Builder{}
	_, _ = sb.WriteString(i.id.Type())
	_, _ = sb.WriteString(":")
	_, _ = sb.WriteString(i.String())
	return sb.String()
}
