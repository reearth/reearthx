package accountrepo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
)

type Workspace interface {
	Filtered(WorkspaceFilter) Workspace
	FindByID(context.Context, workspace.ID) (*workspace.Workspace, error)
	FindByIDs(context.Context, workspace.IDList) (workspace.List, error)
	FindByUser(context.Context, user.ID) (workspace.List, error)
	FindByIntegration(context.Context, workspace.IntegrationID) (workspace.List, error)
	// FindByIntegrations finds workspace list based on integrations IDs
	FindByIntegrations(context.Context, workspace.IntegrationIDList) (workspace.List, error)
	Create(context.Context, *workspace.Workspace) error
	Save(context.Context, *workspace.Workspace) error
	SaveAll(context.Context, workspace.List) error
	Remove(context.Context, workspace.ID) error
	RemoveAll(context.Context, workspace.IDList) error
}
