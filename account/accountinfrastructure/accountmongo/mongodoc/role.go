package mongodoc

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/role"
	"github.com/reearth/reearthx/mongox"
)

type RoleDocument struct {
	ID   string
	Name string
}

type RoleConsumer = mongox.SliceFuncConsumer[*RoleDocument, *role.Role]

func NewRoleConsumer() *RoleConsumer {
	return NewConsumer[*RoleDocument, *role.Role]()
}

func NewRole(g role.Role) (*RoleDocument, string) {
	id := g.ID().String()
	return &RoleDocument{
		ID:   id,
		Name: g.Name(),
	}, id
}

func (d *RoleDocument) Model() (*role.Role, error) {
	if d == nil {
		return nil, nil
	}

	rid, err := accountdomain.RoleIDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	return role.New().
		ID(rid).
		Name(d.Name).
		Build()
}
