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

func NewUser(endpoint string, h HTTPClient) accountinterfaces.User {
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

func (u *User) UpdateMe(ctx context.Context, param accountinterfaces.UpdateMeParam, op *accountusecase.Operator) (*user.User, error) {
	input := UpdateMeInput{
		Name:                 *param.Name,
		Email:                *param.Email,
		Lang:                 param.Lang.String(),
		Theme:                string(*param.Theme),
		Password:             *param.Password,
		PasswordConfirmation: *param.PasswordConfirmation,
	}
	res, err := UpdateMe(ctx, u.gql, input)
	if err != nil {
		return nil, err
	}
	return MeToUser(res.UpdateMe.Me.TemplateMe)
}

func (u *User) RemoveMyAuth(ctx context.Context, auth string, op *accountusecase.Operator) (*user.User, error) {
	res, err := RemoveMyAuth(ctx, u.gql, RemoveMyAuthInput{Auth: auth})
	if err != nil {
		return nil, err
	}
	return MeToUser(res.RemoveMyAuth.Me.TemplateMe)
}

func (*User) SearchUser(context.Context, string, *accountusecase.Operator) (*user.User, error) {
	panic("not implemented")
}

func (u *User) DeleteMe(ctx context.Context, id accountdomain.UserID, op *accountusecase.Operator) error {
	_, err := DeleteMe(ctx, u.gql, DeleteMeInput{UserId: id.String()})
	if err != nil {
		return err
	}
	return nil
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
