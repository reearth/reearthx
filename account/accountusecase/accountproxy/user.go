//go:generate go run github.com/Khan/genqlient

package accountproxy

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/util"

	_ "github.com/Khan/genqlient/generate"
	"github.com/Khan/genqlient/graphql"
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

func (u *User) FetchBySub(ctx context.Context, sub string) (*user.User, error) {
	panic("unsupported")
}

func (u *User) FetchByID(ctx context.Context, ids user.IDList) (user.List, error) {
	return UserByIDsResponseTo(UserByIDs(ctx, u.gql, ids.Strings()))
}

func (u *User) Signup(ctx context.Context, param accountinterfaces.SignupParam) (*user.User, error) {
	input := SignUpInput{
		Id:          param.UserID.String(),
		WorkspaceID: param.WorkspaceID.String(),
		Name:        param.Name,
		Email:       param.Email,
		Password:    param.Password,
		Secret:      *param.Secret,
		Lang:        param.Lang.String(),
		Theme:       string(*param.Theme),
	}
	res, err := SignUp(ctx, u.gql, input)
	if err != nil {
		return nil, err
	}
	return FragmentToUser(res.SignUp.User.FragmentUser)
}

func (u *User) SignupOIDC(ctx context.Context, param accountinterfaces.SignupOIDCParam) (*user.User, error) {
	input := SignupOIDCInput{
		Name:   param.Name,
		Email:  param.Email,
		Secret: *param.Secret,
		Sub:    param.Sub,
	}
	res, err := SignupOIDC(ctx, u.gql, input)
	if err != nil {
		return nil, err
	}
	return FragmentToUser(res.SignUpOIDC.User.FragmentUser)
}

func (u *User) FindOrCreate(ctx context.Context, param accountinterfaces.UserFindOrCreateParam) (*user.User, error) {
	input := FindOrCreateInput{
		Sub:   param.Sub,
		Iss:   param.ISS,
		Token: param.Token,
	}
	res, err := FindOrCreate(ctx, u.gql, input)
	if err != nil {
		return nil, err
	}
	return FragmentToUser(res.FindOrCreate.User.FragmentUser)
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
	return MeToUser(res.UpdateMe.Me.FragmentMe)
}

func (u *User) RemoveMyAuth(ctx context.Context, auth string, op *accountusecase.Operator) (*user.User, error) {
	res, err := RemoveMyAuth(ctx, u.gql, RemoveMyAuthInput{Auth: auth})
	if err != nil {
		return nil, err
	}
	return MeToUser(res.RemoveMyAuth.Me.FragmentMe)
}

func (u *User) FetchByNameOrEmail(ctx context.Context, nameOrEmail string) (*user.Simple, error) {
	res, err := UserByNameOrEmail(ctx, u.gql, nameOrEmail)
	if err != nil {
		return nil, err
	}
	r, err := FragmentToUser(res.UserByNameOrEmail.FragmentUser)
	if err != nil {
		return nil, err
	}
	return user.SimpleFrom(r), nil
}

func (u *User) SearchUser(ctx context.Context, keyword string) (user.SimpleList, error) {
	res, err := SearchUser(ctx, u.gql, keyword)
	if err != nil {
		return nil, err
	}
	r, err := util.TryMap(res.SearchUser, func(u SearchUserSearchUser) (*user.User, error) {
		return FragmentToUser(u.FragmentUser)
	})
	if err != nil {
		return nil, err
	}
	return util.Map(r, func(u *user.User) *user.Simple {
		return user.SimpleFrom(u)
	}), nil
}

func (u *User) DeleteMe(ctx context.Context, id user.ID, op *accountusecase.Operator) error {
	_, err := DeleteMe(ctx, u.gql, DeleteMeInput{UserId: id.String()})
	if err != nil {
		return err
	}
	return nil
}

func (u *User) CreateVerification(ctx context.Context, email string) error {
	_, err := CreateVerification(ctx, u.gql, CreateVerificationInput{Email: email})
	if err != nil {
		return err
	}
	return nil
}

func (u *User) VerifyUser(ctx context.Context, code string) (*user.User, error) {
	res, err := VerifyUser(ctx, u.gql, VerifyUserInput{Code: code})
	if err != nil {
		return nil, err
	}
	return FragmentToUser(res.VerifyUser.User.FragmentUser)

}

func (u *User) StartPasswordReset(ctx context.Context, email string) error {
	_, err := StartPasswordReset(ctx, u.gql, StartPasswordResetInput{Email: email})
	if err != nil {
		return err
	}
	return nil
}

func (u *User) PasswordReset(ctx context.Context, password string, token string) error {
	_, err := PasswordReset(ctx, u.gql, PasswordResetInput{Password: password, Token: token})
	if err != nil {
		return err
	}
	return nil
}
