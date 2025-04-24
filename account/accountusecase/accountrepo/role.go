// TODO: Delete this file once the permission check migration is complete.

package accountrepo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/role"
)

type Role interface {
	FindAll(context.Context) (role.List, error)
	FindByID(context.Context, accountdomain.RoleID) (*role.Role, error)
	FindByIDs(context.Context, accountdomain.RoleIDList) (role.List, error)
	Save(context.Context, role.Role) error
	Remove(context.Context, accountdomain.RoleID) error
}
