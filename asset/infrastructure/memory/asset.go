package memory

import (
	"context"
	"sort"
	"strings"

	"github.com/reearth/reearthx/account/accountdomain"

	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/usecase/gateway"
	"github.com/reearth/reearthx/asset/usecase/interfaces"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
)

type Asset struct {
	data            *util.SyncMap[asset.ID, *asset.Asset]
	err             error
	projectFilter   repo.ProjectFilter
	workspaceFilter repo.WorkspaceFilter
}

func NewAsset() repo.Asset {
	return &Asset{
		data: &util.SyncMap[id.AssetID, *asset.Asset]{},
	}
}

func (r *Asset) Filtered(f repo.ProjectFilter) repo.Asset {
	return &Asset{
		data:            r.data,
		projectFilter:   r.projectFilter.Merge(f),
		workspaceFilter: r.workspaceFilter,
	}
}

func (r *Asset) FindByID(_ context.Context, id id.AssetID) (*asset.Asset, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.Find(func(key asset.ID, value *asset.Asset) bool {
		return key == id && r.projectFilter.CanRead(value.Project())
	}), rerror.ErrNotFound)
}

func (r *Asset) FindByUUID(_ context.Context, uuid string) (*asset.Asset, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.Find(func(key asset.ID, value *asset.Asset) bool {
		return value.UUID() == uuid && r.projectFilter.CanRead(value.Project())
	}), rerror.ErrNotFound)
}

func (r *Asset) FindByIDs(_ context.Context, ids id.AssetIDList) (asset.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := asset.List(r.data.FindAll(func(key asset.ID, value *asset.Asset) bool {
		return ids.Has(key) && r.projectFilter.CanRead(value.Project())
	})).SortByID()
	return res, nil
}

func (r *Asset) Search(
	_ context.Context,
	id id.ProjectID,
	filter repo.AssetFilter,
) (asset.List, *usecasex.PageInfo, error) {
	if !r.projectFilter.CanRead(id) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	if r.err != nil {
		return nil, nil, r.err
	}

	result := asset.List(r.data.FindAll(func(_ asset.ID, v *asset.Asset) bool {
		// Base filter: project ID match
		if v.Project() != id {
			return false
		}

		// Keyword filter
		if filter.Keyword != nil && *filter.Keyword != "" {
			if !strings.Contains(strings.ToLower(v.FileName()), strings.ToLower(*filter.Keyword)) {
				return false
			}
		}
		// Content type filter can't be performed as it's not stored in memory

		return true
	})).SortByID()

	var startCursor, endCursor *usecasex.Cursor
	if len(result) > 0 {
		startCursor = lo.ToPtr(usecasex.Cursor(result[0].ID().String()))
		endCursor = lo.ToPtr(usecasex.Cursor(result[len(result)-1].ID().String()))
	}

	return result, usecasex.NewPageInfo(
		int64(len(result)),
		startCursor,
		endCursor,
		true,
		true,
	), nil
}

func (r *Asset) Save(_ context.Context, a *asset.Asset) error {
	if !r.projectFilter.CanWrite(a.Project()) {
		return repo.ErrOperationDenied
	}

	if r.err != nil {
		return r.err
	}

	r.data.Store(a.ID(), a)
	return nil
}

func (r *Asset) Delete(_ context.Context, id id.AssetID) error {
	if r.err != nil {
		return r.err
	}

	if a, ok := r.data.Load(id); ok && r.projectFilter.CanWrite(a.Project().Clone()) {
		r.data.Delete(id)
	}
	return nil
}

func (r *Asset) BatchDelete(_ context.Context, ids id.AssetIDList) error {
	if r.err != nil {
		return r.err
	}

	for _, aId := range ids {
		if a, ok := r.data.Load(aId); ok && r.projectFilter.CanWrite(a.Project().Clone()) {
			r.data.Delete(aId)
		}
	}
	return nil
}

func (r *Asset) RemoveByProjectWithFile(
	ctx context.Context,
	pid id.ProjectID,
	f gateway.File,
) error {
	r.data.FindAll(func(id id.AssetID, a *asset.Asset) bool {
		if r.workspaceFilter.CanWrite(a.Workspace()) {
			if a.Project() == pid {
				r.data.Delete(id)
			}
		}
		return true
	})
	return nil
}

