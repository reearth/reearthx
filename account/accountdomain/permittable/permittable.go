package permittable

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/role"
	"github.com/reearth/reearthx/account/accountdomain/user"
)

type Permittable struct {
	id      ID
	userID  user.ID
	roleIDs []role.ID
}

func (p *Permittable) ID() ID {
	if p == nil {
		return ID{}
	}
	return p.id
}

func (p *Permittable) UserID() user.ID {
	if p == nil {
		return user.ID{}
	}
	return p.userID
}

func (p *Permittable) RoleIDs() []accountdomain.RoleID {
	if p == nil {
		return nil
	}
	return p.roleIDs
}

func (p *Permittable) EditRoleIDs(roleIDs accountdomain.RoleIDList) {
	if p == nil {
		return
	}
	p.roleIDs = roleIDs
}
