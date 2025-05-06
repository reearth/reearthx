package assetmemory

import (
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmemory"
	"time"

	repo "github.com/reearth/reearthx/asset/assetusecase/assetrepo"
	"github.com/reearth/reearthx/usecasex"
)

func New() *repo.Container {
	return &repo.Container{
		Asset:       NewAsset(),
		AssetFile:   NewAssetFile(),
		User:        accountmemory.NewUser(),
		Workspace:   accountmemory.NewWorkspace(),
		Integration: NewIntegration(),
		Project:     NewProject(),
		Thread:      NewThread(),
		Event:       NewEvent(),
		Transaction: &usecasex.NopTransaction{},
	}
}

func MockNow(r *repo.Container, t time.Time) func() {
	p := r.Project.(*Project).now.Mock(t)

	return func() {
		p()
	}
}
