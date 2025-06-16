package interfaces

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/group"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/model"
	"github.com/reearth/reearthx/asset/usecase"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

var ErrDelGroupUsed = rerror.NewE(i18n.T("can't delete a group as it's used by some models"))

type CreateGroupParam struct {
	Description *string
	Name        string
	Key         string
	ProjectId   id.ProjectID
}

type UpdateGroupParam struct {
	Name        *string
	Description *string
	Key         *string
	GroupID     id.GroupID
}

type Group interface {
	FindByID(context.Context, id.GroupID, *usecase.Operator) (*group.Group, error)
	FindByIDs(context.Context, id.GroupIDList, *usecase.Operator) (group.List, error)
	Filter(
		context.Context,
		id.ProjectID,
		*group.Sort,
		*usecasex.Pagination,
		*usecase.Operator,
	) (group.List, *usecasex.PageInfo, error)
	FindByProject(context.Context, id.ProjectID, *usecase.Operator) (group.List, error)
	FindByModel(context.Context, id.ModelID, *usecase.Operator) (group.List, error)
	FindByKey(context.Context, id.ProjectID, string, *usecase.Operator) (*group.Group, error)
	FindByIDOrKey(
		context.Context,
		id.ProjectID,
		group.IDOrKey,
		*usecase.Operator,
	) (*group.Group, error)
	Create(context.Context, CreateGroupParam, *usecase.Operator) (*group.Group, error)
	Update(context.Context, UpdateGroupParam, *usecase.Operator) (*group.Group, error)
	UpdateOrder(context.Context, id.GroupIDList, *usecase.Operator) (group.List, error)
	CheckKey(context.Context, id.ProjectID, string) (bool, error)
	FindModelsByGroup(context.Context, id.GroupID, *usecase.Operator) (model.List, error)
	Delete(context.Context, id.GroupID, *usecase.Operator) error
}
