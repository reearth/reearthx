package asset

import (
	"context"
)

type PolicyRepository interface {
	Save(ctx context.Context, policy *Policy) error
	FindByID(ctx context.Context, id PolicyID) (*Policy, error)
	Delete(ctx context.Context, id PolicyID) error
}
