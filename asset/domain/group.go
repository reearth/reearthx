package domain

import (
	"errors"
	"time"

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

var (
	ErrEmptyGroupName = errors.New("group name is required")
	ErrEmptyPolicy    = errors.New("policy is required")
)

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
func (g *Group) UpdateName(name string) {
	g.name = name
	g.updatedAt = time.Now()
}

func (g *Group) UpdatePolicy(policy string) {
	g.policy = policy
	g.updatedAt = time.Now()
}

func (g *Group) UpdateDescription(description string) {
	g.description = description
	g.updatedAt = time.Now()
}
