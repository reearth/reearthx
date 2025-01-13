package builder

import (
	"time"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

type AssetBuilder struct {
	a *entity.Asset
}

func NewAssetBuilder() *AssetBuilder {
	return &AssetBuilder{a: &entity.Asset{}}
}

func (b *AssetBuilder) Build() (*entity.Asset, error) {
	if b.a.ID() == (id.ID{}) {
		return nil, id.ErrInvalidID
	}
	if b.a.WorkspaceID() == (id.WorkspaceID{}) {
		return nil, domain.ErrEmptyWorkspaceID
	}
	if b.a.URL() == "" {
		return nil, domain.ErrEmptyURL
	}
	if b.a.Size() <= 0 {
		return nil, domain.ErrEmptySize
	}
	if b.a.CreatedAt().IsZero() {
		now := time.Now()
		b = b.CreatedAt(now).UpdatedAt(now)
	}
	if b.a.Status() == "" {
		b = b.Status(entity.StatusPending)
	}
	return b.a, nil
}

func (b *AssetBuilder) MustBuild() *entity.Asset {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *AssetBuilder) ID(id id.ID) *AssetBuilder {
	b.a = entity.NewAsset(id, b.a.Name(), b.a.Size(), b.a.ContentType())
	return b
}

func (b *AssetBuilder) NewID() *AssetBuilder {
	return b.ID(id.NewID())
}

func (b *AssetBuilder) GroupID(groupID id.GroupID) *AssetBuilder {
	b.a.MoveToGroup(groupID)
	return b
}

func (b *AssetBuilder) ProjectID(projectID id.ProjectID) *AssetBuilder {
	b.a.MoveToProject(projectID)
	return b
}

func (b *AssetBuilder) WorkspaceID(workspaceID id.WorkspaceID) *AssetBuilder {
	b.a.MoveToWorkspace(workspaceID)
	return b
}

func (b *AssetBuilder) Name(name string) *AssetBuilder {
	b.a.UpdateMetadata(name, b.a.URL(), b.a.ContentType())
	return b
}

func (b *AssetBuilder) Size(size int64) *AssetBuilder {
	b.a.SetSize(size)
	return b
}

func (b *AssetBuilder) URL(url string) *AssetBuilder {
	b.a.UpdateMetadata(b.a.Name(), url, b.a.ContentType())
	return b
}

func (b *AssetBuilder) ContentType(contentType string) *AssetBuilder {
	b.a.UpdateMetadata(b.a.Name(), b.a.URL(), contentType)
	return b
}

func (b *AssetBuilder) Status(status entity.Status) *AssetBuilder {
	b.a.UpdateStatus(status, b.a.Error())
	return b
}

func (b *AssetBuilder) Error(err string) *AssetBuilder {
	b.a.UpdateStatus(b.a.Status(), err)
	return b
}

// CreatedAt sets the creation time of the asset
func (b *AssetBuilder) CreatedAt(createdAt time.Time) *AssetBuilder {
	// We need to create a new asset to set createdAt
	b.a = entity.NewAsset(b.a.ID(), b.a.Name(), b.a.Size(), b.a.ContentType())
	// Restore other fields
	if b.a.GroupID() != (id.GroupID{}) {
		b.GroupID(b.a.GroupID())
	}
	if b.a.ProjectID() != (id.ProjectID{}) {
		b.ProjectID(b.a.ProjectID())
	}
	if b.a.WorkspaceID() != (id.WorkspaceID{}) {
		b.WorkspaceID(b.a.WorkspaceID())
	}
	if b.a.URL() != "" {
		b.URL(b.a.URL())
	}
	if b.a.Status() != "" {
		b.Status(b.a.Status())
	}
	if b.a.Error() != "" {
		b.Error(b.a.Error())
	}
	return b
}

// UpdatedAt is not needed as it's handled internally by the entity
func (b *AssetBuilder) UpdatedAt(updatedAt time.Time) *AssetBuilder {
	return b
}
