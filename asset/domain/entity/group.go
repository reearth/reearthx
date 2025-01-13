package entity

import (
	"context"
	"time"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/validation"
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

// Validate implements the Validator interface
func (g *Group) Validate(ctx context.Context) validation.ValidationResult {
	validationCtx := validation.NewValidationContext(
		&validation.RequiredRule{Field: "id"},
		&validation.RequiredRule{Field: "name"},
		&validation.MaxLengthRule{Field: "name", MaxLength: 100},
		&validation.RequiredRule{Field: "policy"},
		&validation.MaxLengthRule{Field: "description", MaxLength: 500},
	)

	// Create a map of fields to validate
	fields := map[string]interface{}{
		"id":          g.id,
		"name":        g.name,
		"policy":      g.policy,
		"description": g.description,
	}

	return validationCtx.Validate(ctx, fields)
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

func (g *Group) UpdateDescription(description string) error {
	g.description = description
	g.updatedAt = time.Now()
	return nil
}
