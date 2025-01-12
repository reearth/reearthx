package assetinteractor

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/infrastructure/decompress"
	"github.com/reearth/reearthx/asset/repository"
	"github.com/reearth/reearthx/log"
)

type AssetInteractor struct {
	repo         repository.PersistenceRepository
	decompressor repository.Decompressor
	pubsub       repository.PubSubRepository
}

func NewAssetInteractor(repo repository.PersistenceRepository, pubsub repository.PubSubRepository) *AssetInteractor {
	return &AssetInteractor{
		repo:         repo,
		decompressor: decompress.NewZipDecompressor(),
		pubsub:       pubsub,
	}
}

// CreateAsset creates a new asset
func (i *AssetInteractor) CreateAsset(ctx context.Context, asset *domain.Asset) error {
	if err := i.repo.Create(ctx, asset); err != nil {
		return err
	}

	if err := i.pubsub.PublishAssetCreated(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish asset created event: %v", err)
	}

	return nil
}

// GetAsset retrieves an asset by ID
func (i *AssetInteractor) GetAsset(ctx context.Context, id domain.ID) (*domain.Asset, error) {
	return i.repo.Read(ctx, id)
}

// UpdateAsset updates an existing asset
func (i *AssetInteractor) UpdateAsset(ctx context.Context, asset *domain.Asset) error {
	if err := i.repo.Update(ctx, asset); err != nil {
		return err
	}

	if err := i.pubsub.PublishAssetUpdated(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish asset updated event: %v", err)
	}

	return nil
}

// DeleteAsset removes an asset by ID
func (i *AssetInteractor) DeleteAsset(ctx context.Context, id domain.ID) error {
	if err := i.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := i.pubsub.PublishAssetDeleted(ctx, id); err != nil {
		log.Errorfc(ctx, "failed to publish asset deleted event: %v", err)
	}

	return nil
}

// UploadAssetContent uploads content for an asset with the given ID
func (i *AssetInteractor) UploadAssetContent(ctx context.Context, id domain.ID, content io.Reader) error {
	if err := i.repo.Upload(ctx, id, content); err != nil {
		return err
	}

	asset, err := i.repo.Read(ctx, id)
	if err != nil {
		return err
	}

	if err := i.pubsub.PublishAssetUploaded(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish asset uploaded event: %v", err)
	}

	return nil
}

// DownloadAssetContent retrieves the content of an asset by ID
func (i *AssetInteractor) DownloadAssetContent(ctx context.Context, id domain.ID) (io.ReadCloser, error) {
	return i.repo.Download(ctx, id)
}

// GetAssetUploadURL generates a URL for uploading content to an asset
func (i *AssetInteractor) GetAssetUploadURL(ctx context.Context, id domain.ID) (string, error) {
	return i.repo.GetUploadURL(ctx, id)
}

// ListAssets returns all assets
func (i *AssetInteractor) ListAssets(ctx context.Context) ([]*domain.Asset, error) {
	return i.repo.List(ctx)
}

// DecompressZipContent decompresses zip content and returns a channel of decompressed files
func (i *AssetInteractor) DecompressZipContent(ctx context.Context, content []byte) (<-chan repository.DecompressedFile, error) {
	ch, err := i.decompressor.DecompressWithContent(ctx, content)
	if err != nil {
		return nil, err
	}

	// Get asset ID from context if available
	if assetID, ok := ctx.Value("assetID").(domain.ID); ok {
		asset, err := i.repo.Read(ctx, assetID)
		if err != nil {
			return nil, err
		}

		asset.UpdateStatus(domain.StatusExtracting, "")
		if err := i.repo.Update(ctx, asset); err != nil {
			return nil, err
		}

		if err := i.pubsub.PublishAssetExtracted(ctx, asset); err != nil {
			log.Errorfc(ctx, "failed to publish asset extracted event: %v", err)
		}
	}

	return ch, nil
}

// CompressToZip compresses the provided files into a zip archive
func (i *AssetInteractor) CompressToZip(ctx context.Context, files map[string]io.Reader) (<-chan repository.CompressResult, error) {
	return i.decompressor.CompressWithContent(ctx, files)
}

// DeleteAllAssetsInGroup deletes all assets in a group
func (i *AssetInteractor) DeleteAllAssetsInGroup(ctx context.Context, groupID domain.GroupID) error {
	// Get all assets in the group
	assets, err := i.repo.FindByGroup(ctx, groupID)
	if err != nil {
		return err
	}

	// Delete each asset
	for _, asset := range assets {
		if err := i.DeleteAsset(ctx, asset.ID()); err != nil {
			log.Errorfc(ctx, "failed to delete asset %s in group %s: %v", asset.ID(), groupID, err)
			return err
		}
	}

	return nil
}
