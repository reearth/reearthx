package idx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNID(t *testing.T) {
	id1 := New[T]()
	id2 := New[T]()
	ids := []ID[T]{id1, id2}
	nids := newNIDs(ids)
	assert.Equal(t, []*nid{newNID(id1), newNID(id2)}, nids)
	assert.Equal(t, ids, nidsTo[T](nids))
}

func TestNID_Text(t *testing.T) {
	id := &nid{}
	assert.NoError(t, id.UnmarshalText([]byte(`01fzxycwmq7n84q8kessktvb8z`)))
	assert.Equal(t, &nid{id: dummyULID}, id)
	got, err := id.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`01fzxycwmq7n84q8kessktvb8z`), got)

	// Test nil handling
	var nilID *nid
	got, err = nilID.MarshalText()
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestNID_JSON(t *testing.T) {
	id := &nid{}
	assert.NoError(t, json.Unmarshal([]byte(`"01fzxycwmq7n84q8kessktvb8z"`), id))
	assert.Equal(t, &nid{id: dummyULID}, id)
	got, err := json.Marshal(id)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"01fzxycwmq7n84q8kessktvb8z"`), got)

	// Test nil handling
	var nilID *nid
	got, err = json.Marshal(nilID)
	assert.NoError(t, err)
	assert.Equal(t, []byte("null"), got)
}
