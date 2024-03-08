package accountinteractor

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	htmlTmpl "html/template"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/mailer"
	"github.com/reearth/reearthx/rerror"
)

type User struct {
	repos           *accountrepo.Container
	gateways        *accountgateway.Container
	signupSecret    string
	authSrvUIDomain string
	query           accountinterfaces.UserQuery
}

var (
	passwordResetMailContent = mailContent{
		Message:     "Thank you for using Re:Earth. We've received a request to reset your password. If this was you, please click the link below to confirm and change your password.",
		Suffix:      "If you did not mean to reset your password, then you can ignore this email.",
		ActionLabel: "Confirm to reset your password",
	}
)

func NewUser(r *accountrepo.Container, g *accountgateway.Container, signupSecret, authSrcUIDomain string) accountinterfaces.User {
	var repos []accountrepo.User
	if r != nil {
		repos = []accountrepo.User{r.User}
	}
	return &User{
		repos:           r,
		gateways:        g,
		signupSecret:    signupSecret,
		authSrvUIDomain: authSrcUIDomain,
		query: &UserQuery{
			repos: repos,
		},
	}
}

func NewMultiUser(r *accountrepo.Container, g *accountgateway.Container, signupSecret, authSrcUIDomain string, users []accountrepo.User) accountinterfaces.User {
	return &User{
		repos:           r,
		gateways:        g,
		signupSecret:    signupSecret,
		authSrvUIDomain: authSrcUIDomain,
		query: &UserQuery{
			repos: append([]accountrepo.User{r.User}, users...),
		},
	}
}

func (i *User) FetchByID(ctx context.Context, ids user.IDList) (user.List, error) {
	return i.query.FetchByID(ctx, ids)
}

func (i *User) FetchBySub(ctx context.Context, sub string) (*user.User, error) {
	return i.query.FetchBySub(ctx, sub)
}

func (i *User) SearchUser(ctx context.Context, nameOrEmail string) (*user.Simple, error) {
	return i.query.SearchUser(ctx, nameOrEmail)
}

func (i *User) GetUserByCredentials(ctx context.Context, inp accountinterfaces.GetUserByCredentials) (u *user.User, err error) {
	return Run1(ctx, nil, i.repos, Usecase().Transaction(), func(ctx context.Context) (*user.User, error) {
		u, err = i.repos.User.FindByNameOrEmail(ctx, inp.Email)
		if err != nil && !errors.Is(rerror.ErrNotFound, err) {
			return nil, err
		} else if u == nil {
			return nil, accountinterfaces.ErrInvalidUserEmail
		}
		matched, err := u.MatchPassword(inp.Password)
		if err != nil {
			return nil, err
		}
		if !matched {
			return nil, accountinterfaces.ErrInvalidEmailOrPassword
		}
		if u.Verification() == nil || !u.Verification().IsVerified() {
			return nil, accountinterfaces.ErrNotVerifiedUser
		}
		return u, nil
	})
}

func (i *User) GetUserBySubject(ctx context.Context, sub string) (u *user.User, err error) {
	return Run1(ctx, nil, i.repos, Usecase().Transaction(), func(ctx context.Context) (*user.User, error) {
		u, err = i.repos.User.FindBySub(ctx, sub)
		if err != nil {
			return nil, err
		}
		return u, nil
	})
}

func (i *User) UpdateMe(ctx context.Context, p accountinterfaces.UpdateMeParam, operator *accountusecase.Operator) (u *user.User, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) (*user.User, error) {
		if p.Password != nil {
			if p.PasswordConfirmation == nil || *p.Password != *p.PasswordConfirmation {
				return nil, accountinterfaces.ErrUserInvalidPasswordConfirmation
			}
		}

		var workspace *workspace.Workspace

		u, err = i.repos.User.FindByID(ctx, *operator.User)
		if err != nil {
			return nil, err
		}

		if p.Name != nil && *p.Name != u.Name() {
			oldName := u.Name()
			u.UpdateName(*p.Name)

			workspace, err = i.repos.Workspace.FindByID(ctx, u.Workspace())
			if err != nil && !errors.Is(err, rerror.ErrNotFound) {
				return nil, err
			}

			tn := workspace.Name()
			if tn == "" || tn == oldName {
				workspace.Rename(*p.Name)
			} else {
				workspace = nil
			}
		}
		if p.Email != nil {
			if err := u.UpdateEmail(*p.Email); err != nil {
				return nil, err
			}
		}
		if p.Lang != nil {
			u.UpdateLang(*p.Lang)
		}
		if p.Theme != nil {
			u.UpdateTheme(*p.Theme)
		}

		if p.Password != nil && u.HasAuthProvider("reearth") {
			if err := u.SetPassword(*p.Password); err != nil {
				return nil, err
			}
		}

		// Update Auth0 users
		if p.Name != nil || p.Email != nil || p.Password != nil {
			for _, a := range u.Auths() {
				if a.Provider != "auth0" {
					continue
				}
				if _, err := i.gateways.Authenticator.UpdateUser(ctx, accountgateway.AuthenticatorUpdateUserParam{
					ID:       a.Sub,
					Name:     p.Name,
					Email:    p.Email,
					Password: p.Password,
				}); err != nil {
					return nil, err
				}
			}
		}

		if workspace != nil {
			err = i.repos.Workspace.Save(ctx, workspace)
			if err != nil {
				return nil, err
			}
		}

		err = i.repos.User.Save(ctx, u)
		if err != nil {
			return nil, err
		}

		return u, nil
	})
}

func (i *User) RemoveMyAuth(ctx context.Context, authProvider string, operator *accountusecase.Operator) (u *user.User, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) (*user.User, error) {
		u, err = i.repos.User.FindByID(ctx, *operator.User)
		if err != nil {
			return nil, err
		}

		u.RemoveAuthByProvider(authProvider)

		err = i.repos.User.Save(ctx, u)
		if err != nil {
			return nil, err
		}

		return u, nil
	})
}

