//go:generate go run github.com/Khan/genqlient

package accountproxy

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain/user"

	_ "github.com/Khan/genqlient/generate"
	"github.com/Khan/genqlient/graphql"
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
)

type User struct {
	http     HTTPClient
	gql      graphql.Client
	endpoint string
}

func NewUser(endpoint string, h HTTPClient) *User {
	return &User{
		http:     h,
		endpoint: endpoint,
		gql:      graphql.NewClient(endpoint, h),
	}
}

func (u *User) Fetch(ctx context.Context, ids accountdomain.UserIDList, op *accountusecase.Operator) ([]*user.User, error) {
	panic("not implemented")
}

func (u *User) FetchSimple(ctx context.Context, ids accountdomain.UserIDList, op *accountusecase.Operator) ([]*user.Simple, error) {
	return UserByIDsResponseTo(UserByIDs(ctx, u.gql, ids.Strings()))
}

func (*User) Signup(context.Context, accountinterfaces.SignupParam) (*user.User, error) {
	panic("not implemented")
}

func (*User) SignupOIDC(context.Context, accountinterfaces.SignupOIDCParam) (*user.User, error) {
	panic("not implemented")
}

func (*User) FindOrCreate(context.Context, accountinterfaces.UserFindOrCreateParam) (*user.User, error) {
	panic("not implemented")
}

func (*User) UpdateMe(context.Context, accountinterfaces.UpdateMeParam, *accountusecase.Operator) (*user.User, error) {
	panic("not implemented")
}

func (*User) RemoveMyAuth(context.Context, string, *accountusecase.Operator) (*user.User, error) {
	panic("not implemented")
}

func (*User) SearchUser(context.Context, string, *accountusecase.Operator) (*user.User, error) {
	panic("not implemented")
}

func (*User) DeleteMe(context.Context, accountdomain.UserID, *accountusecase.Operator) error {
	panic("not implemented")
}

func (*User) CreateVerification(context.Context, string) error {
	panic("not implemented")
}

func (*User) VerifyUser(context.Context, string) (*user.User, error) {
	panic("not implemented")
}

func (*User) StartPasswordReset(context.Context, string) error {
	panic("not implemented")
}

func (*User) PasswordReset(context.Context, string, string) error {
	panic("not implemented")
}
