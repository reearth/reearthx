package repository

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain"
)

type Reader interface {
	Read(ctx context.Context, id domain.ID) (*domain.Asset, error)
	List(ctx context.Context) ([]*domain.Asset, error)
}

type Writer interface {
	Create(ctx context.Context, asset *domain.Asset) error
	Update(ctx context.Context, asset *domain.Asset) error
	Delete(ctx context.Context, id domain.ID) error
}

type FileOperator interface {
	Upload(ctx context.Context, id domain.ID, content io.Reader) error
	Download(ctx context.Context, id domain.ID) (io.ReadCloser, error)
	GetUploadURL(ctx context.Context, id domain.ID) (string, error)
}

type Repository interface {
	Reader
	Writer
	FileOperator
}
