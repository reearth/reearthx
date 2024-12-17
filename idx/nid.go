package idx

import (
	"strings"
	"time"

	"github.com/oklog/ulid"
)

type nid struct {
	id ulid.ULID
}

func newNID[T Type](i ID[T]) *nid {
	return &nid{id: i.nid.id}
}

func nidTo[T Type](i *nid) ID[T] {
	if i == nil {
		return ID[T]{}
	}
	return ID[T]{nid: i}
}

func newNIDs[T Type](ids []ID[T]) []*nid {
	res := make([]*nid, 0, len(ids))
	for _, id := range ids {
		res = append(res, newNID(id))
	}
	return res
}

func nidsTo[T Type](ids []*nid) []ID[T] {
	res := make([]ID[T], 0, len(ids))
	for _, id := range ids {
		res = append(res, nidTo[T](id))
	}
	return res
}

func newRefNIDs[T Type](ids []*ID[T]) []*nid {
	res := make([]*nid, 0, len(ids))
	for _, id := range ids {
		var i *nid
		if id != nil {
			i = newNID(*id)
		}
		res = append(res, i)
	}
	return res
}

func refNIDsTo[T Type](ids []*nid) []*ID[T] {
	res := make([]*ID[T], 0, len(ids))
	for _, id := range ids {
		var i *ID[T]
		if id != nil {
			i2 := nidTo[T](id)
			i = &i2
		}
		res = append(res, i)
	}
	return res
}

func fromNID(id string) (*nid, error) {
	parsedID, e := parseID(id)
	if e != nil {
		return nil, ErrInvalidID
	}
	return &nid{id: parsedID}, nil
}

func refNIDTo[T Type](n *nid) *ID[T] {
	if n == nil {
		return nil
	}
	nid2 := nidTo[T](n)
	return &nid2
}

func (id *nid) Ref() *nid {
	if id == nil {
		return nil
	}
	return &nid{id: id.id}
}

func (id *nid) Clone() *nid {
	if id == nil {
		return nil
	}
	return &nid{id: id.id}
}

func (id *nid) Timestamp() time.Time {
	if id == nil {
		return time.Time{}
	}
	return ulid.Time(id.id.Time())
}

// String implements fmt.Stringer interface.
func (id *nid) String() string {
	if id == nil || id.IsEmpty() {
		return ""
	}
	return strings.ToLower(id.id.String())
}

func (id *nid) StringRef() *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// GoString implements fmt.GoStringer interface.
func (id *nid) GoString() string {
	if id == nil {
		return "ID(nil)"
	}
	return "ID(" + id.String() + ")"
}

func (id *nid) Compare(id2 *nid) int {
	if id == nil {
		if id2 == nil {
			return 0
		}
		return -1
	}
	if id2 == nil {
		return 1
	}
	return id.id.Compare(id2.id)
}

func (id *nid) IsEmpty() bool {
	return id == nil || id.id.Compare(ulid.ULID{}) == 0
}

func (id *nid) IsNil() bool {
	return id == nil || id.IsEmpty()
}

// MarshalText implements encoding.TextMarshaler interface
func (id *nid) MarshalText() ([]byte, error) {
	if id.IsNil() {
		return nil, nil
	}
	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface
func (id *nid) UnmarshalText(b []byte) (err error) {
	newID, err := fromNID(string(b))
	if err != nil {
		return err
	}
	*id = *newID
	return nil
}
