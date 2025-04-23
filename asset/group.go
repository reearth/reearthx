package asset

import (
	"time"
)

type Group struct {
	ID        GroupID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	PolicyID  *PolicyID
}

func NewGroup(name string) *Group {
	now := time.Now()
	return &Group{
		ID:        NewGroupID(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (g *Group) Clone() *Group {
	if g == nil {
		return nil
	}

	clone := *g

	if g.PolicyID != nil {
		policyID := *g.PolicyID
		clone.PolicyID = &policyID
	}

	return &clone
}

func (g *Group) AssignPolicy(policyID *PolicyID) {
	g.PolicyID = policyID
	g.UpdatedAt = time.Now()
}
