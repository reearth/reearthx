package repo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
)

type Workspace interface {
	FindByID(context.Context, accountdomain.WorkspaceID) (*workspace.Workspace, error)
	FindByIDs(context.Context, accountdomain.WorkspaceIDList) (workspace.WorkspaceList, error)
	FindByUser(context.Context, accountdomain.UserID) (workspace.WorkspaceList, error)
	FindByIntegration(context.Context, accountdomain.IntegrationID) (workspace.WorkspaceList, error)
	Save(context.Context, *workspace.Workspace) error
	SaveAll(context.Context, []*workspace.Workspace) error
	Remove(context.Context, accountdomain.WorkspaceID) error
	RemoveAll(context.Context, accountdomain.WorkspaceIDList) error
}
