package interfaces

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/request"
	"github.com/reearth/reearthx/asset/usecase"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

var ErrAlreadyPublished = rerror.NewE(i18n.T("already published"))

type CreateRequestParam struct {
	Description *string
	State       *request.State
	Title       string
	Reviewers   accountdomain.UserIDList
	Items       request.ItemList
	ProjectID   id.ProjectID
}

type UpdateRequestParam struct {
	Title       *string
	Description *string
	State       *request.State
	Reviewers   accountdomain.UserIDList
	Items       request.ItemList
	RequestID   id.RequestID
}

type RequestFilter struct {
	Keyword   *string
	Reviewer  *accountdomain.UserID
	CreatedBy *accountdomain.UserID
	State     []request.State
}

type Request interface {
	FindByID(context.Context, id.RequestID, *usecase.Operator) (*request.Request, error)
	FindByIDs(context.Context, id.RequestIDList, *usecase.Operator) (request.List, error)
	FindByProject(context.Context, id.ProjectID, RequestFilter, *usecasex.Sort, *usecasex.Pagination, *usecase.Operator) (request.List, *usecasex.PageInfo, error)
	FindByItem(context.Context, id.ItemID, *RequestFilter, *usecase.Operator) (request.List, error)
	Create(context.Context, CreateRequestParam, *usecase.Operator) (*request.Request, error)
	Update(context.Context, UpdateRequestParam, *usecase.Operator) (*request.Request, error)
	Approve(context.Context, id.RequestID, *usecase.Operator) (*request.Request, error)
	CloseAll(context.Context, id.ProjectID, id.RequestIDList, *usecase.Operator) error
}
