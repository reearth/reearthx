package interfaces

import (
	"context"
	"github.com/reearth/reearthx/asset/domain/asset"

	"io"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/idx"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

type AssetFilterType string

type CreateAssetParam struct {
	GroupID           idx.ID[asset.GroupID]
	File              *asset.File
	Token             string
	SkipDecompression bool
}

type UpdateAssetParam struct {
	AssetID     idx.ID[asset.ID]
	PreviewType *asset.PreviewType
}

type CreateAssetUploadParam struct {
	GroupID         idx.ID[asset.GroupID]
	Filename        string
	ContentLength   int64
	ContentType     string
	ContentEncoding string

	Cursor string
}

var (
	ErrCreateAssetFailed error = rerror.NewE(i18n.T("failed to create asset"))
	ErrFileNotIncluded   error = rerror.NewE(i18n.T("file not included"))
)

type AssetFilter struct {
	Sort         *usecasex.Sort
	Keyword      *string
	Pagination   *usecasex.Pagination
	ContentTypes []string
}

type AssetUpload struct {
	URL             string
	UUID            string
	ContentType     string
	ContentLength   int64
	ContentEncoding string
	Next            string
}

type Asset interface {
	FindByID(context.Context, asset.ID, *asset.Operator) (*asset.Asset, error)
	FindByUUID(context.Context, string, *asset.Operator) (*asset.Asset, error)
	FindByIDs(context.Context, asset.IDList, *asset.Operator) (asset.List, error)
	Search(context.Context, asset.GroupID, AssetFilter, *asset.Operator) (asset.List, *usecasex.PageInfo, error)
	FindFileByID(context.Context, asset.ID, *asset.Operator) (*asset.File, error)
	FindFilesByIDs(context.Context, asset.IDList, *asset.Operator) (map[asset.ID]*asset.File, error)
	DownloadByID(context.Context, asset.ID, map[string]string, *asset.Operator) (io.ReadCloser, map[string]string, error)
	Create(context.Context, CreateAssetParam, *asset.Operator) (*asset.Asset, *asset.File, error)
	Update(context.Context, UpdateAssetParam, *asset.Operator) (*asset.Asset, error)
	UpdateFiles(context.Context, asset.ID, *asset.ExtractionStatus, *asset.Operator) (*asset.Asset, error)
	Delete(context.Context, asset.ID, *asset.Operator) (asset.ID, error)
	BatchDelete(context.Context, asset.IDList, *asset.Operator) (asset.IDList, error)
	Decompress(context.Context, asset.ID, *asset.Operator) (*asset.Asset, error)
	Publish(context.Context, asset.ID, *asset.Operator) (*asset.Asset, error)
	Unpublish(context.Context, asset.ID, *asset.Operator) (*asset.Asset, error)
	CreateUpload(context.Context, CreateAssetUploadParam, *asset.Operator) (*AssetUpload, error)
	RetryDecompression(context.Context, string) error
}
