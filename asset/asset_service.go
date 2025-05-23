package asset

import (
	"context"
	"io"

	"github.com/reearth/reearthx/usecasex"
)

type AssetService interface {
	CreateAsset(ctx context.Context, param CreateAssetParam) (*Asset, error)
	GetAsset(ctx context.Context, id AssetID) (*Asset, error)
	GetAssetFile(ctx context.Context, id AssetID) (*File, error)
	ListAssets(ctx context.Context, groupID GroupID, filter AssetFilter, sort AssetSort, pagination Pagination) ([]*Asset, int64, error)
	UpdateAsset(ctx context.Context, param UpdateAssetParam) (*Asset, error)
	DeleteAsset(ctx context.Context, id AssetID) error
	DeleteAssets(ctx context.Context, ids []AssetID) error
	DecompressAsset(ctx context.Context, id AssetID) error
	CreateAssetUpload(ctx context.Context, param CreateAssetUploadParam) (*AssetUploadInfo, error)

	Filtered(ProjectFilter) AssetService
	FindByProject(context.Context, GroupID, AssetFilter) (List, *usecasex.PageInfo, error)
	FindByID(context.Context, AssetID) (*Asset, error)
	FindByUUID(context.Context, string) (*Asset, error)
	FindByIDs(context.Context, AssetIDList) (List, error)
	Save(context.Context, *Asset) error
	Delete(context.Context, AssetID) error
	BatchDelete(context.Context, AssetIDList) error
}

type AssetFile interface {
	FindByID(context.Context, AssetID) (*File, error)
	FindByIDs(context.Context, AssetIDList) (map[AssetID]*File, error)
	Save(context.Context, AssetID, *File) error
	SaveFlat(context.Context, AssetID, *File, []*File) error
}

type CreateAssetParam struct {
	GroupID           GroupID
	File              io.Reader
	FileName          string
	Size              int64
	URL               string
	Token             string
	ContentType       string
	ContentEncoding   string
	SkipDecompression bool
}

type UpdateAssetParam struct {
	ID          AssetID
	PreviewType *PreviewType
}

type CreateAssetUploadParam struct {
	GroupID         GroupID
	FileName        string
	ContentLength   int64
	ContentEncoding string
	Cursor          string
}

type AssetUploadInfo struct {
	Token           string
	URL             string
	ContentType     string
	ContentLength   int64
	ContentEncoding string
	Next            string
}

type PageInfo struct {
	TotalCount int64
	HasNext    bool
}

type AssetIDList []AssetID

func (l AssetIDList) Add(id AssetID) AssetIDList {
	return append(l, id)
}

type AssetUpload struct {
	Token           string
	URL             string
	ContentType     string
	ContentLength   int64
	ContentEncoding string
	Next            string
}
