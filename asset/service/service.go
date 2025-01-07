package service

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/infrastructure/decompress"
	"github.com/reearth/reearthx/asset/repository"
)

type Service struct {
	repo         repository.PersistenceRepository
	decompressor repository.Decompressor
}

func NewService(repo repository.PersistenceRepository) *Service {
	return &Service{
		repo:         repo,
		decompressor: decompress.NewZipDecompressor(),
	}
}

func (s *Service) Create(ctx context.Context, asset *domain.Asset) error {
	return s.repo.Create(ctx, asset)
}

func (s *Service) Get(ctx context.Context, id domain.ID) (*domain.Asset, error) {
	return s.repo.Read(ctx, id)
}

func (s *Service) Update(ctx context.Context, asset *domain.Asset) error {
	return s.repo.Update(ctx, asset)
}

func (s *Service) Delete(ctx context.Context, id domain.ID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Upload(ctx context.Context, id domain.ID, content io.Reader) error {
	return s.repo.Upload(ctx, id, content)
}

func (s *Service) Download(ctx context.Context, id domain.ID) (io.ReadCloser, error) {
	return s.repo.Download(ctx, id)
}

func (s *Service) GetUploadURL(ctx context.Context, id domain.ID) (string, error) {
	return s.repo.GetUploadURL(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*domain.Asset, error) {
	return s.repo.List(ctx)
}

// DecompressZip decompresses zip content and returns a channel of decompressed files
func (s *Service) DecompressZip(ctx context.Context, content []byte) (<-chan repository.DecompressedFile, error) {
	return s.decompressor.DecompressWithContent(ctx, content)
}

// CompressZip compresses the provided files into a zip archive
func (s *Service) CompressZip(ctx context.Context, files map[string]io.Reader) ([]byte, error) {
	return s.decompressor.CompressWithContent(ctx, files)
}
