package accountmemory

import (
	"context"

	slices0 "slices"

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

func (r *Workspace) Filtered(f accountrepo.WorkspaceFilter) accountrepo.Workspace {
	return r // TODO
}

func (r *Workspace) FindByUser(_ context.Context, i accountdomain.UserID) (workspace.List, error) {
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

// FindByIntegrations finds workspace list based on integrations IDs
func (r *Workspace) FindByIntegrations(_ context.Context, ids workspace.IntegrationIDList) (workspace.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	res := r.data.FindAll(func(key workspace.ID, value *workspace.Workspace) bool {
		return slices0.ContainsFunc(ids, value.Members().HasIntegration)
	})

	slices.SortFunc(res, func(a, b *workspace.Workspace) int { return a.ID().Compare(b.ID()) })

	return res, nil
}

func (r *Workspace) FindByIDs(_ context.Context, ids workspace.IDList) (workspace.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := r.data.FindAll(func(key workspace.ID, value *workspace.Workspace) bool {
		return ids.Has(key)
	})
	slices.SortFunc(res, func(a, b *workspace.Workspace) int {
		idA := a.ID()
		idB := b.ID()
		return idA.Compare(&idB)
	})

	return res, nil
}

func (r *Workspace) FindByID(_ context.Context, v workspace.ID) (*workspace.Workspace, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.Find(func(key workspace.ID, value *workspace.Workspace) bool {
		return key == v
	}), rerror.ErrNotFound)
}

func (r *Workspace) Create(_ context.Context, t *workspace.Workspace) error {
	if r.err != nil {
		return r.err
	}

	if _, ok := r.data.Load(t.ID()); ok {
		return rerror.ErrAlreadyExists
	}

	r.data.Store(t.ID(), t)
	return nil
}

func (r *Workspace) Save(_ context.Context, t *workspace.Workspace) error {
	if r.err != nil {
		return r.err
	}

	r.data.Store(t.ID(), t)
	return nil
}

func (r *Workspace) SaveAll(_ context.Context, workspaces workspace.List) error {
	if r.err != nil {
		return r.err
	}

	for _, t := range workspaces {
		r.data.Store(t.ID(), t)
	}
	return nil
}

func (r *Workspace) Remove(_ context.Context, wid workspace.ID) error {
	if r.err != nil {
		return r.err
	}

	r.data.Delete(wid)
	return nil
}

func (r *Workspace) RemoveAll(_ context.Context, ids workspace.IDList) error {
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
