package accountinteractor

import (
	"context"
	_ "embed"
	"errors"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

type User struct {
	repos           *accountrepo.Container
	gateways        *accountgateway.Container
	signupSecret    string
	authSrvUIDomain string
}

func NewUser(r *accountrepo.Container, g *accountgateway.Container, signupSecret, authSrcUIDomain string) accountinterfaces.User {
	return &User{
		repos:           r,
		gateways:        g,
		signupSecret:    signupSecret,
		authSrvUIDomain: authSrcUIDomain,
	}
}

func (i *User) Fetch(ctx context.Context, ids []user.ID, operator *accountusecase.Operator) ([]*user.User, error) {
	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func() ([]*user.User, error) {
		res, err := i.repos.User.FindByIDs(ctx, ids)
		if err != nil {
			return res, err
		}
		// filter
		for k, u := range res {
			workspaces, err := i.repos.Workspace.FindByUser(ctx, u.ID())
			if err != nil {
				return res, err
			}
			workspaceIDs := make([]user.WorkspaceID, 0, len(workspaces))
			for _, t := range workspaces {
				if t != nil {
					workspaceIDs = append(workspaceIDs, t.ID())
				}
			}
			if !operator.IsReadableWorkspace(workspaceIDs...) {
				res[k] = nil
			}
		}
		return res, nil
	})
}

func (i *User) GetUserByCredentials(ctx context.Context, inp accountinterfaces.GetUserByCredentials) (u *user.User, err error) {
	return Run1(ctx, nil, i.repos, Usecase().Transaction(), func() (*user.User, error) {
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
	return Run1(ctx, nil, i.repos, Usecase().Transaction(), func() (*user.User, error) {
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
	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func() (*user.User, error) {
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
				if _, err := i.gateways.Authenticator.UpdateUser(accountgateway.AuthenticatorUpdateUserParam{
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
	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func() (*user.User, error) {
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

func (i *User) SearchUser(ctx context.Context, nameOrEmail string, operator *accountusecase.Operator) (u *user.User, err error) {
	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func() (*user.User, error) {
		u, err = i.repos.User.FindByNameOrEmail(ctx, nameOrEmail)
		if err != nil && !errors.Is(err, rerror.ErrNotFound) {
			return nil, err
		}
		return u, nil
	})
}

func (i *User) DeleteMe(ctx context.Context, userID user.ID, operator *accountusecase.Operator) (err error) {
	if operator.User == nil {
		return accountinterfaces.ErrInvalidOperator
	}
	return Run0(ctx, operator, i.repos, Usecase().Transaction(), func() error {
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