func (r *Asset) FindByWorkspaceProject(
	_ context.Context,
	wid accountdomain.WorkspaceID,
	pid *id.ProjectID,
	filter repo.AssetFilter,
) ([]*asset.Asset, *usecasex.PageInfo, error) {
	if !r.workspaceFilter.CanRead(wid) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	result := r.data.FindAll(func(k id.AssetID, v *asset.Asset) bool {
		if pid != nil {
			return v.Project() == *pid && v.CoreSupport() &&
				(filter.Keyword == nil || strings.Contains(v.Name(), *filter.Keyword))
		}
		return v.Workspace() == wid && v.CoreSupport() &&
			(filter.Keyword == nil || strings.Contains(v.Name(), *filter.Keyword))
	})

	if filter.SortType != nil {
		s := *filter.SortType
		sort.SliceStable(result, func(i, j int) bool {
			if s == asset.SortTypeID {
				return result[i].ID().Compare(result[j].ID()) < 0
			}
			if s == asset.SortTypeSize {
				return result[i].Size() < result[j].Size()
			}
			if s == asset.SortTypeName {
				return strings.Compare(result[i].Name(), result[j].Name()) < 0
			}
			return false
		})
	}

	var startCursor, endCursor *usecasex.Cursor
	if len(result) > 0 {
		_startCursor := usecasex.Cursor(result[0].ID().String())
		_endCursor := usecasex.Cursor(result[len(result)-1].ID().String())
		startCursor = &_startCursor
		endCursor = &_endCursor
	}

	return result, usecasex.NewPageInfo(
		int64(len(result)),
		startCursor,
		endCursor,
		true,
		true,
	), nil
}

func (r *Asset) TotalSizeByWorkspace(
	_ context.Context,
	wid accountdomain.WorkspaceID,
) (t int64, err error) {
	if !r.workspaceFilter.CanRead(wid) {
		return 0, nil
	}

	r.data.Range(func(k id.AssetID, v *asset.Asset) bool {
		if v.Workspace() == wid {
			t += int64(v.Size())
		}
		return true
	})
	return
}

func (r *Asset) FindByWorkspace(
	_ context.Context,
	wid accountdomain.WorkspaceID,
	filter repo.AssetFilter,
) ([]*asset.Asset, *interfaces.PageBasedInfo, error) {
	if !r.workspaceFilter.CanRead(wid) {
		return nil, interfaces.NewPageBasedInfo(0, 1, 1), nil
	}

	result := r.data.FindAll(func(k id.AssetID, v *asset.Asset) bool {
		return v.Workspace() == wid &&
			(filter.Keyword == nil || strings.Contains(v.Name(), *filter.Keyword))
	})

	if filter.SortType != nil {
		s := *filter.SortType
		sort.SliceStable(result, func(i, j int) bool {
			if s == asset.SortTypeID {
				return result[i].ID().Compare(result[j].ID()) < 0
			}
			if s == asset.SortTypeSize {
				return result[i].Size() < result[j].Size()
			}
			if s == asset.SortTypeName {
				return strings.Compare(result[i].Name(), result[j].Name()) < 0
			}
			return false
		})
	}

	total := int64(len(result))
	if total == 0 {
		return nil, interfaces.NewPageBasedInfo(0, 1, 1), nil
	}

	if filter.Pagination != nil && filter.Pagination.Offset != nil {
		skip := int(filter.Pagination.Offset.Offset)
		limit := int(filter.Pagination.Offset.Limit)

		if skip >= len(result) {
			page := skip/limit + 1
			return nil, interfaces.NewPageBasedInfo(total, page, limit), nil
		}

		end := skip + limit
		if end > len(result) {
			end = len(result)
		}

		page := skip/limit + 1
		return result[skip:end], interfaces.NewPageBasedInfo(total, page, limit), nil
	}

	return result, interfaces.NewPageBasedInfo(total, 1, int(total)), nil
}
