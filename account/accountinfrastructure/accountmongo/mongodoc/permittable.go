// TODO: Delete this file once the permission check migration is complete.

package mongodoc

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/permittable"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/mongox"
)

type PermittableDocument struct {
	ID      string
	UserID  string
	RoleIDs []string
}

type PermittableConsumer = mongox.SliceFuncConsumer[*PermittableDocument, *permittable.Permittable]

func NewPermittableConsumer() *PermittableConsumer {
	return NewConsumer[*PermittableDocument, *permittable.Permittable]()
}

func NewPermittable(p permittable.Permittable) (*PermittableDocument, string) {
	id := p.ID().String()

	roleIds := make([]string, 0, len(p.RoleIDs()))
	for _, r := range p.RoleIDs() {
		roleIds = append(roleIds, r.String())
	}

	return &PermittableDocument{
		ID:      id,
		UserID:  p.UserID().String(),
		RoleIDs: roleIds,
	}, id
}

func (d *PermittableDocument) Model() (*permittable.Permittable, error) {
	if d == nil {
		return nil, nil
	}

	uid, err := accountdomain.PermittableIDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	userId, err := user.IDFrom(d.UserID)
	if err != nil {
		return nil, err
	}

	roleIds, err := accountdomain.RoleIDListFrom(d.RoleIDs)
	if err != nil {
		return nil, err
	}

	return permittable.New().
		ID(uid).
		UserID(userId).
		RoleIDs(roleIds).
		Build()
}
