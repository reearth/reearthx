package accountinterfaces

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

var (
	ErrOwnerCannotLeaveTheWorkspace = rerror.NewE(i18n.T("owner user cannot leave from the workspace"))
	ErrCannotChangeOwnerRole        = rerror.NewE(i18n.T("cannot change the role of the workspace owner"))
	ErrCannotDeleteWorkspace        = rerror.NewE(i18n.T("cannot delete workspace because at least one project is left"))
	ErrWorkspaceWithProjects        = rerror.NewE(i18n.T("target workspace still has some project"))
)

type Workspace interface {
	Fetch(context.Context, []accountdomain.WorkspaceID, *accountusecase.Operator) ([]*workspace.Workspace, error)
	FindByUser(context.Context, accountdomain.UserID, *accountusecase.Operator) ([]*workspace.Workspace, error)
	FetchPolicy(context.Context, []workspace.PolicyID, *accountusecase.Operator) ([]*workspace.Policy, error)
	Create(context.Context, string, accountdomain.UserID, *accountusecase.Operator) (*workspace.Workspace, error)
	Update(context.Context, accountdomain.WorkspaceID, string, *accountusecase.Operator) (*workspace.Workspace, error)
	AddUserMember(context.Context, accountdomain.WorkspaceID, map[accountdomain.UserID]workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	AddIntegrationMember(context.Context, accountdomain.WorkspaceID, accountdomain.IntegrationID, workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	UpdateUser(context.Context, accountdomain.WorkspaceID, accountdomain.UserID, workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	UpdateIntegration(context.Context, accountdomain.WorkspaceID, accountdomain.IntegrationID, workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	RemoveUser(context.Context, accountdomain.WorkspaceID, accountdomain.UserID, *accountusecase.Operator) (*workspace.Workspace, error)
	RemoveIntegration(context.Context, accountdomain.WorkspaceID, accountdomain.IntegrationID, *accountusecase.Operator) (*workspace.Workspace, error)
	Remove(context.Context, accountdomain.WorkspaceID, *accountusecase.Operator) error
}
