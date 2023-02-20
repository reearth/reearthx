package accountinteractor

import (
	"context"
	"errors"
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmemory"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWorkspace_Create(t *testing.T) {
	ctx := context.Background()

	db := accountmemory.New()

	u := user.New().NewID().Name("aaa").Email("aaa@bbb.com").Workspace(accountdomain.NewWorkspaceID()).MustBuild()
	workspaceUC := NewWorkspace(db)
	op := &accountusecase.Operator{User: lo.ToPtr(u.ID())}
	workspace, err := workspaceUC.Create(ctx, "workspace name", u.ID(), op)

	assert.NoError(t, err)
	assert.NotNil(t, workspace)

	resultWorkspaces, _ := workspaceUC.Fetch(ctx, []accountdomain.WorkspaceID{workspace.ID()}, &accountusecase.Operator{
		ReadableWorkspaces: []accountdomain.WorkspaceID{workspace.ID()},
	})

	assert.NotNil(t, resultWorkspaces)
	assert.NotEmpty(t, resultWorkspaces)
	assert.Equal(t, resultWorkspaces[0].ID(), workspace.ID())
	assert.Equal(t, resultWorkspaces[0].Name(), "workspace name")
	assert.Equal(t, accountdomain.WorkspaceIDList{resultWorkspaces[0].ID()}, op.OwningWorkspaces)

	// mock workspace error
	wantErr := errors.New("test")
	accountmemory.SetWorkspaceError(db.Workspace, wantErr)
	workspace2, err := workspaceUC.Create(ctx, "workspace name 2", u.ID(), op)
	assert.Nil(t, workspace2)
	assert.Equal(t, wantErr, err)
}

