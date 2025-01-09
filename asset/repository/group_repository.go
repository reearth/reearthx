package repository

import (
	"context"

	"github.com/reearth/reearthx/asset/domain"
)

type GroupReader interface {
	FindByID(ctx context.Context, id domain.GroupID) (*domain.Group, error)
	FindByIDs(ctx context.Context, ids []domain.GroupID) ([]*domain.Group, error)
	List(ctx context.Context) ([]*domain.Group, error)
}

type GroupWriter interface {
	Create(ctx context.Context, group *domain.Group) error
	Update(ctx context.Context, group *domain.Group) error
	Delete(ctx context.Context, id domain.GroupID) error
}

type GroupRepository interface {
	GroupReader
	GroupWriter
}
