package repo

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/request"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/usecasex"
)

type RequestFilter struct {
	Keyword   *string
	Reviewer  *accountdomain.UserID
	CreatedBy *accountdomain.UserID
	State     []request.State
}

type Request interface {
	Filtered(ProjectFilter) Request
	FindByProject(
		context.Context,
		id.ProjectID,
		RequestFilter,
		*usecasex.Sort,
		*usecasex.Pagination,
	) (request.List, *usecasex.PageInfo, error)
	FindByID(context.Context, id.RequestID) (*request.Request, error)
	FindByIDs(context.Context, id.RequestIDList) (request.List, error)
	FindByItems(context.Context, id.ItemIDList, *RequestFilter) (request.List, error)
	Save(context.Context, *request.Request) error
	SaveAll(context.Context, id.ProjectID, request.List) error
}
