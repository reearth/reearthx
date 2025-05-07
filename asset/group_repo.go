package asset

import (
	"context"
)

type GroupRepository interface {
	Save(ctx context.Context, group *Group) error
	FindByID(ctx context.Context, id GroupID) (*Group, error)
	Delete(ctx context.Context, id GroupID) error
	UpdatePolicy(ctx context.Context, id GroupID, policyID *PolicyID) error
}
