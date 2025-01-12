package domain

import (
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
)

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

func (b *GroupBuilder) ID(id id.GroupID) *GroupBuilder {
	b.g.id = id
	return b
}

func (b *GroupBuilder) NewID() *GroupBuilder {
	b.g.id = id.NewGroupID()
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