func TestWorkspace_Fetch(t *testing.T) {
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.NewWorkspace().ID(id2).MustBuild()

	u := user.New().NewID().Name("aaa").Email("aaa@bbb.com").Workspace(id1).MustBuild()
	op := &accountusecase.Operator{
		User:               lo.ToPtr(u.ID()),
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2},
	}

	tests := []struct {
		name  string
		seeds []*workspace.Workspace
		args  struct {
			ids      []accountdomain.WorkspaceID
			operator *accountusecase.Operator
		}
		want             []*workspace.Workspace
		mockWorkspaceErr bool
		wantErr          error
	}{
		{
			name:  "Fetch 1 of 2",
			seeds: []*workspace.Workspace{w1, w2},
			args: struct {
				ids      []accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				ids:      []accountdomain.WorkspaceID{id1},
				operator: op,
			},
			want:    []*workspace.Workspace{w1},
			wantErr: nil,
		},
		{
			name:  "Fetch 2 of 2",
			seeds: []*workspace.Workspace{w1, w2},
			args: struct {
				ids      []accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				ids:      []accountdomain.WorkspaceID{id1, id2},
				operator: op,
			},
			want:    []*workspace.Workspace{w1, w2},
			wantErr: nil,
		},
		{
			name:  "Fetch 1 of 0",
			seeds: []*workspace.Workspace{},
			args: struct {
				ids      []accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				ids:      []accountdomain.WorkspaceID{id1},
				operator: op,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name:  "Fetch 2 of 0",
			seeds: []*workspace.Workspace{},
			args: struct {
				ids      []accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				ids:      []accountdomain.WorkspaceID{id1, id2},
				operator: op,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name:             "mock error",
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			workspaceUC := NewWorkspace(db)

			got, err := workspaceUC.Fetch(ctx, tc.args.ids, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWorkspace_FindByUser(t *testing.T) {
	userID := accountdomain.NewUserID()
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleReader}}).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.NewWorkspace().ID(id2).MustBuild()

	u := user.New().NewID().Name("aaa").Email("aaa@bbb.com").Workspace(id1).MustBuild()
	op := &accountusecase.Operator{
		User:               lo.ToPtr(u.ID()),
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2},
	}

	tests := []struct {
		name  string
		seeds []*workspace.Workspace
		args  struct {
			userID   accountdomain.UserID
			operator *accountusecase.Operator
		}
		want             []*workspace.Workspace
		mockWorkspaceErr bool
		wantErr          error
	}{
		{
			name:  "Fetch 1 of 2",
			seeds: []*workspace.Workspace{w1, w2},
			args: struct {
				userID   accountdomain.UserID
				operator *accountusecase.Operator
			}{
				userID:   userID,
				operator: op,
			},
			want:    []*workspace.Workspace{w1},
			wantErr: nil,
		},
		{
			name:  "Fetch 1 of 0",
			seeds: []*workspace.Workspace{},
			args: struct {
				userID   accountdomain.UserID
				operator *accountusecase.Operator
			}{
				userID:   userID,
				operator: op,
			},
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name:  "Fetch 0 of 1",
			seeds: []*workspace.Workspace{w2},
			args: struct {
				userID   accountdomain.UserID
				operator *accountusecase.Operator
			}{
				userID:   userID,
				operator: op,
			},
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name:             "mock error",
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			workspaceUC := NewWorkspace(db)

			got, err := workspaceUC.FindByUser(ctx, tc.args.userID, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWorkspace_Update(t *testing.T) {
	userID := accountdomain.NewUserID()
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	w1Updated := workspace.NewWorkspace().ID(id1).Name("WW1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.NewWorkspace().ID(id2).Name("W2").MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.NewWorkspace().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleReader}}).MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2, id3},
		OwningWorkspaces:   []accountdomain.WorkspaceID{id1},
	}

	tests := []struct {
		name  string
		seeds []*workspace.Workspace
		args  struct {
			wId      accountdomain.WorkspaceID
			newName  string
			operator *accountusecase.Operator
		}
		want             *workspace.Workspace
		wantErr          error
		mockWorkspaceErr bool
	}{
		{
			name:  "Update 1",
			seeds: []*workspace.Workspace{w1, w2},
			args: struct {
				wId      accountdomain.WorkspaceID
				newName  string
				operator *accountusecase.Operator
			}{
				wId:      id1,
				newName:  "WW1",
				operator: op,
			},
			want:    w1Updated,
			wantErr: nil,
		},
		{
			name:  "Update 2",
			seeds: []*workspace.Workspace{},
			args: struct {
				wId      accountdomain.WorkspaceID
				newName  string
				operator *accountusecase.Operator
			}{
				wId:      id2,
				newName:  "WW2",
				operator: op,
			},
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name:  "Update 3",
			seeds: []*workspace.Workspace{w3},
			args: struct {
				wId      accountdomain.WorkspaceID
				newName  string
				operator *accountusecase.Operator
			}{
				wId:      id3,
				newName:  "WW3",
				operator: op,
			},
			want:    nil,
			wantErr: accountinterfaces.ErrOperationDenied,
		},
		{
			name: "mock error",
			args: struct {
				wId      accountdomain.WorkspaceID
				newName  string
				operator *accountusecase.Operator
			}{
				operator: op,
			},
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			workspaceUC := NewWorkspace(db)

			got, err := workspaceUC.Update(ctx, tc.args.wId, tc.args.newName, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
			got2, err := db.Workspace.FindByID(ctx, tc.args.wId)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got2)
		})
	}
}

func TestWorkspace_Remove(t *testing.T) {
	userID := accountdomain.NewUserID()
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.NewWorkspace().ID(id2).Name("W2").MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.NewWorkspace().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleReader}}).MustBuild()
	id4 := accountdomain.NewWorkspaceID()
	w4 := workspace.NewWorkspace().ID(id4).Name("W4").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()
	id5 := accountdomain.NewWorkspaceID()
	id6 := accountdomain.NewWorkspaceID()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2, id3},
		OwningWorkspaces:   []accountdomain.WorkspaceID{id1, id4, id5, id6},
	}

	tests := []struct {
		name  string
		seeds []*workspace.Workspace
		args  struct {
			wId      accountdomain.WorkspaceID
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Workspace
	}{
		{
			name:  "Remove 1",
			seeds: []*workspace.Workspace{w1, w2},
			args: struct {
				wId      accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				wId:      id1,
				operator: op,
			},
			wantErr: nil,
			want:    nil,
		},
		{
			name:  "Update 2",
			seeds: []*workspace.Workspace{w1, w2},
			args: struct {
				wId      accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				wId:      id2,
				operator: op,
			},
			wantErr: accountinterfaces.ErrOperationDenied,
			want:    w2,
		},
		{
			name:  "Update 3",
			seeds: []*workspace.Workspace{w3},
			args: struct {
				wId      accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				wId:      id3,
				operator: op,
			},
			wantErr: accountinterfaces.ErrOperationDenied,
			want:    w3,
		},
		{
			name:  "Remove 4",
			seeds: []*workspace.Workspace{w4},
			args: struct {
				wId      accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				wId:      id4,
				operator: op,
			},
			wantErr: workspace.ErrCannotModifyPersonalWorkspace,
			want:    w4,
		},
		{
			name: "mock workspace error",
			args: struct {
				wId      accountdomain.WorkspaceID
				operator *accountusecase.Operator
			}{
				wId:      id5,
				operator: op,
			},
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			workspaceUC := NewWorkspace(db)
			err := workspaceUC.Remove(ctx, tc.args.wId, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
			got, err := db.Workspace.FindByID(ctx, tc.args.wId)
			if tc.want == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWorkspace_AddMember(t *testing.T) {
	userID := accountdomain.NewUserID()
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.NewWorkspace().ID(id2).Name("W2").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.NewWorkspace().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()
	id4 := accountdomain.NewWorkspaceID()
	w4 := workspace.NewWorkspace().ID(id3).Name("W4").Members(map[user.ID]workspace.Member{accountdomain.NewUserID(): {Role: workspace.RoleOwner}}).Personal(true).MustBuild()

	u := user.New().NewID().Name("aaa").Email("a@b.c").MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2},
		OwningWorkspaces:   []accountdomain.WorkspaceID{id1, id2, id3},
	}

	tests := []struct {
		name       string
		seeds      []*workspace.Workspace
		usersSeeds []*user.User
		args       struct {
			wId      accountdomain.WorkspaceID
			users    map[accountdomain.UserID]workspace.Role
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Members
	}{
		{
			name:       "Add non existing",
			seeds:      []*workspace.Workspace{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				users    map[accountdomain.UserID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id1,
				users:    map[accountdomain.UserID]workspace.Role{accountdomain.NewUserID(): workspace.RoleReader},
				operator: op,
			},
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Add",
			seeds:      []*workspace.Workspace{w2},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				users    map[accountdomain.UserID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id2,
				users:    map[accountdomain.UserID]workspace.Role{u.ID(): workspace.RoleReader},
				operator: op,
			},
			wantErr: nil,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}, u.ID(): {Role: workspace.RoleReader, InvitedBy: userID}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Add to personal workspace",
			seeds:      []*workspace.Workspace{w3},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				users    map[accountdomain.UserID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id3,
				users:    map[accountdomain.UserID]workspace.Role{u.ID(): workspace.RoleReader},
				operator: op,
			},
			wantErr: workspace.ErrCannotModifyPersonalWorkspace,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, true),
		},
		{
			name:  "op denied",
			seeds: []*workspace.Workspace{w4},
			args: struct {
				wId      accountdomain.WorkspaceID
				users    map[accountdomain.UserID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id4,
				users:    map[accountdomain.UserID]workspace.Role{accountdomain.NewUserID(): workspace.RoleReader},
				operator: op,
			},
			wantErr:          accountinterfaces.ErrOperationDenied,
			mockWorkspaceErr: false,
		},
		{
			name: "mock error",
			args: struct {
				wId      accountdomain.WorkspaceID
				users    map[accountdomain.UserID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id3,
				users:    map[accountdomain.UserID]workspace.Role{u.ID(): workspace.RoleReader},
				operator: op,
			},
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			for _, p := range tc.usersSeeds {
				err := db.User.Save(ctx, p)
				assert.NoError(t, err)
			}
			workspaceUC := NewWorkspace(db)

			got, err := workspaceUC.AddUserMember(ctx, tc.args.wId, tc.args.users, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got.Members())

			got, err = db.Workspace.FindByID(ctx, tc.args.wId)
			if tc.want == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got.Members())
		})
	}
}

func TestWorkspace_AddIntegrationMember(t *testing.T) {
	userID := accountdomain.NewUserID()
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	id3 := accountdomain.NewWorkspaceID()
	u := user.New().NewID().Name("aaa").Email("a@b.c").MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2},
		OwningWorkspaces:   []accountdomain.WorkspaceID{id1, id2, id3},
	}

	iid1 := accountdomain.NewIntegrationID()

	tests := []struct {
		name       string
		seeds      []*workspace.Workspace
		usersSeeds []*user.User
		args       struct {
			wId           accountdomain.WorkspaceID
			integrationID accountdomain.IntegrationID
			role          workspace.Role
			operator      *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             []accountdomain.IntegrationID
	}{
		{
			name:       "Add non existing",
			seeds:      []*workspace.Workspace{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId           accountdomain.WorkspaceID
				integrationID accountdomain.IntegrationID
				role          workspace.Role
				operator      *accountusecase.Operator
			}{
				wId:           id1,
				integrationID: iid1,
				role:          workspace.RoleReader,
				operator:      op,
			},
			want: []accountdomain.IntegrationID{iid1},
		},
		{
			name: "mock error",
			args: struct {
				wId           accountdomain.WorkspaceID
				integrationID accountdomain.IntegrationID
				role          workspace.Role
				operator      *accountusecase.Operator
			}{
				wId:           id1,
				integrationID: iid1,
				role:          workspace.RoleReader,
				operator:      op,
			},
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			for _, p := range tc.usersSeeds {
				err := db.User.Save(ctx, p)
				assert.NoError(t, err)
			}

			workspaceUC := NewWorkspace(db)

			got, err := workspaceUC.AddIntegrationMember(ctx, tc.args.wId, tc.args.integrationID, tc.args.role, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got.Members().IntegrationIDs())
		})
	}
}

