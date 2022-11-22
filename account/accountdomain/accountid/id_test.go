package accountid

import (
	"testing"

	"github.com/reearth/reearthx/idx"
	"github.com/stretchr/testify/assert"
)

type u struct{}

func (*u) Type() string { return "u" }

type tid = idx.ID[*u]

func TestNew(t *testing.T) {
	id := idx.New[*u]()
	assert.Equal(t, ID[*u]{domain: "example.com", id: id}, New(id, "example.com"))
}

func TestGenerate(t *testing.T) {
	id := Generate[*u]("example.com")
	assert.Equal(t, ID[*u]{domain: "example.com", id: id.ID()}, id)
}

func TestParse(t *testing.T) {
	id := idx.New[*u]()
	p, err := Parse[*u](id.String() + "@example.com")
	assert.NoError(t, err)
	assert.Equal(t, ID[*u]{domain: "example.com", id: id}, p)

	p, err = Parse[*u](id.String())
	assert.NoError(t, err)
	assert.Equal(t, ID[*u]{domain: "", id: id}, p)

	p, err = Parse[*u]("")
	assert.Equal(t, idx.ErrInvalidID, err)
	assert.Empty(t, p)
}

func TestMust(t *testing.T) {
	id := idx.New[*u]()
	assert.Equal(t, ID[*u]{domain: "example.com", id: id}, Must[*u](id.String()+"@example.com"))

	assert.Panics(t, func() {
		_ = Must[*u]("")
	})
}

func TestMethods(t *testing.T) {
	id := idx.New[*u]()
	aid := ID[*u]{domain: "example.com", id: id}
	assert.Equal(t, "example.com", aid.Domain())
	assert.True(t, aid.HasDomain())
	assert.Equal(t, id, aid.ID())
	assert.Equal(t, id.String()+"@example.com", aid.String())
	assert.Equal(t, "u:"+id.String()+"@example.com", aid.GoString())
	aid = ID[*u]{id: id}
	assert.Equal(t, id.String(), aid.String())
	assert.Equal(t, "u:"+id.String(), aid.GoString())
}
