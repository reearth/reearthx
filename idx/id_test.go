package idx

import (
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type TID = ID[T]

// T is a dummy ID type for unit tests
type T struct{}

func (T) Type() string { return "_" }

var dummyULID = mustParseID("01fzxycwmq7n84q8kessktvb8z")
var dummyID = TID{nid: &nid{id: dummyULID}}

func TestNew(t *testing.T) {
	id := New[T]()
	assert.Equal(t, TID{nid: id.nid}, id)
	assert.NotNil(t, id.nid)
	assert.NotZero(t, id.nid.id)
}

func TestNewAll(t *testing.T) {
	ids := NewAll[T](2)
	assert.Equal(t, List[T]{{nid: ids[0].nid}, {nid: ids[1].nid}}, ids)
	assert.NotNil(t, ids[0].nid)
	assert.NotNil(t, ids[1].nid)
	assert.NotZero(t, ids[0].nid.id)
	assert.NotZero(t, ids[1].nid.id)
}

func TestFrom(t *testing.T) {
	got, err := From[T]("01fzxycwmq7n84q8kessktvb8z")
	assert.NoError(t, err)
	assert.Equal(t, dummyID, got)

	got, err = From[T]("01f")
	assert.Same(t, ErrInvalidID, err)
	assert.Zero(t, got)
}

func TestMust(t *testing.T) {
	assert.Equal(t, dummyID, Must[T]("01fzxycwmq7n84q8kessktvb8z"))
	assert.Panics(t, func() {
		_ = Must[T]("xxx")
	})
}

func TestFromRef(t *testing.T) {
	assert.Equal(t, &dummyID, FromRef[T](lo.ToPtr("01fzxycwmq7n84q8kessktvb8z")))
	assert.Nil(t, FromRef[T](lo.ToPtr("xxx")))
	assert.Nil(t, FromRef[T](nil))
}

func TestID_Ref(t *testing.T) {
	id := dummyID
	ref := id.Ref()
	assert.Equal(t, &id, ref)
}

func TestID_Clone(t *testing.T) {
	clone := dummyID.Clone()
	assert.Equal(t, dummyID, clone)
	assert.NotSame(t, dummyID.nid, clone.nid)
}

func TestID_CloneRef(t *testing.T) {
	id := &dummyID
	ref := id.CloneRef()
	assert.Equal(t, id, ref)
	assert.NotSame(t, id, ref)
	assert.Nil(t, (*TID)(nil).CloneRef())
}

func TestID_Type(t *testing.T) {
	id := &TID{}
	assert.Equal(t, "_", id.Type())
}

func TestID_String(t *testing.T) {
	assert.Equal(t, "01fzxycwmq7n84q8kessktvb8z", dummyID.String())
	assert.Equal(t, "", (&TID{}).String())
}

func TestID_GoString(t *testing.T) {
	id := &dummyID
	assert.Equal(t, "_ID(01fzxycwmq7n84q8kessktvb8z)", id.GoString())
	assert.Equal(t, "_ID()", (&TID{}).GoString())
}

func TestID_Text(t *testing.T) {
	id := ID[T]{nid: &nid{}}
	assert.NoError(t, id.UnmarshalText([]byte(`01fzxycwmq7n84q8kessktvb8z`)))
	assert.Equal(t, dummyID, id)
	got, err := id.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`01fzxycwmq7n84q8kessktvb8z`), got)
}

func TestID_JSON(t *testing.T) {
	id := ID[T]{nid: &nid{}}
	assert.NoError(t, json.Unmarshal([]byte(`"01fzxycwmq7n84q8kessktvb8z"`), &id))
	assert.Equal(t, dummyID, id)
	got, err := json.Marshal(&id)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"01fzxycwmq7n84q8kessktvb8z"`), got)
}
