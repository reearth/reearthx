package repo

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/policy"
)

type Policy interface {
	FindByID(context.Context, policy.ID) (*policy.Policy, error)
	FindByIDs(context.Context, []policy.ID) ([]*policy.Policy, error)
}
