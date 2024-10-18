package accountinterfaces

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain/user"
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
	Fetch(context.Context, workspace.IDList, *accountusecase.Operator) (workspace.List, error)
	FindByUser(context.Context, user.ID, *accountusecase.Operator) (workspace.List, error)
	Create(context.Context, string, user.ID, *accountusecase.Operator) (*workspace.Workspace, error)
	Update(context.Context, workspace.ID, string, *accountusecase.Operator) (*workspace.Workspace, error)
	AddUserMember(context.Context, workspace.ID, map[user.ID]workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	AddIntegrationMember(context.Context, workspace.ID, workspace.IntegrationID, workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	UpdateUserMember(context.Context, workspace.ID, user.ID, workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	UpdateIntegration(context.Context, workspace.ID, workspace.IntegrationID, workspace.Role, *accountusecase.Operator) (*workspace.Workspace, error)
	RemoveUserMember(context.Context, workspace.ID, user.ID, *accountusecase.Operator) (*workspace.Workspace, error)
	RemoveMultipleUserMembers(context.Context, workspace.ID, user.IDList, *accountusecase.Operator) (*workspace.Workspace, error)
	RemoveIntegration(context.Context, workspace.ID, workspace.IntegrationID, *accountusecase.Operator) (*workspace.Workspace, error)
	Remove(context.Context, workspace.ID, *accountusecase.Operator) error
}
