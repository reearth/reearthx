package accountrepo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
)

type Workspace interface {
	FindByID(context.Context, workspace.ID) (*workspace.Workspace, error)
	FindByIDs(context.Context, workspace.IDList) (workspace.List, error)
	FindByUser(context.Context, user.ID) (workspace.List, error)
	FindByIntegration(context.Context, workspace.IntegrationID) (workspace.List, error)
	Save(context.Context, *workspace.Workspace) error
	SaveAll(context.Context, workspace.List) error
	Remove(context.Context, workspace.ID) error
	RemoveAll(context.Context, workspace.IDList) error
}
