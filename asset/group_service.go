package asset

import (
	"context"
)

type GroupService interface {
	CreateGroup(ctx context.Context, name string) (*Group, error)
	GetGroup(ctx context.Context, id GroupID) (*Group, error)
	DeleteGroup(ctx context.Context, id GroupID) error
	AssignPolicy(ctx context.Context, groupID GroupID, policyID *PolicyID) error
}
