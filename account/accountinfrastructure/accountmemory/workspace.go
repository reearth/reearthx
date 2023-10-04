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
	data *util.SyncMap[workspace.ID, *workspace.Workspace]
	err  error
}

func NewWorkspace() *Workspace {
	return &Workspace{
		data: &util.SyncMap[workspace.ID, *workspace.Workspace]{},
	}
}

func NewWorkspaceWith(workspaces ...*workspace.Workspace) *Workspace {
	r := NewWorkspace()
	for _, ws := range workspaces {
		r.data.Store(ws.ID(), ws)
	}
	return r
}

func (r *Workspace) FindByUser(ctx context.Context, i accountdomain.UserID) (workspace.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.FindAll(func(key workspace.ID, value *workspace.Workspace) bool {
		return value.Members().HasUser(i)
	}), rerror.ErrNotFound)
}

func (r *Workspace) FindByIntegration(_ context.Context, i workspace.IntegrationID) (workspace.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.FindAll(func(key workspace.ID, value *workspace.Workspace) bool {
		return value.Members().HasIntegration(i)
	}), rerror.ErrNotFound)
}

func (r *Workspace) FindByIDs(ctx context.Context, ids workspace.IDList) (workspace.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := r.data.FindAll(func(key workspace.ID, value *workspace.Workspace) bool {
		return ids.Has(key)
	})
	slices.SortFunc(res, func(a, b *workspace.Workspace) bool { return a.ID().Compare(b.ID()) < 0 })
	return res, nil
}

func (r *Workspace) FindByID(ctx context.Context, v workspace.ID) (*workspace.Workspace, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.Find(func(key workspace.ID, value *workspace.Workspace) bool {
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

func (r *Workspace) SaveAll(ctx context.Context, workspaces workspace.List) error {
	if r.err != nil {
		return r.err
	}

	for _, t := range workspaces {
		r.data.Store(t.ID(), t)
	}
	return nil
}

func (r *Workspace) Remove(ctx context.Context, wid workspace.ID) error {
	if r.err != nil {
		return r.err
	}

	r.data.Delete(wid)
	return nil
}

func (r *Workspace) RemoveAll(ctx context.Context, ids workspace.IDList) error {
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
