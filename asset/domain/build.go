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

type Builder struct {
	a *Asset
}

func New() *Builder {
	return &Builder{a: &Asset{}}
}

func (b *Builder) Build() (*Asset, error) {
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

func (b *Builder) MustBuild() *Asset {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *Builder) ID(id ID) *Builder {
	b.a.id = id
	return b
}

func (b *Builder) NewID() *Builder {
	b.a.id = NewID()
	return b
}

func (b *Builder) GroupID(groupID GroupID) *Builder {
	b.a.groupID = groupID
	return b
}

func (b *Builder) ProjectID(projectID ProjectID) *Builder {
	b.a.projectID = projectID
	return b
}

func (b *Builder) WorkspaceID(workspaceID WorkspaceID) *Builder {
	b.a.workspaceID = workspaceID
	return b
}

func (b *Builder) Name(name string) *Builder {
	b.a.name = name
	return b
}

func (b *Builder) Size(size int64) *Builder {
	b.a.size = size
	return b
}

func (b *Builder) URL(url string) *Builder {
	b.a.url = url
	return b
}

func (b *Builder) ContentType(contentType string) *Builder {
	b.a.contentType = contentType
	return b
}

func (b *Builder) Status(status Status) *Builder {
	b.a.status = status
	return b
}

func (b *Builder) Error(err string) *Builder {
	b.a.error = err
	return b
}

func (b *Builder) CreatedAt(createdAt time.Time) *Builder {
	b.a.createdAt = createdAt
	return b
}

func (b *Builder) UpdatedAt(updatedAt time.Time) *Builder {
	b.a.updatedAt = updatedAt
	return b
}
