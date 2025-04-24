package accountinteractor

import (
	"context"
	"errors"
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/permittable"
	"github.com/reearth/reearthx/account/accountdomain/role"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmemory"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/idx"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWorkspace_Create(t *testing.T) {
	ctx := context.Background()

	db := accountmemory.New()

	u := user.New().NewID().Name("aaa").Email("aaa@bbb.com").Workspace(accountdomain.NewWorkspaceID()).MustBuild()
	_ = db.User.Save(ctx, u)
	workspaceUC := NewWorkspace(db, nil)
	op := &accountusecase.Operator{User: lo.ToPtr(u.ID())}
	ws, err := workspaceUC.Create(ctx, "workspace name", u.ID(), op)

	assert.NoError(t, err)
	assert.NotNil(t, ws)

	resultWorkspaces, _ := workspaceUC.Fetch(ctx, []workspace.ID{ws.ID()}, &accountusecase.Operator{
		ReadableWorkspaces: []workspace.ID{ws.ID()},
	})

	assert.NotNil(t, resultWorkspaces)
	assert.NotEmpty(t, resultWorkspaces)
	assert.Equal(t, resultWorkspaces[0].ID(), ws.ID())
	assert.Equal(t, resultWorkspaces[0].Name(), "workspace name")
	assert.Equal(t, workspace.IDList{resultWorkspaces[0].ID()}, op.OwningWorkspaces)

	// mock workspace error
	wantErr := errors.New("test")
	accountmemory.SetWorkspaceError(db.Workspace, wantErr)
	workspace2, err := workspaceUC.Create(ctx, "workspace name 2", u.ID(), op)
	assert.Nil(t, workspace2)
	assert.Equal(t, wantErr, err)
}

func TestWorkspace_Fetch(t *testing.T) {
	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.New().ID(id1).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).MustBuild()

	u := user.New().NewID().Name("aaa").Email("aaa@bbb.com").Workspace(id1).MustBuild()
	op := &accountusecase.Operator{
		User:               lo.ToPtr(u.ID()),
		ReadableWorkspaces: []workspace.ID{id1, id2},
	}

	tests := []struct {
		name  string
		seeds workspace.List
		args  struct {
			ids      []workspace.ID
			operator *accountusecase.Operator
		}
		want             workspace.List
		mockWorkspaceErr bool
		wantErr          error
	}{
		{
			name:  "Fetch 1 of 2",
			seeds: workspace.List{w1, w2},
			args: struct {
				ids      []workspace.ID
				operator *accountusecase.Operator
			}{
				ids:      []workspace.ID{id1},
				operator: op,
			},
			want:    workspace.List{w1},
			wantErr: nil,
		},
		{
			name:  "Fetch 2 of 2",
			seeds: workspace.List{w1, w2},
			args: struct {
				ids      []workspace.ID
				operator *accountusecase.Operator
			}{
				ids:      []workspace.ID{id1, id2},
				operator: op,
			},
			want:    workspace.List{w1, w2},
			wantErr: nil,
		},
		{
			name:  "Fetch 1 of 0",
			seeds: workspace.List{},
			args: struct {
				ids      []workspace.ID
				operator *accountusecase.Operator
			}{
				ids:      []workspace.ID{id1},
				operator: op,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name:  "Fetch 2 of 0",
			seeds: workspace.List{},
			args: struct {
				ids      []workspace.ID
				operator *accountusecase.Operator
			}{
				ids:      []workspace.ID{id1, id2},
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
			workspaceUC := NewWorkspace(db, nil)

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
	w1 := workspace.New().ID(id1).Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleReader}}).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).MustBuild()

	u := user.New().NewID().Name("aaa").Email("aaa@bbb.com").Workspace(id1).MustBuild()
	op := &accountusecase.Operator{
		User:               lo.ToPtr(u.ID()),
		ReadableWorkspaces: []workspace.ID{id1, id2},
	}

	tests := []struct {
		name  string
		seeds workspace.List
		args  struct {
			userID   user.ID
			operator *accountusecase.Operator
		}
		want             workspace.List
		mockWorkspaceErr bool
		wantErr          error
	}{
		{
			name:  "Fetch 1 of 2",
			seeds: workspace.List{w1, w2},
			args: struct {
				userID   user.ID
				operator *accountusecase.Operator
			}{
				userID:   userID,
				operator: op,
			},
			want:    workspace.List{w1},
			wantErr: nil,
		},
		{
			name:  "Fetch 1 of 0",
			seeds: workspace.List{},
			args: struct {
				userID   user.ID
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
			seeds: workspace.List{w2},
			args: struct {
				userID   user.ID
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
			workspaceUC := NewWorkspace(db, nil)

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
	w1 := workspace.New().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	w1Updated := workspace.New().ID(id1).Name("WW1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).Name("W2").MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.New().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleReader}}).MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []workspace.ID{id1, id2, id3},
		OwningWorkspaces:   []workspace.ID{id1},
	}

	tests := []struct {
		name  string
		seeds workspace.List
		args  struct {
			wId      workspace.ID
			newName  string
			operator *accountusecase.Operator
		}
		want             *workspace.Workspace
		wantErr          error
		mockWorkspaceErr bool
	}{
		{
			name:  "Update 1",
			seeds: workspace.List{w1, w2},
			args: struct {
				wId      workspace.ID
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
			seeds: workspace.List{},
			args: struct {
				wId      workspace.ID
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
			seeds: workspace.List{w3},
			args: struct {
				wId      workspace.ID
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
				wId      workspace.ID
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
			workspaceUC := NewWorkspace(db, nil)

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
	w1 := workspace.New().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).Name("W2").MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.New().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleReader}}).MustBuild()
	id4 := accountdomain.NewWorkspaceID()
	w4 := workspace.New().ID(id4).Name("W4").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()
	id5 := accountdomain.NewWorkspaceID()
	id6 := accountdomain.NewWorkspaceID()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []workspace.ID{id1, id2, id3},
		OwningWorkspaces:   []workspace.ID{id1, id4, id5, id6},
	}

	tests := []struct {
		name  string
		seeds workspace.List
		args  struct {
			wId      workspace.ID
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Workspace
	}{
		{
			name:  "Remove 1",
			seeds: workspace.List{w1, w2},
			args: struct {
				wId      workspace.ID
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
			seeds: workspace.List{w1, w2},
			args: struct {
				wId      workspace.ID
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
			seeds: workspace.List{w3},
			args: struct {
				wId      workspace.ID
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
			seeds: workspace.List{w4},
			args: struct {
				wId      workspace.ID
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
				wId      workspace.ID
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
			workspaceUC := NewWorkspace(db, nil)
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
	w1 := workspace.New().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).Name("W2").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.New().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()
	id4 := accountdomain.NewWorkspaceID()
	w4 := workspace.New().ID(id3).Name("W4").Members(map[user.ID]workspace.Member{accountdomain.NewUserID(): {Role: workspace.RoleOwner}}).Personal(true).MustBuild()

	u := user.New().NewID().Name("aaa").Email("a@b.c").MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []workspace.ID{id1, id2},
		OwningWorkspaces:   []workspace.ID{id1, id2, id3},
	}

	tests := []struct {
		name       string
		seeds      workspace.List
		usersSeeds []*user.User
		enforcer   WorkspaceMemberCountEnforcer
		args       struct {
			wId      workspace.ID
			users    map[user.ID]workspace.Role
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Members
	}{
		{
			name:       "add a member",
			seeds:      workspace.List{w2},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				users    map[user.ID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId: w2.ID(),
				users: map[user.ID]workspace.Role{
					u.ID(): workspace.RoleReader,
				},
				operator: op,
			},
			wantErr: nil,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID: {Role: workspace.RoleOwner},
				u.ID(): {Role: workspace.RoleReader, InvitedBy: userID}, // added
			}, nil, false),
		},
		{
			name:       "add a non existing member",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				users    map[user.ID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId: w1.ID(),
				users: map[user.ID]workspace.Role{
					accountdomain.NewUserID(): workspace.RoleReader,
				},
				operator: op,
			},
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID: {Role: workspace.RoleOwner},
			}, nil, false),
		},
		{
			name:       "add a mamber to personal workspace",
			seeds:      workspace.List{w3},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				users    map[user.ID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId: w3.ID(),
				users: map[user.ID]workspace.Role{
					u.ID(): workspace.RoleReader,
				},
				operator: op,
			},
			wantErr: workspace.ErrCannotModifyPersonalWorkspace,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID: {Role: workspace.RoleOwner},
			}, map[accountdomain.IntegrationID]workspace.Member{}, true),
		},
		{
			name:       "add member but enforcer rejects",
			seeds:      workspace.List{w2},
			usersSeeds: []*user.User{u},
			enforcer: func(_ context.Context, _ *workspace.Workspace, _ user.List, _ *accountusecase.Operator) error {
				return errors.New("test")
			},
			args: struct {
				wId      workspace.ID
				users    map[user.ID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId: w2.ID(),
				users: map[user.ID]workspace.Role{
					u.ID(): workspace.RoleReader,
				},
				operator: op,
			},
			wantErr: errors.New("test"),
		},
		{
			name:  "op denied",
			seeds: workspace.List{w4},
			args: struct {
				wId      workspace.ID
				users    map[user.ID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId: id4,
				users: map[user.ID]workspace.Role{
					accountdomain.NewUserID(): workspace.RoleReader,
				},
				operator: op,
			},
			wantErr:          accountinterfaces.ErrOperationDenied,
			mockWorkspaceErr: false,
		},
		{
			name: "mock error",
			args: struct {
				wId      workspace.ID
				users    map[user.ID]workspace.Role
				operator *accountusecase.Operator
			}{
				wId: id3,
				users: map[user.ID]workspace.Role{
					u.ID(): workspace.RoleReader,
				},
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
			workspaceUC := NewWorkspace(db, tc.enforcer)

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
	w1 := workspace.New().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	id3 := accountdomain.NewWorkspaceID()
	u := user.New().NewID().Name("aaa").Email("a@b.c").MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []workspace.ID{id1, id2},
		OwningWorkspaces:   []workspace.ID{id1, id2, id3},
	}

	iid1 := accountdomain.NewIntegrationID()

	tests := []struct {
		name       string
		seeds      workspace.List
		usersSeeds []*user.User
		args       struct {
			wId           workspace.ID
			integrationID accountdomain.IntegrationID
			role          workspace.Role
			operator      *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             []accountdomain.IntegrationID
	}{
		{
			name:       "add non existing",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId           workspace.ID
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
				wId           workspace.ID
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

			workspaceUC := NewWorkspace(db, nil)

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
	w1 := workspace.New().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).Name("W2").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}, u.ID(): {Role: workspace.RoleReader}}).Personal(false).MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.New().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()
	id4 := accountdomain.NewWorkspaceID()
	w4 := workspace.New().ID(id4).Name("W4").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []workspace.ID{id1, id2},
		OwningWorkspaces:   []workspace.ID{id1},
	}

	tests := []struct {
		name       string
		seeds      workspace.List
		usersSeeds []*user.User
		args       struct {
			wId      workspace.ID
			uId      user.ID
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Members
	}{
		{
			name:       "Remove non existing",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uId      user.ID
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
			seeds:      workspace.List{w2},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uId      user.ID
				operator *accountusecase.Operator
			}{
				wId:      id2,
				uId:      u.ID(),
				operator: op,
			},
			wantErr: nil,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, nil, false),
		},
		{
			name:       "Remove personal workspace",
			seeds:      workspace.List{w3},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uId      user.ID
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
			seeds:      workspace.List{w4},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uId      user.ID
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
				wId      workspace.ID
				uId      user.ID
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
			workspaceUC := NewWorkspace(db, nil)

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

func TestWorkspace_RemoveMultipleMembers(t *testing.T) {
	userID := accountdomain.NewUserID()
	userID2 := accountdomain.NewUserID()
	userID3 := accountdomain.NewUserID()
	userID4 := accountdomain.NewUserID()
	u := user.New().ID(userID).Name("aaa").Email("a@b.c").MustBuild()
	u2 := user.New().ID(userID2).Name("bbb").Email("b@b.c").MustBuild()
	u3 := user.New().ID(userID3).Name("ccc").Email("c@b.c").MustBuild()
	u4 := user.New().ID(userID4).Name("ddd").Email("d@b.c").MustBuild()

	id1 := accountdomain.NewWorkspaceID()
	w1 := workspace.New().ID(id1).Name("W1").
		Members(map[user.ID]workspace.Member{
			userID:  {Role: workspace.RoleOwner},
			userID2: {Role: workspace.RoleReader},
			userID3: {Role: workspace.RoleReader},
			userID4: {Role: workspace.RoleReader},
		}).Personal(false).MustBuild()

	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).Name("W2").
		Members(map[user.ID]workspace.Member{
			userID:  {Role: workspace.RoleOwner},
			userID2: {Role: workspace.RoleReader},
		}).Personal(true).MustBuild()

	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.New().ID(id3).Name("W3").
		Members(map[user.ID]workspace.Member{
			userID:  {Role: workspace.RoleOwner},
			userID2: {Role: workspace.RoleReader},
		}).Personal(false).MustBuild()

	id4 := accountdomain.NewWorkspaceID()
	w4 := workspace.New().ID(id4).Name("W4").
		Members(map[user.ID]workspace.Member{
			userID:  {Role: workspace.RoleOwner},
			userID2: {Role: workspace.RoleReader},
		}).Personal(false).MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []workspace.ID{id1},
		OwningWorkspaces:   []workspace.ID{id1, id2, id3, id4},
	}

	tests := []struct {
		name       string
		seeds      workspace.List
		usersSeeds []*user.User
		args       struct {
			wId      workspace.ID
			uIds     workspace.UserIDList
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Members
	}{
		{
			name:       "Remove non existing",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				wId:      id1,
				uIds:     workspace.UserIDList{accountdomain.NewUserID()},
				operator: op,
			},
			wantErr: workspace.ErrTargetUserNotInTheWorkspace,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}, map[accountdomain.IntegrationID]workspace.Member{}, false),
		},
		{
			name:       "Remove multiple existing members",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u, u2, u3, u4},
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				wId:      id1,
				uIds:     workspace.UserIDList{userID2, userID3},
				operator: op,
			},
			wantErr: nil,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID:  {Role: workspace.RoleOwner},
				userID4: {Role: workspace.RoleReader},
			}, nil, false),
		},
		{
			name:       "Invalid Operator",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u, u2, u3, u4},
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				wId:      id1,
				uIds:     workspace.UserIDList{userID2, userID3},
				operator: &accountusecase.Operator{},
			},
			wantErr: accountinterfaces.ErrInvalidOperator,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID:  {Role: workspace.RoleOwner},
				userID4: {Role: workspace.RoleReader},
			}, nil, false),
		},
		{
			name:       "Operation Denied",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u, u2, u3, u4},
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				wId:  id1,
				uIds: workspace.UserIDList{userID2},
				operator: &accountusecase.Operator{
					User:               &userID3,
					ReadableWorkspaces: []workspace.ID{id1},
				},
			},
			wantErr: accountinterfaces.ErrOperationDenied,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID:  {Role: workspace.RoleOwner},
				userID3: {Role: workspace.RoleReader},
				userID4: {Role: workspace.RoleReader},
			}, nil, false),
		},
		{
			name:       "Remove multiple members, cannot remove from personal workspace",
			seeds:      workspace.List{w2},
			usersSeeds: []*user.User{u, u2},
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				wId:      id2,
				uIds:     workspace.UserIDList{userID, userID2},
				operator: op,
			},
			wantErr: workspace.ErrCannotModifyPersonalWorkspace,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID:  {Role: workspace.RoleOwner},
				userID2: {Role: workspace.RoleReader},
			}, nil, false),
		},
		{
			name:       "Remove multiple members, cannot remove owner",
			seeds:      workspace.List{w3},
			usersSeeds: []*user.User{u, u2},
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				wId:      id3,
				uIds:     workspace.UserIDList{userID, userID2},
				operator: op,
			},
			wantErr: accountinterfaces.ErrOwnerCannotLeaveTheWorkspace,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID:  {Role: workspace.RoleOwner},
				userID2: {Role: workspace.RoleReader},
			}, nil, false),
		},
		{
			name:       "Remove multiple members, empty user id list",
			seeds:      workspace.List{w4},
			usersSeeds: []*user.User{u, u2},
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				wId:      id4,
				uIds:     workspace.UserIDList{},
				operator: op,
			},
			wantErr: workspace.ErrNoSpecifiedUsers,
			want: workspace.NewMembersWith(map[user.ID]workspace.Member{
				userID:  {Role: workspace.RoleOwner},
				userID2: {Role: workspace.RoleReader},
			}, nil, false),
		},
		{
			name: "mock error",
			args: struct {
				wId      workspace.ID
				uIds     workspace.UserIDList
				operator *accountusecase.Operator
			}{
				uIds:     workspace.UserIDList{userID2, userID3},
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
			workspaceUC := NewWorkspace(db, nil)

			got, err := workspaceUC.RemoveMultipleUserMembers(ctx, tc.args.wId, tc.args.uIds, tc.args.operator)
			if tc.wantErr != nil {
				assert.ErrorIs(t, tc.wantErr, err)
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
	w1 := workspace.New().ID(id1).Name("W1").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(false).MustBuild()
	id2 := accountdomain.NewWorkspaceID()
	w2 := workspace.New().ID(id2).Name("W2").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}, u.ID(): {Role: workspace.RoleReader}}).Personal(false).MustBuild()
	id3 := accountdomain.NewWorkspaceID()
	w3 := workspace.New().ID(id3).Name("W3").Members(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}}).Personal(true).MustBuild()

	op := &accountusecase.Operator{
		User:               &userID,
		ReadableWorkspaces: []workspace.ID{id1, id2},
		OwningWorkspaces:   []workspace.ID{id1, id2, id3},
	}

	tests := []struct {
		name       string
		seeds      workspace.List
		usersSeeds []*user.User
		args       struct {
			wId      workspace.ID
			uId      user.ID
			role     workspace.Role
			operator *accountusecase.Operator
		}
		wantErr          error
		mockWorkspaceErr bool
		want             *workspace.Members
	}{
		{
			name:       "Update non existing",
			seeds:      workspace.List{w1},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uId      user.ID
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
			seeds:      workspace.List{w2},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uId      user.ID
				role     workspace.Role
				operator *accountusecase.Operator
			}{
				wId:      id2,
				uId:      u.ID(),
				role:     workspace.RoleWriter,
				operator: op,
			},
			wantErr: nil,
			want:    workspace.NewMembersWith(map[user.ID]workspace.Member{userID: {Role: workspace.RoleOwner}, u.ID(): {Role: workspace.RoleWriter}}, nil, false),
		},
		{
			name:       "Update personal workspace",
			seeds:      workspace.List{w3},
			usersSeeds: []*user.User{u},
			args: struct {
				wId      workspace.ID
				uId      user.ID
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
				wId      workspace.ID
				uId      user.ID
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
			workspaceUC := NewWorkspace(db, nil)

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

// TODO: Delete this once the permission check migration is complete.
func TestWorkspace_AddMember_Migration(t *testing.T) {
	// prepare
	ctx := context.Background()

	uId1 := user.NewID()
	uId2 := user.NewID()
	uId3 := user.NewID()
	u1 := user.New().ID(uId1).Name("user1").Email("user1@test.com").MustBuild()
	u2 := user.New().ID(uId2).Name("user2").Email("user2@test.com").MustBuild()
	u3 := user.New().ID(uId3).Name("user3").Email("user3@test.com").MustBuild()

	roleOwner := workspace.Member{
		Role:      workspace.RoleOwner,
		InvitedBy: uId1,
	}

	wId1 := workspace.NewID()
	wId2 := workspace.NewID()
	wId3 := workspace.NewID()
	wId4 := workspace.NewID()
	w1 := workspace.New().ID(wId1).
		Name("w1").
		Members(map[idx.ID[accountdomain.User]]workspace.Member{
			uId1: roleOwner,
		}).
		MustBuild()
	w2 := workspace.New().ID(wId2).
		Name("w2").
		Members(map[idx.ID[accountdomain.User]]workspace.Member{
			uId1: roleOwner,
		}).
		MustBuild()
	w3 := workspace.New().ID(wId3).
		Name("w3").
		Members(map[idx.ID[accountdomain.User]]workspace.Member{
			uId1: roleOwner,
		}).
		MustBuild()
	w4 := workspace.New().ID(wId4).
		Name("w4").
		Members(map[idx.ID[accountdomain.User]]workspace.Member{
			uId1: roleOwner,
		}).
		MustBuild()

	users := map[user.ID]workspace.Role{
		uId2: workspace.RoleOwner,
		uId3: workspace.RoleOwner,
	}

	op := &accountusecase.Operator{
		User:               &uId1,
		ReadableWorkspaces: []workspace.ID{wId1, wId2, wId3, wId4},
		OwningWorkspaces:   []workspace.ID{wId1, wId2, wId3, wId4},
	}

	tests := []struct {
		name    string
		wId     workspace.ID
		setup   func(ctx context.Context, repos *accountrepo.Container)
		assert  func(t *testing.T, ctx context.Context, repos *accountrepo.Container)
		wantErr bool
	}{
		{
			name: "should create maintainer role and assign it to workspace users",
			wId:  wId1,
			setup: func(ctx context.Context, repos *accountrepo.Container) {
				userRepo := accountrepo.NewMultiUser(accountmemory.NewUserWith(u1, u2, u3))
				workspaceRepo := accountmemory.NewWorkspaceWith(w1)
				repos.User = userRepo
				repos.Workspace = workspaceRepo
			},
			assert: func(t *testing.T, ctx context.Context, repos *accountrepo.Container) {
				assertPermittablesAndRoles(t, ctx, repos, user.IDList{uId2, uId3})
			},
		},
		{
			name: "should not duplicate maintainer role when it already exists",
			wId:  wId2,
			setup: func(ctx context.Context, repos *accountrepo.Container) {
				existingRole, _ := role.New().NewID().Name("maintainer").Build()
				err := repos.Role.Save(ctx, *existingRole)
				if err != nil {
					t.Fatal(err)
				}

				userRepo := accountrepo.NewMultiUser(accountmemory.NewUserWith(u1, u2, u3))
				workspaceRepo := accountmemory.NewWorkspaceWith(w2)
				repos.User = userRepo
				repos.Workspace = workspaceRepo
			},
			assert: func(t *testing.T, ctx context.Context, repos *accountrepo.Container) {
				assertPermittablesAndRoles(t, ctx, repos, user.IDList{uId2, uId3})
			},
		},
		{
			name: "should not add maintainer role if user already has it",
			wId:  wId3,
			setup: func(ctx context.Context, repos *accountrepo.Container) {
				existingRole, _ := role.New().NewID().Name("maintainer").Build()
				err := repos.Role.Save(ctx, *existingRole)
				if err != nil {
					t.Fatal(err)
				}

				p, _ := permittable.New().
					NewID().
					UserID(uId2).
					RoleIDs([]accountdomain.RoleID{existingRole.ID()}).
					Build()
				err = repos.Permittable.Save(ctx, *p)
				if err != nil {
					t.Fatal(err)
				}

				userRepo := accountrepo.NewMultiUser(accountmemory.NewUserWith(u1, u2, u3))
				workspaceRepo := accountmemory.NewWorkspaceWith(w3)
				repos.User = userRepo
				repos.Workspace = workspaceRepo
			},
			assert: func(t *testing.T, ctx context.Context, repos *accountrepo.Container) {
				permittable, err := repos.Permittable.FindByUserID(ctx, uId2)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(permittable.RoleIDs()))

				assertPermittablesAndRoles(t, ctx, repos, user.IDList{uId2, uId3})
			},
		},
		{
			name: "should add maintainer role when user has other roles",
			wId:  wId4,
			setup: func(ctx context.Context, repos *accountrepo.Container) {
				otherRole, _ := role.New().NewID().Name("other_role").Build()
				err := repos.Role.Save(ctx, *otherRole)
				if err != nil {
					t.Fatal(err)
				}

				p, _ := permittable.New().
					NewID().
					UserID(uId2).
					RoleIDs([]accountdomain.RoleID{otherRole.ID()}).
					Build()
				err = repos.Permittable.Save(ctx, *p)
				if err != nil {
					t.Fatal(err)
				}

				userRepo := accountrepo.NewMultiUser(accountmemory.NewUserWith(u1, u2, u3))
				workspaceRepo := accountmemory.NewWorkspaceWith(w4)
				repos.User = userRepo
				repos.Workspace = workspaceRepo
			},
			assert: func(t *testing.T, ctx context.Context, repos *accountrepo.Container) {
				roles, err := repos.Role.FindAll(ctx)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(roles))

				permittable, err := repos.Permittable.FindByUserID(ctx, uId2)
				assert.NoError(t, err)
				assert.Equal(t, 2, len(permittable.RoleIDs()))

				assertPermittablesAndRoles(t, ctx, repos, user.IDList{uId2, uId3})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memoryRepo := accountmemory.New()

			if tt.setup != nil {
				tt.setup(ctx, memoryRepo)
			}

			enforcer := func(_ context.Context, _ *workspace.Workspace, _ user.List, _ *accountusecase.Operator) error {
				return nil
			}
			workspaceUC := NewWorkspace(memoryRepo, enforcer)

			_, err := workspaceUC.AddUserMember(ctx, tt.wId, users, op)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.assert != nil {
				tt.assert(t, ctx, memoryRepo)
			}
		})
	}
}

// TODO: Delete this once the permission check migration is complete.
func assertPermittablesAndRoles(t *testing.T, ctx context.Context, repos *accountrepo.Container, expectedUserIDs user.IDList) {
	// role
	roles, err := repos.Role.FindAll(ctx)
	assert.NoError(t, err)
	var maintainerRole *role.Role
	for _, r := range roles {
		if r.Name() == "maintainer" {
			if maintainerRole != nil {
				t.Fatal("maintainer role already exists")
			}
			maintainerRole = r
		}
	}
	assert.NotNil(t, maintainerRole)

	// permittable
	permittables, err := repos.Permittable.FindByUserIDs(ctx, expectedUserIDs)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedUserIDs), len(permittables))

	// userID
	userIds := make(user.IDList, 0, len(permittables))
	for _, p := range permittables {
		userIds = append(userIds, p.UserID())
	}
	for _, expectedID := range expectedUserIDs {
		assert.Contains(t, userIds, expectedID)
	}

	// role assignment
	for _, p := range permittables {
		assert.Contains(t, p.RoleIDs(), maintainerRole.ID())
	}
}
