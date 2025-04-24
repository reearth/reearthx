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
var dummyID = TID{nid: nid{id: dummyULID}}

func TestNew(t *testing.T) {
	id := New[T]()
	assert.Equal(t, TID(id), id)
	assert.NotZero(t, id.nid)
}

func TestNewAll(t *testing.T) {
	ids := NewAll[T](2)
	assert.Equal(t, List[T]{{nid: ids[0].nid}, {nid: ids[1].nid}}, ids)
	assert.NotZero(t, ids[0].nid)
	assert.NotZero(t, ids[1].nid)
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
	ref := dummyID.Ref()
	assert.Equal(t, &dummyID, ref)
	assert.NotSame(t, &dummyID, ref)
}

func TestID_Clone(t *testing.T) {
	clone := dummyID.Clone()
	assert.Equal(t, dummyID, clone)
	assert.NotSame(t, &dummyID, &clone)
}

func TestID_CloneRef(t *testing.T) {
	cloneRef := dummyID.CloneRef()
	assert.Equal(t, &dummyID, cloneRef)
	assert.NotSame(t, &dummyID, cloneRef)
	assert.Nil(t, (*TID)(nil).CloneRef())
}

func TestID_Type(t *testing.T) {
	assert.Equal(t, "_", TID{}.Type())
}

func TestID_String(t *testing.T) {
	assert.Equal(t, "01fzxycwmq7n84q8kessktvb8z", dummyID.String())
	assert.Equal(t, "", ID[T]{}.String())
}

func TestID_GoString(t *testing.T) {
	assert.Equal(t, "_ID(01fzxycwmq7n84q8kessktvb8z)", dummyID.GoString())
	assert.Equal(t, "_ID()", TID{}.GoString())
}

func TestID_Text(t *testing.T) {
	var id TID
	assert.NoError(t, (&id).UnmarshalText([]byte(`01fzxycwmq7n84q8kessktvb8z`)))
	assert.Equal(t, dummyID, id)
	got, err := id.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`01fzxycwmq7n84q8kessktvb8z`), got)
}

func TestID_JSON(t *testing.T) {
	var id TID
	assert.NoError(t, json.Unmarshal([]byte(`"01fzxycwmq7n84q8kessktvb8z"`), &id))
	assert.Equal(t, dummyID, id)
	got, err := json.Marshal(id)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"01fzxycwmq7n84q8kessktvb8z"`), got)
}
