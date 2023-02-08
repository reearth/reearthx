package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceBuilder_ID(t *testing.T) {
	wid := NewID()
	b := NewWorkspace().ID(wid)
	assert.Equal(t, wid, b.w.id)
}

func TestWorkspaceBuilder_Members(t *testing.T) {
	m := map[UserID]Member{NewUserID(): {Role: RoleOwner}}
	b := NewWorkspace().Members(m)
	assert.Equal(t, m, b.members)
}

func TestWorkspaceBuilder_Name(t *testing.T) {
	w := NewWorkspace().Name("xxx")
	assert.Equal(t, "xxx", w.w.name)
}

func TestWorkspaceBuilder_NewID(t *testing.T) {
	b := NewWorkspace().NewID()
	assert.False(t, b.w.id.IsEmpty())
}

func TestWorkspaceBuilder_Build(t *testing.T) {
	m := map[UserID]Member{NewUserID(): {Role: RoleOwner}}
	id := NewID()
	b := NewWorkspace().ID(id).Name("a").Members(m)

	w, err := b.Build()
	assert.NoError(t, err)

	assert.Equal(t, &Workspace{
		id:      id,
		name:    "a",
		members: NewMembersWith(m),
	}, w)

	w, err = NewWorkspace().Build()
	assert.Equal(t, ErrInvalidID, err)
	assert.Nil(t, w)

}
