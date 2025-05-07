package asset

import (
	"context"
)

type PolicyService interface {
	CreatePolicy(ctx context.Context, name string, storageLimit int64) (*Policy, error)
	GetPolicy(ctx context.Context, id PolicyID) (*Policy, error)
	DeletePolicy(ctx context.Context, id PolicyID) error
}
