package accountrepo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
)

var ErrDuplicatedUser = rerror.NewE(i18n.T("duplicated user"))

type User interface {
	UserQuery
	FindByVerification(context.Context, string) (*user.User, error)
	FindByPasswordResetRequest(context.Context, string) (*user.User, error)
	FindBySubOrCreate(context.Context, *user.User, string) (*user.User, error)
	Create(context.Context, *user.User) error
	Save(context.Context, *user.User) error
	Remove(context.Context, user.ID) error
}

type UserQuery interface {
	FindAll(context.Context) (user.List, error)
	FindByID(context.Context, user.ID) (*user.User, error)
	FindByIDs(context.Context, user.IDList) (user.List, error)
	FindByIDsWithPagination(context.Context, user.IDList, *usecasex.Pagination, ...string) (user.List, *usecasex.PageInfo, error)
	FindBySub(context.Context, string) (*user.User, error)
	FindByEmail(context.Context, string) (*user.User, error)
	FindByName(context.Context, string) (*user.User, error)
	FindByAlias(context.Context, string) (*user.User, error)
	FindByNameOrEmail(context.Context, string) (*user.User, error)
	SearchByKeyword(context.Context, string, ...string) (user.List, error)
}
