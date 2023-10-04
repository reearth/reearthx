package accountinterfaces

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"golang.org/x/text/language"
)

var (
	ErrUserInvalidPasswordConfirmation = rerror.NewE(i18n.T("invalid password confirmation"))
	ErrUserInvalidPasswordReset        = rerror.NewE(i18n.T("invalid password reset request"))
	ErrUserInvalidLang                 = rerror.NewE(i18n.T("invalid lang"))
	ErrSignupInvalidSecret             = rerror.NewE(i18n.T("invalid secret"))
	ErrInvalidUserEmail                = rerror.NewE(i18n.T("invalid email"))
	ErrNotVerifiedUser                 = rerror.NewE(i18n.T("not verified user"))
	ErrInvalidEmailOrPassword          = rerror.NewE(i18n.T("invalid email or password"))
	ErrUserAlreadyExists               = rerror.NewE(i18n.T("user already exists"))
)

type SignupOIDCParam struct {
	AccessToken string
	Issuer      string
	Sub         string
	Email       string
	Name        string
	Secret      *string
	User        SignupUserParam
}

type SignupUserParam struct {
	UserID      *user.ID
	Lang        *language.Tag
	Theme       *user.Theme
	WorkspaceID *workspace.ID
}

type SignupParam struct {
	Email       string
	Name        string
	Password    string
	Secret      *string
	Lang        *language.Tag
	Theme       *user.Theme
	UserID      *user.ID
	WorkspaceID *workspace.ID
}

type UserFindOrCreateParam struct {
	Sub   string
	ISS   string
	Token string
}

type GetUserByCredentials struct {
	Email    string
	Password string
}

type UpdateMeParam struct {
	Name                 *string
	Email                *string
	Lang                 *language.Tag
	Theme                *user.Theme
	Password             *string
	PasswordConfirmation *string
}

type User interface {
	Fetch(context.Context, user.IDList, *accountusecase.Operator) (user.List, error)
	FetchSimple(context.Context, user.IDList, *accountusecase.Operator) (user.SimpleList, error)
	Signup(context.Context, SignupParam) (*user.User, error)
	SignupOIDC(context.Context, SignupOIDCParam) (*user.User, error)
	FindOrCreate(context.Context, UserFindOrCreateParam) (*user.User, error)
	UpdateMe(context.Context, UpdateMeParam, *accountusecase.Operator) (*user.User, error)
	RemoveMyAuth(context.Context, string, *accountusecase.Operator) (*user.User, error)
	SearchUser(context.Context, string, *accountusecase.Operator) (*user.User, error)
	DeleteMe(context.Context, user.ID, *accountusecase.Operator) error

	// from reearth/server
	CreateVerification(context.Context, string) error
	VerifyUser(context.Context, string) (*user.User, error)
	StartPasswordReset(context.Context, string) error
	PasswordReset(context.Context, string, string) error
}
