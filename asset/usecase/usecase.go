package assetusecase

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
)

type Usecase interface {
	// CreateAsset creates a new asset
	CreateAsset(ctx context.Context, asset *domain.Asset) error
	// GetAsset retrieves an asset by ID
	GetAsset(ctx context.Context, id domain.ID) (*domain.Asset, error)
	// UpdateAsset updates an existing asset
	UpdateAsset(ctx context.Context, asset *domain.Asset) error
	// DeleteAsset removes an asset by ID
	DeleteAsset(ctx context.Context, id domain.ID) error
	// UploadAssetContent uploads content for an asset with the given ID
	UploadAssetContent(ctx context.Context, id domain.ID, content io.Reader) error
	// DownloadAssetContent retrieves the content of an asset by ID
	DownloadAssetContent(ctx context.Context, id domain.ID) (io.ReadCloser, error)
	// GetAssetUploadURL generates a URL for uploading content to an asset
	GetAssetUploadURL(ctx context.Context, id domain.ID) (string, error)
	// ListAssets returns all assets
	ListAssets(ctx context.Context) ([]*domain.Asset, error)
	// DecompressZipContent decompresses zip content and returns a channel of decompressed files
	DecompressZipContent(ctx context.Context, content []byte) (<-chan repository.DecompressedFile, error)
	// CompressToZip compresses the provided files into a zip archive
	CompressToZip(ctx context.Context, files map[string]io.Reader) (<-chan repository.CompressResult, error)
	// DeleteAllAssetsInGroup deletes all assets in a group
	DeleteAllAssetsInGroup(ctx context.Context, groupID domain.GroupID) error
}
