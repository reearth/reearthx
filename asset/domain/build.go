package domain

import (
	"errors"
	"time"
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

func (b *AssetBuilder) ID(id ID) *AssetBuilder {
	b.a.id = id
	return b
}

func (b *AssetBuilder) NewID() *AssetBuilder {
	b.a.id = NewID()
	return b
}

func (b *AssetBuilder) GroupID(groupID GroupID) *AssetBuilder {
	b.a.groupID = groupID
	return b
}

func (b *AssetBuilder) ProjectID(projectID ProjectID) *AssetBuilder {
	b.a.projectID = projectID
	return b
}

func (b *AssetBuilder) WorkspaceID(workspaceID WorkspaceID) *AssetBuilder {
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

type GroupBuilder struct {
	g *Group
}

func NewGroupBuilder() *GroupBuilder {
	return &GroupBuilder{g: &Group{}}
}

func (b *GroupBuilder) Build() (*Group, error) {
	if b.g.id.IsNil() {
		return nil, ErrInvalidID
	}
	if b.g.name == "" {
		return nil, ErrEmptyGroupName
	}
	if b.g.createdAt.IsZero() {
		now := time.Now()
		b.g.createdAt = now
		b.g.updatedAt = now
	}
	return b.g, nil
}

func (b *GroupBuilder) MustBuild() *Group {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *GroupBuilder) ID(id GroupID) *GroupBuilder {
	b.g.id = id
	return b
}

func (b *GroupBuilder) NewID() *GroupBuilder {
	b.g.id = NewGroupID()
	return b
}

func (b *GroupBuilder) Name(name string) *GroupBuilder {
	b.g.name = name
	return b
}

func (b *GroupBuilder) Policy(policy string) *GroupBuilder {
	b.g.policy = policy
	return b
}

func (b *GroupBuilder) Description(description string) *GroupBuilder {
	b.g.description = description
	return b
}

func (b *GroupBuilder) CreatedAt(createdAt time.Time) *GroupBuilder {
	b.g.createdAt = createdAt
	return b
}

func (b *GroupBuilder) UpdatedAt(updatedAt time.Time) *GroupBuilder {
	b.g.updatedAt = updatedAt
	return b
}
