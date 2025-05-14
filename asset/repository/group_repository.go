package repository

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

type GroupReader interface {
	FindByID(ctx context.Context, id id.GroupID) (*entity.Group, error)
	FindByIDs(ctx context.Context, ids []id.GroupID) ([]*entity.Group, error)
	List(ctx context.Context) ([]*entity.Group, error)
}

type GroupWriter interface {
	Create(ctx context.Context, group *entity.Group) error
	Update(ctx context.Context, group *entity.Group) error
	Delete(ctx context.Context, id id.GroupID) error
}

type GroupRepository interface {
	GroupReader
	GroupWriter
}
