package accountmemory

import (
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/usecasex"
)

func New() *accountrepo.Container {
	return &accountrepo.Container{
		User:        NewUser(),
		Workspace:   NewWorkspace(),
		Role:        NewRole(),        // TODO: Delete it once the permission check migration is complete.
		Permittable: NewPermittable(), // TODO: Delete it once the permission check migration is complete.
		Transaction: &usecasex.NopTransaction{},
	}
}
