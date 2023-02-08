package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceList_FilterByID(t *testing.T) {
	tid1 := NewID()
	tid2 := NewID()
	t1 := &Workspace{id: tid1}
	t2 := &Workspace{id: tid2}

	assert.Equal(t, WorkspaceList{t1}, WorkspaceList{t1, t2}.FilterByID(tid1))
	assert.Equal(t, WorkspaceList{t2}, WorkspaceList{t1, t2}.FilterByID(tid2))
	assert.Equal(t, WorkspaceList{t1, t2}, WorkspaceList{t1, t2}.FilterByID(tid1, tid2))
	assert.Equal(t, WorkspaceList{}, WorkspaceList{t1, t2}.FilterByID(NewID()))
	assert.Equal(t, WorkspaceList(nil), WorkspaceList(nil).FilterByID(tid1))
}

func TestWorkspaceList_FilterByUserRole(t *testing.T) {
	uid := NewUserID()
	tid1 := NewID()
	tid2 := NewID()
	t1 := &Workspace{
		id: tid1,
		members: &Members{
			users: map[UserID]Member{
				uid: {Role: RoleReader},
			},
		},
	}
	t2 := &Workspace{
		id: tid2,
		members: &Members{
			users: map[UserID]Member{
				uid: {Role: RoleOwner},
			},
		},
	}

	assert.Equal(t, WorkspaceList{t1}, WorkspaceList{t1, t2}.FilterByUserRole(uid, RoleReader))
	assert.Equal(t, WorkspaceList{}, WorkspaceList{t1, t2}.FilterByUserRole(uid, RoleWriter))
	assert.Equal(t, WorkspaceList{t2}, WorkspaceList{t1, t2}.FilterByUserRole(uid, RoleOwner))
	assert.Equal(t, WorkspaceList(nil), WorkspaceList(nil).FilterByUserRole(uid, RoleOwner))
}

func TestWorkspaceList_FilterByIntegrationRole(t *testing.T) {
	iid := NewIntegrationID()
	tid1 := NewID()
	tid2 := NewID()
	t1 := &Workspace{
		id: tid1,
		members: &Members{
			integrations: map[IntegrationID]Member{
				iid: {Role: RoleReader},
			},
		},
	}
	t2 := &Workspace{
		id: tid2,
		members: &Members{
			integrations: map[IntegrationID]Member{
				iid: {Role: RoleWriter},
			},
		},
	}

	assert.Equal(t, WorkspaceList{t1}, WorkspaceList{t1, t2}.FilterByIntegrationRole(iid, RoleReader))
	assert.Equal(t, WorkspaceList{}, WorkspaceList{t1, t2}.FilterByIntegrationRole(iid, RoleOwner))
	assert.Equal(t, WorkspaceList{t2}, WorkspaceList{t1, t2}.FilterByIntegrationRole(iid, RoleWriter))
	assert.Equal(t, WorkspaceList(nil), WorkspaceList(nil).FilterByIntegrationRole(iid, RoleOwner))
}

func TestWorkspaceList_FilterByUserRoleIncluding(t *testing.T) {
	uid := NewUserID()
	tid1 := NewID()
	tid2 := NewID()
	t1 := &Workspace{
		id: tid1,
		members: &Members{
			users: map[UserID]Member{
				uid: {Role: RoleReader},
			},
		},
	}
	t2 := &Workspace{
		id: tid2,
		members: &Members{
			users: map[UserID]Member{
				uid: {Role: RoleOwner},
			},
		},
	}

	assert.Equal(t, WorkspaceList{t1, t2}, WorkspaceList{t1, t2}.FilterByUserRoleIncluding(uid, RoleReader))
	assert.Equal(t, WorkspaceList{t2}, WorkspaceList{t1, t2}.FilterByUserRoleIncluding(uid, RoleWriter))
	assert.Equal(t, WorkspaceList{t2}, WorkspaceList{t1, t2}.FilterByUserRoleIncluding(uid, RoleOwner))
	assert.Equal(t, WorkspaceList(nil), WorkspaceList(nil).FilterByUserRoleIncluding(uid, RoleOwner))
}

func TestWorkspaceList_FilterByIntegrationRoleIncluding(t *testing.T) {
	uid := NewIntegrationID()
	tid1 := NewID()
	tid2 := NewID()
	t1 := &Workspace{
		id: tid1,
		members: &Members{
			integrations: map[IntegrationID]Member{
				uid: {Role: RoleReader},
			},
		},
	}
	t2 := &Workspace{
		id: tid2,
		members: &Members{
			integrations: map[IntegrationID]Member{
				uid: {Role: RoleOwner},
			},
		},
	}

	assert.Equal(t, WorkspaceList{t1, t2}, WorkspaceList{t1, t2}.FilterByIntegrationRoleIncluding(uid, RoleReader))
	assert.Equal(t, WorkspaceList{t2}, WorkspaceList{t1, t2}.FilterByIntegrationRoleIncluding(uid, RoleWriter))
	assert.Equal(t, WorkspaceList{t2}, WorkspaceList{t1, t2}.FilterByIntegrationRoleIncluding(uid, RoleOwner))
	assert.Equal(t, WorkspaceList(nil), WorkspaceList(nil).FilterByIntegrationRoleIncluding(uid, RoleOwner))
}

func TestWorkspaceList_IDs(t *testing.T) {
	wid1 := NewID()
	wid2 := NewID()
	t1 := &Workspace{id: wid1}
	t2 := &Workspace{id: wid2}

	assert.Equal(t, []ID{wid1, wid2}, WorkspaceList{t1, t2}.IDs())
	assert.Equal(t, []ID{}, WorkspaceList{}.IDs())
	assert.Equal(t, []ID(nil), WorkspaceList(nil).IDs())
}
