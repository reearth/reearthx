package asset

import (
	"context"
	"io"
)

type AssetService interface {
	CreateAsset(ctx context.Context, param CreateAssetParam) (*Asset, error)
	GetAsset(ctx context.Context, id AssetID) (*Asset, error)
	GetAssetFile(ctx context.Context, id AssetID) (*AssetFile, error)
	ListAssets(ctx context.Context, groupID GroupID, filter AssetFilter, sort AssetSort, pagination Pagination) ([]*Asset, int64, error)
	UpdateAsset(ctx context.Context, param UpdateAssetParam) (*Asset, error)
	DeleteAsset(ctx context.Context, id AssetID) error
	DeleteAssets(ctx context.Context, ids []AssetID) error
	DecompressAsset(ctx context.Context, id AssetID) error
	CreateAssetUpload(ctx context.Context, param CreateAssetUploadParam) (*AssetUploadInfo, error)
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
