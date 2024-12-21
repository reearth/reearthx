package assetmemory

import (
	"context"

	id "github.com/reearth/reearthx/asset/assetdomain"
	"github.com/reearth/reearthx/asset/assetdomain/asset"
	repo "github.com/reearth/reearthx/asset/assetusecase/assetrepo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
)

var _ repo.Asset = (*AssetImpl)(nil)

type AssetImpl struct {
	data *util.SyncMap[asset.ID, *asset.Asset]
	err  error
	f    repo.ProjectFilter
}

func NewAsset() repo.Asset {
	return &AssetImpl{
		data: &util.SyncMap[id.AssetID, *asset.Asset]{},
	}
}

func (r *AssetImpl) Filtered(f repo.ProjectFilter) repo.Asset {
	return &AssetImpl{
		data: r.data,
		f:    r.f.Merge(f),
	}
}

func (r *AssetImpl) FindByID(ctx context.Context, id id.AssetID) (*asset.Asset, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.Find(func(key asset.ID, value *asset.Asset) bool {
		return key == id && r.f.CanRead(value.Project())
	}), rerror.ErrNotFound)
}

func (r *AssetImpl) FindByIDs(ctx context.Context, ids id.AssetIDList) ([]*asset.Asset, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := asset.List(r.data.FindAll(func(key asset.ID, value *asset.Asset) bool {
		return ids.Has(key) && r.f.CanRead(value.Project())
	})).SortByID()
	return res, nil
}

func (r *AssetImpl) FindByProject(ctx context.Context, id id.ProjectID, filter repo.AssetFilter) ([]*asset.Asset, *usecasex.PageInfo, error) {
	if !r.f.CanRead(id) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	if r.err != nil {
		return nil, nil, r.err
	}

	result := asset.List(r.data.FindAll(func(_ asset.ID, v *asset.Asset) bool {
		return v.Project() == id
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

func (r *AssetImpl) Save(ctx context.Context, a *asset.Asset) error {
	if !r.f.CanWrite(a.Project()) {
		return repo.ErrOperationDenied
	}

	if r.err != nil {
		return r.err
	}

	r.data.Store(a.ID(), a)
	return nil
}

func (r *AssetImpl) Delete(ctx context.Context, id id.AssetID) error {
	if r.err != nil {
		return r.err
	}

	if a, ok := r.data.Load(id); ok && r.f.CanWrite(a.Project()) {
		r.data.Delete(id)
	}
	return nil
}
