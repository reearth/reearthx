package accountdomain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceBuilder_ID(t *testing.T) {
	wid := GenerateWorkspaceID("")
	b := NewWorkspace().ID(wid)
	assert.Equal(t, wid, b.w.id)
}

func TestWorkspaceBuilder_Members(t *testing.T) {
	m := InitMembers(GenerateUserID(""))
	b := NewWorkspace().Members(m)
	assert.Equal(t, m, b.w.members)
}

func TestWorkspaceBuilder_Name(t *testing.T) {
	w := NewWorkspace().Name("xxx")
	assert.Equal(t, "xxx", w.w.name)
}

func TestWorkspaceBuilder_NewID(t *testing.T) {
	b := NewWorkspace().NewID("")
	assert.False(t, b.w.id.IsEmpty())
}

func TestWorkspaceBuilder_Build(t *testing.T) {
	m := InitMembers(GenerateUserID(""))
	id := GenerateWorkspaceID("")
	w, err := NewWorkspace().ID(id).Name("a").Members(m).Build()
	assert.NoError(t, err)
	assert.Equal(t, &Workspace{
		id:      id,
		name:    "a",
		members: m,
	}, w)

	w, err = NewWorkspace().Build()
	assert.Equal(t, ErrInvalidID, err)
	assert.Nil(t, w)

	w, err = NewWorkspace().ID(id).Build()
	assert.Equal(t, ErrMembersRequired, err)
	assert.Nil(t, w)

	w, err = NewWorkspace().ID(id).Members(&Members{}).Build()
	assert.Equal(t, ErrMembersRequired, err)
	assert.Nil(t, w)
}
