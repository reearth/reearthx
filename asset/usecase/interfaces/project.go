package interfaces

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/project"
	"github.com/reearth/reearthx/asset/usecase"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

type CreateProjectParam struct {
	Name         *string
	Description  *string
	Alias        *string
	RequestRoles []workspace.Role
	WorkspaceID  accountdomain.WorkspaceID
}

type UpdateProjectParam struct {
	Name         *string
	Description  *string
	Alias        *string
	Publication  *UpdateProjectPublicationParam
	RequestRoles []workspace.Role
	ID           id.ProjectID
}

type UpdateProjectPublicationParam struct {
	Scope       *project.PublicationScope
	AssetPublic *bool
}

var (
	ErrProjectAliasIsNotSet    error = rerror.NewE(i18n.T("project alias is not set"))
	ErrProjectAliasAlreadyUsed error = rerror.NewE(
		i18n.T("project alias is already used by another project"),
	)
	ErrInvalidProject = rerror.NewE(i18n.T("invalid project"))
)

type Project interface {
	Fetch(context.Context, []id.ProjectID, *usecase.Operator) (project.List, error)
	FindByIDOrAlias(context.Context, project.IDOrAlias, *usecase.Operator) (*project.Project, error)
	FindByWorkspace(
		context.Context,
		accountdomain.WorkspaceID,
		*usecasex.Pagination,
		*usecase.Operator,
	) (project.List, *usecasex.PageInfo, error)
	Create(context.Context, CreateProjectParam, *usecase.Operator) (*project.Project, error)
	Update(context.Context, UpdateProjectParam, *usecase.Operator) (*project.Project, error)
	CheckAlias(context.Context, string) (bool, error)
	Delete(context.Context, id.ProjectID, *usecase.Operator) error
	RegenerateToken(context.Context, id.ProjectID, *usecase.Operator) (*project.Project, error)
}
