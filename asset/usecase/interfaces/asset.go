package interfaces

import (
	"archive/zip"
	"context"
	"io"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/project"

	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/file"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/usecase"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/idx"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

type AssetFilterType string

type CreateAssetParam struct {
	File              *file.File
	Token             string
	WorkspaceID       accountdomain.WorkspaceID // viz
	ProjectID         idx.ID[id.Project]
	SkipDecompression bool
}

type UpdateAssetParam struct {
	PreviewType *asset.PreviewType
	AssetID     idx.ID[id.Asset]
}

type CreateAssetUploadParam struct {
	Filename        string
	ContentType     string
	ContentEncoding string

	Cursor        string
	ContentLength int64
	ProjectID     idx.ID[id.Project]
}

var (
	ErrCreateAssetFailed error = rerror.NewE(i18n.T("failed to create asset"))
	ErrFileNotIncluded   error = rerror.NewE(i18n.T("file not included"))
)

type AssetFilter struct {
	Sort         *usecasex.Sort
	VizSort      *asset.SortType // viz sort
	Keyword      *string
	Pagination   *usecasex.Pagination
	ContentTypes []string
}

type AssetUpload struct {
	URL             string
	UUID            string
	ContentType     string
	ContentEncoding string
	Next            string
	ContentLength   int64
}

type PageBasedPaginationParam struct {
	OrderBy  *string
	OrderDir *string
	Page     int
	PageSize int
}

type PaginationParam struct {
	Page *PageBasedPaginationParam
}

type Asset interface {
	FindByID(context.Context, id.AssetID, *usecase.Operator) (*asset.Asset, error)
	FindByUUID(context.Context, string, *usecase.Operator) (*asset.Asset, error)
	FindByIDs(context.Context, []id.AssetID, *usecase.Operator) (asset.List, error)
	Search(
		context.Context,
		id.ProjectID,
		AssetFilter,
		*usecase.Operator,
	) (asset.List, *usecasex.PageInfo, error)
	FindFileByID(context.Context, id.AssetID, *usecase.Operator) (*asset.File, error)
	FindFilesByIDs(
		context.Context,
		id.AssetIDList,
		*usecase.Operator,
	) (map[id.AssetID]*asset.File, error)
	DownloadByID(
		context.Context,
		id.AssetID,
		map[string]string,
		*usecase.Operator,
	) (io.ReadCloser, map[string]string, error)
	Create(context.Context, CreateAssetParam, *usecase.Operator) (*asset.Asset, *asset.File, error)
	CreateWithWorkspace(context.Context, CreateAssetParam, *usecase.Operator) (*asset.Asset, *asset.File, error)
	Update(context.Context, UpdateAssetParam, *usecase.Operator) (*asset.Asset, error)
	UpdateFiles(
		context.Context,
		id.AssetID,
		*asset.ArchiveExtractionStatus,
		*usecase.Operator,
	) (*asset.Asset, error)
	Delete(context.Context, id.AssetID, *usecase.Operator) (id.AssetID, error)
	BatchDelete(context.Context, id.AssetIDList, *usecase.Operator) ([]id.AssetID, error)
	Decompress(context.Context, id.AssetID, *usecase.Operator) (*asset.Asset, error)
	Publish(context.Context, id.AssetID, *usecase.Operator) (*asset.Asset, error)
	Unpublish(context.Context, id.AssetID, *usecase.Operator) (*asset.Asset, error)
	CreateUpload(context.Context, CreateAssetUploadParam, *usecase.Operator) (*AssetUpload, error)
	RetryDecompression(context.Context, string) error

	FindByWorkspaceProject(
		context.Context,
		accountdomain.WorkspaceID,
		*id.ProjectID,
		*string,
		*asset.SortType,
		*usecasex.Pagination,
		*usecase.Operator,
	) ([]*asset.Asset, *usecasex.PageInfo, error)
	ImportAssetFiles(
		context.Context,
		map[string]*zip.File,
		*[]byte,
		*project.Project,
	) (*[]byte, error) // viz

	FindByWorkspace(
		context.Context,
		accountdomain.WorkspaceID,
		*string,
		*asset.SortType,
		*PaginationParam,
	) ([]*asset.Asset, *PageBasedInfo, error)
	Fetch(context.Context, []id.AssetID) ([]*asset.Asset, error)
}
