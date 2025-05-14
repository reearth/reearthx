package repository

import (
	"context"
	"io"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

type Reader interface {
	Read(ctx context.Context, id id.ID) (*entity.Asset, error)
	List(ctx context.Context) ([]*entity.Asset, error)
	FindByGroup(ctx context.Context, groupID id.GroupID) ([]*entity.Asset, error)
}

type Writer interface {
	Create(ctx context.Context, asset *entity.Asset) error
	Update(ctx context.Context, asset *entity.Asset) error
	Delete(ctx context.Context, id id.ID) error
}

type FileOperator interface {
	Upload(ctx context.Context, id id.ID, content io.Reader) error
	Download(ctx context.Context, id id.ID) (io.ReadCloser, error)
	GetUploadURL(ctx context.Context, id id.ID) (string, error)
}

type PersistenceRepository interface {
	Reader
	Writer
	FileOperator
}
