package entity

import (
	"time"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type Group struct {
	id          id.GroupID
	name        string
	policy      string
	description string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewGroup(id id.GroupID, name string) *Group {
	now := time.Now()
	return &Group{
		id:        id,
		name:      name,
		createdAt: now,
		updatedAt: now,
	}
}

// Getters
func (g *Group) ID() id.GroupID       { return g.id }
func (g *Group) Name() string         { return g.name }
func (g *Group) Policy() string       { return g.policy }
func (g *Group) Description() string  { return g.description }
func (g *Group) CreatedAt() time.Time { return g.createdAt }
func (g *Group) UpdatedAt() time.Time { return g.updatedAt }

// Setters
func (g *Group) UpdateName(name string) error {
	if name == "" {
		return domain.ErrEmptyGroupName
	}
	g.name = name
	g.updatedAt = time.Now()
	return nil
}

func (g *Group) UpdatePolicy(policy string) error {
	if policy == "" {
		return domain.ErrEmptyPolicy
	}
	g.policy = policy
	g.updatedAt = time.Now()
	return nil
}

func (g *Group) UpdateDescription(description string) {
	g.description = description
	g.updatedAt = time.Now()
}
