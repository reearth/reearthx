// TODO: Delete this file once the permission check migration is complete.

package accountrepo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/permittable"
	"github.com/reearth/reearthx/account/accountdomain/user"
)

type Permittable interface {
	FindByUserID(context.Context, user.ID) (*permittable.Permittable, error)
	FindByUserIDs(context.Context, user.IDList) (permittable.List, error)
	FindByRoleID(context.Context, accountdomain.RoleID) (permittable.List, error)
	Save(context.Context, permittable.Permittable) error
}
