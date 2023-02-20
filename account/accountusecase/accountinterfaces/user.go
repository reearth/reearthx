package accountinterfaces

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
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

type SignupOIDC struct {
	Email  string
	Name   string
	Secret *string
	Sub    string
}
type SignupParam struct {
	Email       string
	Name        string
	Password    string
	Secret      *string
	Lang        *language.Tag
	Theme       *user.Theme
	UserID      *accountdomain.UserID
	WorkspaceID *accountdomain.WorkspaceID
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
	Fetch(context.Context, []accountdomain.UserID, *accountusecase.Operator) ([]*user.User, error)
	Signup(context.Context, SignupParam) (*user.User, error)
	SignupOIDC(context.Context, SignupOIDC) (*user.User, error)
	FindOrCreate(context.Context, UserFindOrCreateParam) (*user.User, error)
	UpdateMe(context.Context, UpdateMeParam, *accountusecase.Operator) (*user.User, error)
	RemoveMyAuth(context.Context, string, *accountusecase.Operator) (*user.User, error)
	SearchUser(context.Context, string, *accountusecase.Operator) (*user.User, error)
	DeleteMe(context.Context, accountdomain.UserID, *accountusecase.Operator) error

	// from reearth/server
	CreateVerification(context.Context, string) error
	VerifyUser(context.Context, string) (*user.User, error)
	StartPasswordReset(context.Context, string) error
	PasswordReset(context.Context, string, string) error
}