func TestWorkspace_RemoveMember(t *testing.T) {
	userID := accountdomain.NewUserID()
	u := user.New().NewID().Name("aaa").Email("a@b.c").MustBuild()
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.NewWorkspace().ID(id2).Name("W2").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}, u.ID(): {Role: workspace.RoleReader}}).Personal(false).MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.NewWorkspace().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()
	id4 := accountdomain.NewWorkspaceID()
	w4 := workspace.NewWorkspace().ID(id4).Name("W4").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2},
		OwningWorkspaces:   []accountdomain.WorkspaceID{id1},
	}

	tests := []struct {
		name       string
		seeds      []*workspace.Workspace
		usersSeeds []*user.User
		args       struct {
			wId      accountdomain.WorkspaceID
			uId      accountdomain.UserID
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Members
	}{
		{
			name:       "Remove non existing",
			seeds:      []*workspace.Workspace{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				operator *accountusecase.Operator
			}{
				wId:      id1,
				uId:      accountdomain.NewUserID(),
				operator: op,
			},
			wantErr: workspace.ErrTargetUserNotInTheWorkspace,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Remove",
			seeds:      []*workspace.Workspace{w2},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				operator *accountusecase.Operator
			}{
				wId:      id2,
				uId:      u.ID(),
				operator: op,
			},
			wantErr: nil,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Remove personal workspace",
			seeds:      []*workspace.Workspace{w3},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				operator *accountusecase.Operator
			}{
				wId:      id3,
				uId:      userID,
				operator: op,
			},
			wantErr: workspace.ErrCannotModifyPersonalWorkspace,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Remove single member",
			seeds:      []*workspace.Workspace{w4},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				operator *accountusecase.Operator
			}{
				wId:      id4,
				uId:      userID,
				operator: op,
			},
			wantErr: accountinterfaces.ErrOwnerCannotLeaveTheWorkspace,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name: "mock error",
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				operator *accountusecase.Operator
			}{operator: op},
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			for _, p := range tc.usersSeeds {
				err := db.User.Save(ctx, p)
				assert.NoError(t, err)
			}
			workspaceUC := NewWorkspace(db)

			got, err := workspaceUC.RemoveUserMember(ctx, tc.args.wId, tc.args.uId, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got.Members())

			got, err = db.Workspace.FindByID(ctx, tc.args.wId)
			if tc.want == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got.Members())
		})
	}
}

