package repo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
)

type Workspace interface {
	FindByUser(context.Context, accountdomain.UserID) (workspace.WorkspaceList, error)
	FindByIDs(context.Context, accountdomain.WorkspaceIDList) (workspace.WorkspaceList, error)
	FindByID(context.Context, accountdomain.WorkspaceID) (*workspace.Workspace, error)
	Save(context.Context, *workspace.Workspace) error
	SaveAll(context.Context, []*workspace.Workspace) error
	Remove(context.Context, workspace.ID) error
	RemoveAll(context.Context, accountdomain.WorkspaceIDList) error
	FindByIntegration(context.Context, accountdomain.IntegrationID) (workspace.WorkspaceList, error)
}