func (i *User) DeleteMe(ctx context.Context, userID user.ID, operator *accountusecase.Operator) (err error) {
	if operator.User == nil {
		return accountinterfaces.ErrInvalidOperator
	}
	return Run0(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) error {
		if userID.IsNil() || userID != *operator.User {
			return rerror.NewE(i18n.T("invalid user id"))
		}

		u, err := i.repos.User.FindByID(ctx, userID)
		if err != nil && !errors.Is(err, rerror.ErrNotFound) {
			return err
		}
		if u == nil {
			return nil
		}

		workspaces, err := i.repos.Workspace.FindByUser(ctx, u.ID())
		if err != nil {
			return err
		}

		updatedWorkspaces := make([]*workspace.Workspace, 0, len(workspaces))
		deletedWorkspaces := []user.WorkspaceID{}

		for _, workspace := range workspaces {
			if !workspace.IsPersonal() && !workspace.Members().IsOnlyOwner(u.ID()) {
				_ = workspace.Members().Leave(u.ID())
				updatedWorkspaces = append(updatedWorkspaces, workspace)
				continue
			}

			deletedWorkspaces = append(deletedWorkspaces, workspace.ID())
		}

		// Save workspaces
		if err := i.repos.Workspace.SaveAll(ctx, updatedWorkspaces); err != nil {
			return err
		}

		// Delete workspaces
		if err := i.repos.Workspace.RemoveAll(ctx, deletedWorkspaces); err != nil {
			return err
		}

		// Delete user
		if err := i.repos.User.Remove(ctx, u.ID()); err != nil {
			return err
		}

		return nil
	})

}

func (i *User) VerifyUser(ctx context.Context, code string) (*user.User, error) {
	return Run1(ctx, nil, i.repos, Usecase().Transaction(), func(ctx context.Context) (*user.User, error) {

		u, err := i.repos.User.FindByVerification(ctx, code)
		if err != nil {
			return nil, err
		}
		if u.Verification().IsExpired() {
			return nil, errors.New("verification expired")
		}
		u.Verification().SetVerified(true)
		err = i.repos.User.Save(ctx, u)
		if err != nil {
			return nil, err
		}

		return u, nil
	})
}
func (i *User) StartPasswordReset(ctx context.Context, email string) error {
	return Run0(ctx, nil, i.repos, Usecase().Transaction(), func(ctx context.Context) error {
		u, err := i.repos.User.FindByEmail(ctx, email)
		if err != nil {
			return err
		}

		a := u.Auths().GetByProvider(user.ProviderReearth)
		if a == nil || a.Sub == "" {
			return accountinterfaces.ErrUserInvalidPasswordReset
		}

		pr := user.NewPasswordReset()
		u.SetPasswordReset(pr)

		if err := i.repos.User.Save(ctx, u); err != nil {
			return err
		}

		var TextOut, HTMLOut bytes.Buffer
		link := i.authSrvUIDomain + "/?pwd-reset-token=" + pr.Token
		passwordResetMailContent.UserName = u.Name()
		passwordResetMailContent.ActionURL = htmlTmpl.URL(link)

		if err := authTextTMPL.Execute(&TextOut, passwordResetMailContent); err != nil {
			return err
		}
		if err := authHTMLTMPL.Execute(&HTMLOut, passwordResetMailContent); err != nil {
			return err
		}

		err = i.gateways.Mailer.SendMail(ctx, []mailer.Contact{
			{
				Email: u.Email(),
				Name:  u.Name(),
			},
		}, "Password reset", TextOut.String(), HTMLOut.String())
		if err != nil {
			return err
		}

		return nil
	})
}

func (i *User) PasswordReset(ctx context.Context, password string, token string) error {
	return Run0(ctx, nil, i.repos, Usecase().Transaction(), func(ctx context.Context) error {
		u, err := i.repos.User.FindByPasswordResetRequest(ctx, token)
		if err != nil {
			return err
		}

		passwordReset := u.PasswordReset()
		ok := passwordReset.Validate(token)
		if !ok {
			return accountinterfaces.ErrUserInvalidPasswordReset
		}

		a := u.Auths().GetByProvider(user.ProviderReearth)
		if a == nil || a.Sub == "" {
			return accountinterfaces.ErrUserInvalidPasswordReset
		}

		if err := u.SetPassword(password); err != nil {
			return err
		}

		u.SetPasswordReset(nil)

		if err := i.repos.User.Save(ctx, u); err != nil {
			return err
		}

		return nil
	})
}

type UserQuery struct {
	repos []accountrepo.User
}

func NewUserQuery(primary accountrepo.User, repos ...accountrepo.User) *UserQuery {
	return &UserQuery{
		repos: append([]accountrepo.User{primary}, repos...),
	}
}

func (q *UserQuery) FetchByID(ctx context.Context, ids user.IDList) (user.List, error) {
	var us user.List
	for _, r := range q.repos {
		u, err := r.FindByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		us = append(us, u...)
	}
	return us, nil
}

func (q *UserQuery) FetchBySub(ctx context.Context, sub string) (*user.User, error) {
	for _, r := range q.repos {
		u, err := r.FindBySub(ctx, sub)
		if errors.Is(err, rerror.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	return nil, rerror.ErrNotFound
}

func (q *UserQuery) SearchUser(ctx context.Context, nameOrEmail string) (*user.Simple, error) {
	for _, r := range q.repos {
		u, err := r.FindByNameOrEmail(ctx, nameOrEmail)
		if errors.Is(err, rerror.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return user.SimpleFrom(u), nil
	}
	return nil, nil
}
