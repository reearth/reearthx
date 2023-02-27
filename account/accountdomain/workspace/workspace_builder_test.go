package workspace

import (
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/idx"
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
	i := map[IntegrationID]Member{NewIntegrationID(): {Role: RoleOwner}}
	id := NewID()
	w, err := NewWorkspace().ID(id).Name("a").Integrations(i).Members(m).Build()
	assert.NoError(t, err)

	assert.Equal(t, &Workspace{
		id:      id,
		name:    "a",
		members: NewMembersWith(m, i, false),
	}, w)

	w, err = NewWorkspace().ID(id).Name("a").Build()
	assert.NoError(t, err)

	assert.Equal(t, &Workspace{
		id:   id,
		name: "a",
		members: &Members{
			users:        map[idx.ID[accountdomain.User]]Member{},
			integrations: map[idx.ID[accountdomain.Integration]]Member{},
		},
	}, w)

	w, err = NewWorkspace().Build()
	assert.Equal(t, ErrInvalidID, err)
	assert.Nil(t, w)
}

func TestWorkspaceBuilder_MustBuild(t *testing.T) {
	m := map[UserID]Member{NewUserID(): {Role: RoleOwner}}
	i := map[IntegrationID]Member{NewIntegrationID(): {Role: RoleOwner}}
	id := NewID()
	w := NewWorkspace().ID(id).Name("a").Integrations(i).Members(m).MustBuild()

	assert.Equal(t, &Workspace{
		id:      id,
		name:    "a",
		members: NewMembersWith(m, i, false),
	}, w)

	assert.Panics(t, func() { NewWorkspace().MustBuild() })
}

func TestWorkspaceBuilder_Integrations(t *testing.T) {
	i := map[IntegrationID]Member{NewIntegrationID(): {Role: RoleOwner}}
	assert.Equal(t, &WorkspaceBuilder{
		w:            &Workspace{},
		integrations: i,
	}, NewWorkspace().Integrations(i))
}

func TestWorkspaceBuilder_Personal(t *testing.T) {
	assert.Equal(t, &WorkspaceBuilder{
		w:        &Workspace{},
		personal: true,
	}, NewWorkspace().Personal(true))
}
