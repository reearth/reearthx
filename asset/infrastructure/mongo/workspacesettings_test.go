package mongo

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/workspacesettings"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/reearth/reearthx/rerror"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceSettingsRepo_FindByID(t *testing.T) {
	wid := accountdomain.NewWorkspaceID()
	w := workspacesettings.New().ID(wid).Tiles(nil).Terrains(nil).MustBuild()

	initDB := mongotest.Connect(t)
	client := mongox.NewClientWithDatabase(initDB(t))
	r := NewWorkspaceSettings(
		client,
	).Filtered(repo.WorkspaceFilter{Readable: []accountdomain.WorkspaceID{wid}, Writable: []accountdomain.WorkspaceID{wid}})
	ctx := context.Background()

	err := r.Save(ctx, w.Clone())
	assert.NoError(t, err)

	got, err := r.FindByID(ctx, wid)
	assert.NoError(t, err)
	assert.Equal(t, w, got)
}

func TestWorkspaceSettingsRepo_FindByIDs(t *testing.T) {
	wid1 := accountdomain.NewWorkspaceID()
	w1 := workspacesettings.New().ID(wid1).Tiles(nil).Terrains(nil).MustBuild()
	wid2 := accountdomain.NewWorkspaceID()
	w2 := workspacesettings.New().ID(wid2).Tiles(nil).Terrains(nil).MustBuild()
	ids := accountdomain.WorkspaceIDList{wid1, wid2}

	initDB := mongotest.Connect(t)
	client := mongox.NewClientWithDatabase(initDB(t))
	r := NewWorkspaceSettings(
		client,
	).Filtered(repo.WorkspaceFilter{Readable: []accountdomain.WorkspaceID{wid1, wid2}, Writable: []accountdomain.WorkspaceID{wid1, wid2}})
	ctx := context.Background()

	err := r.Save(ctx, w1.Clone())
	assert.NoError(t, err)
	err = r.Save(ctx, w2.Clone())
	assert.NoError(t, err)

	want := workspacesettings.List{w1, w2}
	got, err := r.FindByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	ids2 := accountdomain.WorkspaceIDList{}
	got2, err2 := r.FindByIDs(ctx, ids2)
	assert.NoError(t, err2)
	assert.Nil(t, got2)
}

func TestWorkspaceSettingsRepo_Remove(t *testing.T) {
	wid1 := accountdomain.NewWorkspaceID()
	w1 := workspacesettings.New().ID(wid1).Tiles(nil).Terrains(nil).MustBuild()
	wid2 := accountdomain.NewWorkspaceID()
	w2 := workspacesettings.New().ID(wid2).Tiles(nil).Terrains(nil).MustBuild()

	initDB := mongotest.Connect(t)
	client := mongox.NewClientWithDatabase(initDB(t))
	r := NewWorkspaceSettings(
		client,
	).Filtered(repo.WorkspaceFilter{Readable: []accountdomain.WorkspaceID{wid1}, Writable: []accountdomain.WorkspaceID{wid1}})
	ctx := context.Background()

	err := r.Save(ctx, w1.Clone())
	assert.NoError(t, err)
	err = r.Remove(ctx, wid1)
	assert.NoError(t, err)
	_, err = r.FindByID(ctx, wid1)
	assert.ErrorIs(t, err, rerror.ErrNotFound)

	err = r.Save(ctx, w2.Clone())
	assert.ErrorIs(t, err, repo.ErrOperationDenied)
	err = r.Remove(ctx, wid2)
	assert.ErrorIs(t, err, repo.ErrOperationDenied)
}
