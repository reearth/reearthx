package repository

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

type Asset interface {
	Save(ctx context.Context, asset *entity.Asset) error
	FindByID(ctx context.Context, id id.ID) (*entity.Asset, error)
	FindByIDs(ctx context.Context, ids []id.ID) ([]*entity.Asset, error)
	FindByWorkspace(ctx context.Context, workspaceID id.WorkspaceID) ([]*entity.Asset, error)
	FindByProject(ctx context.Context, projectID id.ProjectID) ([]*entity.Asset, error)
	FindByGroup(ctx context.Context, groupID id.GroupID) ([]*entity.Asset, error)
	Remove(ctx context.Context, id id.ID) error
}

type Group interface {
	Save(ctx context.Context, group *entity.Group) error
	FindByID(ctx context.Context, id id.GroupID) (*entity.Group, error)
	FindByIDs(ctx context.Context, ids []id.GroupID) ([]*entity.Group, error)
	Remove(ctx context.Context, id id.GroupID) error
}
