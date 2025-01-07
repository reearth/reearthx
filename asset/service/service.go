package service

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/infrastructure/decompress"
	"github.com/reearth/reearthx/asset/repository"
	"github.com/reearth/reearthx/log"
)

// Service handles asset operations including CRUD, upload/download, and compression
type Service struct {
	repo         repository.PersistenceRepository
	decompressor repository.Decompressor
	pubsub       repository.PubSubRepository
}

// NewService creates a new Service instance with the given persistence repository
func NewService(repo repository.PersistenceRepository, pubsub repository.PubSubRepository) *Service {
	return &Service{
		repo:         repo,
		decompressor: decompress.NewZipDecompressor(),
		pubsub:       pubsub,
	}
}

// Create creates a new asset
func (s *Service) Create(ctx context.Context, asset *domain.Asset) error {
	if err := s.repo.Create(ctx, asset); err != nil {
		return err
	}

	if err := s.pubsub.PublishAssetCreated(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish asset created event: %v", err)
	}

	return nil
}

// Get retrieves an asset by ID
func (s *Service) Get(ctx context.Context, id domain.ID) (*domain.Asset, error) {
	return s.repo.Read(ctx, id)
}

// Update updates an existing asset
func (s *Service) Update(ctx context.Context, asset *domain.Asset) error {
	if err := s.repo.Update(ctx, asset); err != nil {
		return err
	}

	if err := s.pubsub.PublishAssetUpdated(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish asset updated event: %v", err)
	}

	return nil
}

// Delete removes an asset by ID
func (s *Service) Delete(ctx context.Context, id domain.ID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.pubsub.PublishAssetDeleted(ctx, id); err != nil {
		log.Errorfc(ctx, "failed to publish asset deleted event: %v", err)
	}

	return nil
}

// Upload uploads content for an asset with the given ID
func (s *Service) Upload(ctx context.Context, id domain.ID, content io.Reader) error {
	if err := s.repo.Upload(ctx, id, content); err != nil {
		return err
	}

	asset, err := s.repo.Read(ctx, id)
	if err != nil {
		return err
	}

	if err := s.pubsub.PublishAssetUploaded(ctx, asset); err != nil {
		log.Errorfc(ctx, "failed to publish asset uploaded event: %v", err)
	}

	return nil
}

// Download retrieves the content of an asset by ID
func (s *Service) Download(ctx context.Context, id domain.ID) (io.ReadCloser, error) {
	return s.repo.Download(ctx, id)
}

// GetUploadURL generates a URL for uploading content to an asset
func (s *Service) GetUploadURL(ctx context.Context, id domain.ID) (string, error) {
	return s.repo.GetUploadURL(ctx, id)
}

// List returns all assets
func (s *Service) List(ctx context.Context) ([]*domain.Asset, error) {
	return s.repo.List(ctx)
}

// DecompressZip decompresses zip content and returns a channel of decompressed files.
// The channel will be closed when all files have been processed or an error occurs.
func (s *Service) DecompressZip(ctx context.Context, content []byte) (<-chan repository.DecompressedFile, error) {
	ch, err := s.decompressor.DecompressWithContent(ctx, content)
	if err != nil {
		return nil, err
	}

	// Get asset ID from context if available
	if assetID, ok := ctx.Value("assetID").(domain.ID); ok {
		asset, err := s.repo.Read(ctx, assetID)
		if err != nil {
			return nil, err
		}

		asset.UpdateStatus(domain.StatusExtracting, "")
		if err := s.repo.Update(ctx, asset); err != nil {
			return nil, err
		}

		if err := s.pubsub.PublishAssetExtracted(ctx, asset); err != nil {
			log.Errorfc(ctx, "failed to publish asset extracted event: %v", err)
		}
	}

	return ch, nil
}

// CompressZip compresses the provided files into a zip archive.
// Returns a channel that will receive the compressed bytes or an error.
// The channel will be closed when compression is complete or if an error occurs.
func (s *Service) CompressZip(ctx context.Context, files map[string]io.Reader) (<-chan repository.CompressResult, error) {
	return s.decompressor.CompressWithContent(ctx, files)
}
