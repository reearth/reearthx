package memory

import (
	"github.com/reearth/reearthx/account/accountusecase/repo"
	"github.com/reearth/reearthx/usecasex"
)

func New() *repo.Container {
	return &repo.Container{
		User:        NewUser(),
		Workspace:   NewWorkspace(),
		Transaction: &usecasex.NopTransaction{},
	}
}
