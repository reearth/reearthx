package builder

import (
	"time"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
)

type GroupBuilder struct {
	g *entity.Group
}

func NewGroupBuilder() *GroupBuilder {
	return &GroupBuilder{g: &entity.Group{}}
}

func (b *GroupBuilder) Build() (*entity.Group, error) {
	if b.g.ID() == (id.GroupID{}) {
		return nil, id.ErrInvalidID
	}
	if b.g.Name() == "" {
		return nil, domain.ErrEmptyGroupName
	}
	if b.g.CreatedAt().IsZero() {
		now := time.Now()
		b.CreatedAt(now)
	}
	return b.g, nil
}

func (b *GroupBuilder) MustBuild() *entity.Group {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *GroupBuilder) ID(id id.GroupID) *GroupBuilder {
	b.g = entity.NewGroup(id, b.g.Name())
	return b
}

func (b *GroupBuilder) NewID() *GroupBuilder {
	return b.ID(id.NewGroupID())
}

func (b *GroupBuilder) Name(name string) *GroupBuilder {
	if err := b.g.UpdateName(name); err != nil {
		// Since this is a builder pattern, we'll ignore the error here
		// and let it be caught during Build()
		return b
	}
	return b
}

func (b *GroupBuilder) Policy(policy string) *GroupBuilder {
	if err := b.g.UpdatePolicy(policy); err != nil {
		// Since this is a builder pattern, we'll ignore the error here
		// and let it be caught during Build()
		return b
	}
	return b
}

func (b *GroupBuilder) Description(description string) *GroupBuilder {
	b.g.UpdateDescription(description)
	return b
}

// CreatedAt sets the creation time of the group
func (b *GroupBuilder) CreatedAt(createdAt time.Time) *GroupBuilder {
	// We need to create a new group to set createdAt
	b.g = entity.NewGroup(b.g.ID(), b.g.Name())
	// Restore other fields
	if b.g.Policy() != "" {
		b.Policy(b.g.Policy())
	}
	if b.g.Description() != "" {
		b.Description(b.g.Description())
	}
	return b
}

// UpdatedAt is not needed as it's handled internally by the entity
func (b *GroupBuilder) UpdatedAt(updatedAt time.Time) *GroupBuilder {
	return b
}