func TestWorkspace_UpdateMember(t *testing.T) {
	userID := accountdomain.NewUserID()
	u := user.New().NewID().Name("aaa").Email("a@b.c").MustBuild()
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.NewWorkspace().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.NewWorkspace().ID(id2).Name("W2").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}, u.ID(): {Role: workspace.RoleReader}}).Personal(false).MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.NewWorkspace().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []accountdomain.WorkspaceID{id1, id2},
		OwningWorkspaces:   []accountdomain.WorkspaceID{id1, id2, id3},
	}

	tests := []struct {
		name       string
		seeds      []*workspace.Workspace
		usersSeeds []*user.User
		args       struct {
			wId      accountdomain.WorkspaceID
			uId      accountdomain.UserID
			role     workspace.Role
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Members
	}{
		{
			name:       "Update non existing",
			seeds:      []*workspace.Workspace{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				role     workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id1,
				uId:      accountdomain.NewUserID(),
				role:     workspace.RoleWriter,
				operator: op,
			},
			wantErr: workspace.ErrTargetUserNotInTheWorkspace,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Update",
			seeds:      []*workspace.Workspace{w2},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				role     workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id2,
				uId:      u.ID(),
				role:     workspace.RoleWriter,
				operator: op,
			},
			wantErr: nil,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}, u.ID(): {Role: workspace.RoleWriter}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Update personal workspace",
			seeds:      []*workspace.Workspace{w3},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				role     workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id3,
				uId:      userID,
				role:     workspace.RoleReader,
				operator: op,
			},
			wantErr: workspace.ErrCannotModifyPersonalWorkspace,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, true),
		},
		{
			name: "mock error",
			args: struct {
				wId      accountdomain.WorkspaceID
				uId      accountdomain.UserID
				role     workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id3,
				operator: op,
			},
			wantErr:          errors.New("test"),
			mockWorkspaceErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := accountmemory.New()
			if tc.mockWorkspaceErr {
				accountmemory.SetWorkspaceError(db.Workspace, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Workspace.Save(ctx, p)
				assert.NoError(t, err)
			}
			for _, p := range tc.usersSeeds {
				err := db.User.Save(ctx, p)
				assert.NoError(t, err)
			}
			workspaceUC := NewWorkspace(db)

			got, err := workspaceUC.UpdateUserMember(ctx, tc.args.wId, tc.args.uId, tc.args.role, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got.Members())

			got, err = db.Workspace.FindByID(ctx, tc.args.wId)
			if tc.want == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got.Members())
		})
	}
}
