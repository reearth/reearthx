package asset

import (
	"context"
	"io"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateAssetInput) (*Asset, error) {
	asset := &Asset{
		ID:          ID(generateID()),
		Name:        input.Name,
		Size:        input.Size,
		ContentType: input.ContentType,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *Service) Update(ctx context.Context, id ID, input UpdateAssetInput) (*Asset, error) {
	asset, err := s.repo.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		asset.Name = *input.Name
	}
	if input.URL != nil {
		asset.URL = *input.URL
	}
	if input.ContentType != nil {
		asset.ContentType = *input.ContentType
	}
	asset.Status = input.Status
	asset.Error = input.Error
	asset.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *Service) Delete(ctx context.Context, id ID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Get(ctx context.Context, id ID) (*Asset, error) {
	return s.repo.Read(ctx, id)
}

func (s *Service) GetFile(ctx context.Context, id ID) (io.ReadCloser, error) {
	return s.repo.FetchFile(ctx, id)
}

func (s *Service) Upload(ctx context.Context, id ID, file io.Reader) error {
	return s.repo.Upload(ctx, id, file)
}

func (s *Service) GetUploadURL(ctx context.Context, id ID) (string, error) {
	return s.repo.GetUploadURL(ctx, id)
}
