package assetmemory

import (
	"context"

	id "github.com/reearth/reearthx/asset/assetdomain"
	"github.com/reearth/reearthx/asset/assetdomain/asset"
	"github.com/reearth/reearthx/asset/assetusecase/assetrepo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"
	"golang.org/x/exp/slices"
)

var _ assetrepo.AssetFile = (*AssetFileImpl)(nil)

type AssetFileImpl struct {
	data  *util.SyncMap[asset.ID, *asset.File]
	files *util.SyncMap[asset.ID, []*asset.File]
	err   error
}

func NewAssetFile() *AssetFileImpl {
	return &AssetFileImpl{
		data:  &util.SyncMap[id.AssetID, *asset.File]{},
		files: &util.SyncMap[id.AssetID, []*asset.File]{},
	}
}

func (r *AssetFileImpl) FindByID(ctx context.Context, id id.AssetID) (*asset.File, error) {
	if r.err != nil {
		return nil, r.err
	}

	f := r.data.Find(func(key asset.ID, value *asset.File) bool {
		return key == id
	}).Clone()
	fs := r.files.Find(func(key asset.ID, value []*asset.File) bool {
		return key == id
	})
	if len(fs) > 0 {
		// f = asset.FoldFiles(fs, f)
		f.SetFiles(fs)
	}
	return rerror.ErrIfNil(f, rerror.ErrNotFound)
}

func (r *AssetFileImpl) Save(ctx context.Context, id id.AssetID, file *asset.File) error {
	if r.err != nil {
		return r.err
	}

	r.data.Store(id, file.Clone())
	return nil
}

func (r *AssetFileImpl) SaveFlat(ctx context.Context, id id.AssetID, parent *asset.File, files []*asset.File) error {
	if r.err != nil {
		return r.err
	}
	r.data.Store(id, parent.Clone())
	r.files.Store(id, slices.Clone(files))
	return nil
}
