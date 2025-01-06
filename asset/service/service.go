package service

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
)

type Service struct {
	repo repository.PersistenceRepository
}

func NewService(repo repository.PersistenceRepository) *Service {
	return &Service{repo: repo}
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
