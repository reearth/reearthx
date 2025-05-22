package asset

import (
	"context"
	"io"
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

	// CMS
	FindByID(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error)
	FindByUUID(ctx context.Context, uuid string, operator *AssetOperator) (*Asset, error)
	FindByIDs(ctx context.Context, ids []AssetID, operator *AssetOperator) ([]*Asset, error)
	FindByProject(ctx context.Context, groupID GroupID, filter AssetFilter, operator *AssetOperator) ([]*Asset, *PageInfo, error)
	FindFileByID(ctx context.Context, id AssetID, operator *AssetOperator) (*File, error)
	FindFilesByIDs(ctx context.Context, ids AssetIDList, operator *AssetOperator) (map[AssetID]*File, error)
	DownloadByID(ctx context.Context, id AssetID, headers map[string]string, operator *AssetOperator) (io.ReadCloser, map[string]string, error)
	Create(ctx context.Context, param CreateAssetParam, operator *AssetOperator) (*Asset, *File, error)
	Update(ctx context.Context, param UpdateAssetParam, operator *AssetOperator) (*Asset, error)
	UpdateFiles(ctx context.Context, id AssetID, status *ExtractionStatus, operator *AssetOperator) (*Asset, error)
	Delete(ctx context.Context, id AssetID, operator *AssetOperator) (AssetID, error)
	BatchDelete(ctx context.Context, ids AssetIDList, operator *AssetOperator) ([]AssetID, error)
	Decompress(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error)
	Publish(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error)
	Unpublish(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error)
	CreateUpload(ctx context.Context, param CreateAssetUploadParam, operator *AssetOperator) (*AssetUpload, error)
	RetryDecompression(ctx context.Context, id string) error
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

type AssetUpload struct {
	Token           string
	URL             string
	ContentType     string
	ContentLength   int64
	ContentEncoding string
	Next            string
}
