package memory

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/group"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/reearth/reearthx/util"
)

type Group struct {
	err  error
	data *util.SyncMap[id.GroupID, *group.Group]
	now  *util.TimeNow
	f    repo.ProjectFilter
}

func NewGroup() repo.Group {
	return &Group{
		data: &util.SyncMap[id.GroupID, *group.Group]{},
		now:  &util.TimeNow{},
	}
}

func (r *Group) Filtered(filter repo.ProjectFilter) repo.Group {
	return &Group{
		data: r.data,
		f:    r.f.Merge(filter),
		now:  &util.TimeNow{},
	}
}

func (r *Group) FindByID(ctx context.Context, groupID id.GroupID) (*group.Group, error) {
	if r.err != nil {
		return nil, r.err
	}

	m := r.data.Find(func(k id.GroupID, m *group.Group) bool {
		return k == groupID && r.f.CanRead(m.Project())
	})

	if m != nil {
		return m, nil
	}
	return nil, rerror.ErrNotFound
}

func (r *Group) FindByIDs(ctx context.Context, list id.GroupIDList) (group.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	result := r.data.FindAll(func(k id.GroupID, m *group.Group) bool {
		return list.Has(k) && r.f.CanRead(m.Project())
	})

	return group.List(result).SortByID(), nil
}

func (r *Group) Filter(
	ctx context.Context,
	pid id.ProjectID,
	_ *group.Sort,
	_ *usecasex.Pagination,
) (group.List, *usecasex.PageInfo, error) {
	if r.err != nil {
		return nil, nil, r.err
	}

	// TODO: implement sort and pagination

	if !r.f.CanRead(pid) {
		return nil, nil, nil
	}

	result := group.List(r.data.FindAll(func(_ id.GroupID, m *group.Group) bool {
		return m.Project() == pid
	})).SortByID()

	return result, nil, nil
}

func (r *Group) FindByProject(ctx context.Context, pid id.ProjectID) (group.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	if !r.f.CanRead(pid) {
		return nil, nil
	}

	result := group.List(r.data.FindAll(func(_ id.GroupID, m *group.Group) bool {
		return m.Project() == pid
	})).SortByID()

	return result, nil
}

func (r *Group) FindByKey(ctx context.Context, pid id.ProjectID, key string) (*group.Group, error) {
	if r.err != nil {
		return nil, r.err
	}

	if !r.f.CanRead(pid) {
		return nil, nil
	}

	g := r.data.Find(func(_ id.GroupID, m *group.Group) bool {
		return m.Key().String() == key && m.Project() == pid
	})
	if g == nil {
		return nil, rerror.ErrNotFound
	}

	return g, nil
}

func (r *Group) FindByIDOrKey(
	ctx context.Context,
	pid id.ProjectID,
	g group.IDOrKey,
) (*group.Group, error) {
	if r.err != nil {
		return nil, r.err
	}

	groupID := g.ID()
	key := g.Key()
	if groupID == nil && (key == nil || *key == "") {
		return nil, rerror.ErrNotFound
	}

	m := r.data.Find(func(_ id.GroupID, m *group.Group) bool {
		return r.f.CanRead(m.Project()) &&
			(groupID != nil && m.ID() == *groupID || key != nil && m.Key().String() == *key)
	})
	if m == nil {
		return nil, rerror.ErrNotFound
	}

	return m, nil
}

func (r *Group) SaveAll(ctx context.Context, groups group.List) error {
	if r.err != nil {
		return r.err
	}
	if len(groups) == 0 {
		return nil
	}

	if !r.f.CanWrite(groups.Projects()...) {
		return repo.ErrOperationDenied
	}
	inp := map[id.GroupID]*group.Group{}
	for _, m := range groups {
		inp[m.ID()] = m
	}
	r.data.StoreAll(inp)
	return nil
}

func (r *Group) Save(ctx context.Context, g *group.Group) error {
	if r.err != nil {
		return r.err
	}

	if !r.f.CanWrite(g.Project()) {
		return repo.ErrOperationDenied
	}

	r.data.Store(g.ID(), g)
	return nil
}

func (r *Group) Remove(ctx context.Context, groupID id.GroupID) error {
	if r.err != nil {
		return r.err
	}

	if m, ok := r.data.Load(groupID); ok && r.f.CanWrite(m.Project()) {
		r.data.Delete(groupID)
		return nil
	}
	return rerror.ErrNotFound
}
