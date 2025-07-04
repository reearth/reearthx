package memory

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/thread"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"
)

type Thread struct {
	err  error
	data *util.SyncMap[thread.ID, *thread.Thread]
	f    repo.WorkspaceFilter
}

func NewThread() repo.Thread {
	return &Thread{
		data: &util.SyncMap[id.ThreadID, *thread.Thread]{},
	}
}

func (r *Thread) Save(_ context.Context, th *thread.Thread) error {
	if r.err != nil {
		return r.err
	}

	if !r.f.CanWrite(th.Workspace()) {
		return repo.ErrOperationDenied
	}

	r.data.Store(th.ID(), th)
	return nil
}

func (r *Thread) SaveAll(_ context.Context, th thread.List) error {
	if r.err != nil {
		return r.err
	}

	for _, t := range th {
		if !r.f.CanWrite(t.Workspace()) {
			return repo.ErrOperationDenied
		}
	}
	r.data.StoreAll(th.ToMap())
	return nil
}

func (r *Thread) Filtered(f repo.WorkspaceFilter) repo.Thread {
	return &Thread{
		data: r.data,
		f:    r.f.Merge(f),
	}
}

func (r *Thread) FindByID(ctx context.Context, thid id.ThreadID) (*thread.Thread, error) {
	if r.err != nil {
		return nil, r.err
	}

	th := r.data.Find(func(k id.ThreadID, v *thread.Thread) bool {
		return k == thid && r.f.CanRead(v.Workspace())
	})

	if th != nil {
		return th, nil
	}
	return nil, rerror.ErrNotFound
}

func (r *Thread) FindByIDs(ctx context.Context, ids id.ThreadIDList) ([]*thread.Thread, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := thread.List(r.data.FindAll(func(key thread.ID, value *thread.Thread) bool {
		return ids.Has(key) && r.f.CanRead(value.Workspace())
	})).SortByID()
	return res, nil
}

func SetThreadError(r repo.Thread, err error) {
	r.(*Thread).err = err
}
