package domain

import (
	"errors"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
)

var (
	ErrEmptyWorkspaceID = errors.New("workspace id is required")
	ErrEmptyURL         = errors.New("url is required")
	ErrEmptySize        = errors.New("size must be greater than 0")
)

type AssetBuilder struct {
	a *Asset
}

func NewAssetBuilder() *AssetBuilder {
	return &AssetBuilder{a: &Asset{}}
}

func (b *AssetBuilder) Build() (*Asset, error) {
	if b.a.id.IsNil() {
		return nil, ErrInvalidID
	}
	if b.a.workspaceID.IsNil() {
		return nil, ErrEmptyWorkspaceID
	}
	if b.a.url == "" {
		return nil, ErrEmptyURL
	}
	if b.a.size <= 0 {
		return nil, ErrEmptySize
	}
	if b.a.createdAt.IsZero() {
		now := time.Now()
		b.a.createdAt = now
		b.a.updatedAt = now
	}
	if b.a.status == "" {
		b.a.status = StatusPending
	}
	return b.a, nil
}

func (b *AssetBuilder) MustBuild() *Asset {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *AssetBuilder) ID(id id.ID) *AssetBuilder {
	b.a.id = id
	return b
}

func (b *AssetBuilder) NewID() *AssetBuilder {
	b.a.id = id.NewID()
	return b
}

func (b *AssetBuilder) GroupID(groupID id.GroupID) *AssetBuilder {
	b.a.groupID = groupID
	return b
}

func (b *AssetBuilder) ProjectID(projectID id.ProjectID) *AssetBuilder {
	b.a.projectID = projectID
	return b
}

func (b *AssetBuilder) WorkspaceID(workspaceID id.WorkspaceID) *AssetBuilder {
	b.a.workspaceID = workspaceID
	return b
}

func (b *AssetBuilder) Name(name string) *AssetBuilder {
	b.a.name = name
	return b
}

func (b *AssetBuilder) Size(size int64) *AssetBuilder {
	b.a.size = size
	return b
}

func (b *AssetBuilder) URL(url string) *AssetBuilder {
	b.a.url = url
	return b
}

func (b *AssetBuilder) ContentType(contentType string) *AssetBuilder {
	b.a.contentType = contentType
	return b
}

func (b *AssetBuilder) Status(status Status) *AssetBuilder {
	b.a.status = status
	return b
}

func (b *AssetBuilder) Error(err string) *AssetBuilder {
	b.a.error = err
	return b
}

func (b *AssetBuilder) CreatedAt(createdAt time.Time) *AssetBuilder {
	b.a.createdAt = createdAt
	return b
}

func (b *AssetBuilder) UpdatedAt(updatedAt time.Time) *AssetBuilder {
	b.a.updatedAt = updatedAt
	return b
}
