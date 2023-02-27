package accountmongo

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/reearth/reearthx/rerror"
	"github.com/stretchr/testify/assert"
)

func TestWorkspace_FindByID(t *testing.T) {
	ws := workspace.NewWorkspace().NewID().Name("hoge").MustBuild()
	tests := []struct {
		Name               string
		Input              accountdomain.WorkspaceID
		RepoData, Expected *workspace.Workspace
		WantErr            bool
	}{
		{
			Name:     "must find a workspace",
			Input:    ws.ID(),
			RepoData: ws,
			Expected: ws,
		},
		{
			Name:     "must not find any workspace",
			Input:    accountdomain.NewWorkspaceID(),
			RepoData: ws,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewWorkspace(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByID(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
			}
		})
	}
}

func TestWorkspace_FindByIDs(t *testing.T) {
	ws1 := workspace.NewWorkspace().NewID().Name("hoge").MustBuild()
	ws2 := workspace.NewWorkspace().NewID().Name("foo").MustBuild()
	ws3 := workspace.NewWorkspace().NewID().Name("xxx").MustBuild()

	tests := []struct {
		Name               string
		Input              accountdomain.WorkspaceIDList
		RepoData, Expected workspace.WorkspaceList
	}{
		{
			Name:     "must find users",
			RepoData: workspace.WorkspaceList{ws1, ws2},
			Input:    accountdomain.WorkspaceIDList{ws1.ID(), ws2.ID()},
			Expected: workspace.WorkspaceList{ws1, ws2},
		},
		{
			Name:     "must not find any user",
			Input:    accountdomain.WorkspaceIDList{ws3.ID()},
			RepoData: workspace.WorkspaceList{ws2, ws1},
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewWorkspace(client)
			ctx := context.Background()
			err := repo.SaveAll(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByIDs(ctx, tc.Input)
			assert.NoError(tt, err)
			for k, ws := range got {
				if ws != nil {
					assert.Equal(tt, tc.Expected[k].ID(), ws.ID())
					assert.Equal(tt, tc.Expected[k].Name(), ws.Name())
				}
			}
		})
	}
}

func TestWorkspace_FindByUser(t *testing.T) {
	u := user.New().Name("aaa").NewID().Email("aaa@bbb.com").MustBuild()
	ws := workspace.NewWorkspace().NewID().Name("hoge").Members(map[user.ID]workspace.Member{u.ID(): {Role: workspace.RoleOwner, InvitedBy: u.ID()}}).MustBuild()
	tests := []struct {
		Name     string
		Input    accountdomain.UserID
		RepoData *workspace.Workspace
		Expected workspace.WorkspaceList
	}{
		{
			Name:     "must find a workspace",
			Input:    u.ID(),
			RepoData: ws,
			Expected: workspace.WorkspaceList{ws},
		},
		{
			Name:     "must not find any workspace",
			Input:    user.NewID(),
			RepoData: ws,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewWorkspace(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByUser(ctx, tc.Input)
			assert.NoError(tt, err)
			for k, ws := range got {
				if ws != nil {
					assert.Equal(tt, tc.Expected[k].ID(), ws.ID())
					assert.Equal(tt, tc.Expected[k].Name(), ws.Name())
				}
			}
		})
	}
}

func TestWorkspace_Remove(t *testing.T) {
	ws := workspace.NewWorkspace().NewID().Name("hoge").MustBuild()

	init := mongotest.Connect(t)
	client := mongox.NewClientWithDatabase(init(t))

	repo := NewWorkspace(client)
	ctx := context.Background()
	err := repo.Save(ctx, ws)
	assert.NoError(t, err)

	err = repo.Remove(ctx, ws.ID())
	assert.NoError(t, err)
}

func TestWorkspace_RemoveAll(t *testing.T) {
	ws1 := workspace.NewWorkspace().NewID().Name("hoge").MustBuild()
	ws2 := workspace.NewWorkspace().NewID().Name("foo").MustBuild()

	init := mongotest.Connect(t)
	client := mongox.NewClientWithDatabase(init(t))

	repo := NewWorkspace(client)
	ctx := context.Background()
	err := repo.SaveAll(ctx, workspace.WorkspaceList{ws1, ws2})
	assert.NoError(t, err)

	err = repo.RemoveAll(ctx, accountdomain.WorkspaceIDList{ws1.ID(), ws2.ID()})
	assert.NoError(t, err)
}
