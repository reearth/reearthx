package accountmemory

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"
	"golang.org/x/exp/slices"
)

type Workspace struct {
	data *util.SyncMap[accountdomain.WorkspaceID, *workspace.Workspace]
	err  error
}

func NewWorkspace() accountrepo.Workspace {
	return &Workspace{
		data: &util.SyncMap[accountdomain.WorkspaceID, *workspace.Workspace]{},
	}
}

func (r *Workspace) FindByUser(ctx context.Context, i accountdomain.UserID) (workspace.WorkspaceList, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.FindAll(func(key accountdomain.WorkspaceID, value *workspace.Workspace) bool {
		return value.Members().HasUser(i)
	}), rerror.ErrNotFound)
}

func (r *Workspace) FindByIntegration(_ context.Context, i accountdomain.IntegrationID) (workspace.WorkspaceList, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.FindAll(func(key accountdomain.WorkspaceID, value *workspace.Workspace) bool {
		return value.Members().HasIntegration(i)
	}), rerror.ErrNotFound)
}

func (r *Workspace) FindByIDs(ctx context.Context, ids accountdomain.WorkspaceIDList) (workspace.WorkspaceList, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := r.data.FindAll(func(key accountdomain.WorkspaceID, value *workspace.Workspace) bool {
		return ids.Has(key)
	})
	slices.SortFunc(res, func(a, b *workspace.Workspace) bool { return a.ID().Compare(b.ID()) < 0 })
	return res, nil
}

func (r *Workspace) FindByID(ctx context.Context, v accountdomain.WorkspaceID) (*workspace.Workspace, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.Find(func(key accountdomain.WorkspaceID, value *workspace.Workspace) bool {
		return key == v
	}), rerror.ErrNotFound)
}

func (r *Workspace) Save(ctx context.Context, t *workspace.Workspace) error {
	if r.err != nil {
		return r.err
	}

	r.data.Store(t.ID(), t)
	return nil
}

func (r *Workspace) SaveAll(ctx context.Context, workspaces []*workspace.Workspace) error {
	if r.err != nil {
		return r.err
	}

	for _, t := range workspaces {
		r.data.Store(t.ID(), t)
	}
	return nil
}

func (r *Workspace) Remove(ctx context.Context, wid accountdomain.WorkspaceID) error {
	if r.err != nil {
		return r.err
	}

	r.data.Delete(wid)
	return nil
}

func (r *Workspace) RemoveAll(ctx context.Context, ids accountdomain.WorkspaceIDList) error {
	if r.err != nil {
		return r.err
	}

	for _, wid := range ids {
		r.data.Delete(wid)
	}
	return nil
}

func SetWorkspaceError(r accountrepo.Workspace, err error) {
	r.(*Workspace).err = err
}
