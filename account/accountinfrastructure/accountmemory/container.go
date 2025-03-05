package accountmemory

import (
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/usecasex"
)

func New() *accountrepo.Container {
	return &accountrepo.Container{
		User:        NewUser(),
		Workspace:   NewWorkspace(),
		Transaction: &usecasex.NopTransaction{},
	}
}
