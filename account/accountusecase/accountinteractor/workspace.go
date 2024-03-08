package accountinteractor

import (
	"context"
	"strings"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

type WorkspaceMemberCountEnforcer func(context.Context, *workspace.Workspace, user.List, *accountusecase.Operator) error

type Workspace struct {
	repos              *accountrepo.Container
	enforceMemberCount WorkspaceMemberCountEnforcer
	userquery          accountinterfaces.UserQuery
}

func NewWorkspace(r *accountrepo.Container, enforceMemberCount WorkspaceMemberCountEnforcer) accountinterfaces.Workspace {
	return &Workspace{
		repos:              r,
		enforceMemberCount: enforceMemberCount,
		userquery:          NewUserQuery(r.User, r.Users...),
	}
}

func (i *Workspace) Fetch(ctx context.Context, ids workspace.IDList, operator *accountusecase.Operator) (workspace.List, error) {
	res, err := i.repos.Workspace.FindByIDs(ctx, ids)
	return filterWorkspaces(res, operator, err, false, true)
}

func (i *Workspace) FindByUser(ctx context.Context, id workspace.UserID, operator *accountusecase.Operator) (workspace.List, error) {
	res, err := i.repos.Workspace.FindByUser(ctx, id)
	return filterWorkspaces(res, operator, err, true, true)
}

func (i *Workspace) Create(ctx context.Context, name string, firstUser workspace.UserID, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		if len(strings.TrimSpace(name)) == 0 {
			return nil, user.ErrInvalidName
		}

		firstUsers, err := i.userquery.FetchByID(ctx, []user.ID{firstUser})
		if err != nil || len(firstUsers) == 0 {
			if err == nil {
				return nil, rerror.ErrNotFound
			}
			return nil, err
		}

		ws, err := workspace.New().
			NewID().
			Name(name).
			Build()
		if err != nil {
			return nil, err
		}

		if err := ws.Members().Join(firstUsers[0], workspace.RoleOwner, *operator.User); err != nil {
			return nil, err
		}

		if err := i.repos.Workspace.Save(ctx, ws); err != nil {
			return nil, err
		}

		operator.AddNewWorkspace(ws.ID())
		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) Update(ctx context.Context, id workspace.ID, name string, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		if ws.IsPersonal() {
			return nil, workspace.ErrCannotModifyPersonalWorkspace
		}
		if ws.Members().UserRole(*operator.User) != workspace.RoleOwner {
			return nil, accountinterfaces.ErrOperationDenied
		}

		if len(strings.TrimSpace(name)) == 0 {
			return nil, user.ErrInvalidName
		}

		ws.Rename(name)

		err = i.repos.Workspace.Save(ctx, ws)
		if err != nil {
			return nil, err
		}

		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) AddUserMember(ctx context.Context, workspaceID workspace.ID, users map[workspace.UserID]workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction().WithOwnableWorkspaces(workspaceID), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, workspaceID)
		if err != nil {
			return nil, err
		}

		if ws.IsPersonal() {
			return nil, workspace.ErrCannotModifyPersonalWorkspace
		}

		ul, err := i.userquery.FetchByID(ctx, maps.Keys(users))
		if err != nil {
			return nil, err
		}

		if i.enforceMemberCount != nil {
			if err := i.enforceMemberCount(ctx, ws, ul, operator); err != nil {
				return nil, err
			}
		}

		for _, m := range ul {
			if m == nil {
				continue
			}

			err = ws.Members().Join(m, users[m.ID()], *operator.User)
			if err != nil {
				return nil, err
			}
		}

		err = i.repos.Workspace.Save(ctx, ws)
		if err != nil {
			return nil, err
		}

		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) AddIntegrationMember(ctx context.Context, wId workspace.ID, iId workspace.IntegrationID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction().WithOwnableWorkspaces(wId), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, wId)
		if err != nil {
			return nil, err
		}

		err = ws.Members().AddIntegration(iId, role, *operator.User)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, ws)
		if err != nil {
			return nil, err
		}

		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) RemoveUserMember(ctx context.Context, id workspace.ID, u workspace.UserID, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		if ws.IsPersonal() {
			return nil, workspace.ErrCannotModifyPersonalWorkspace
		}

		isOwner := ws.Members().UserRole(*operator.User) == workspace.RoleOwner
		isSelfLeave := *operator.User == u
		if !isOwner && !isSelfLeave {
			return nil, accountinterfaces.ErrOperationDenied
		}

		if isSelfLeave && ws.Members().IsOnlyOwner(u) {
			return nil, accountinterfaces.ErrOwnerCannotLeaveTheWorkspace
		}

		err = ws.Members().Leave(u)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, ws)
		if err != nil {
			return nil, err
		}

		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) RemoveIntegration(ctx context.Context, wId workspace.ID, iId workspace.IntegrationID, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().WithOwnableWorkspaces(wId).Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, wId)
		if err != nil {
			return nil, err
		}

		err = ws.Members().DeleteIntegration(iId)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, ws)
		if err != nil {
			return nil, err
		}

		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) UpdateUserMember(ctx context.Context, id workspace.ID, u workspace.UserID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction().WithOwnableWorkspaces(id), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}

		if ws.IsPersonal() {
			return nil, workspace.ErrCannotModifyPersonalWorkspace
		}

		if u == *operator.User {
			return nil, accountinterfaces.ErrCannotChangeOwnerRole
		}

		err = ws.Members().UpdateUserRole(u, role)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, ws)
		if err != nil {
			return nil, err
		}

		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) UpdateIntegration(ctx context.Context, wId workspace.ID, iId workspace.IntegrationID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().WithOwnableWorkspaces(wId).Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, wId)
		if err != nil {
			return nil, err
		}

		err = ws.Members().UpdateIntegrationRole(iId, role)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, ws)
		if err != nil {
			return nil, err
		}

		i.applyDefaultPolicy(ws, operator)
		return ws, nil
	})
}

func (i *Workspace) Remove(ctx context.Context, id workspace.ID, operator *accountusecase.Operator) error {
	if operator.User == nil {
		return accountinterfaces.ErrInvalidOperator
	}

	return Run0(ctx, operator, i.repos, Usecase().Transaction().WithOwnableWorkspaces(id), func(ctx context.Context) error {
		ws, err := i.repos.Workspace.FindByID(ctx, id)
		if err != nil {
			return err
		}

		if ws.IsPersonal() {
			return workspace.ErrCannotModifyPersonalWorkspace
		}

		err = i.repos.Workspace.Remove(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (i *Workspace) applyDefaultPolicy(ws *workspace.Workspace, o *accountusecase.Operator) {
	if ws.Policy() == nil && o.DefaultPolicy != nil {
		ws.SetPolicy(o.DefaultPolicy)
	}
}

func filterWorkspaces(
	workspaces workspace.List,
	operator *accountusecase.Operator,
	err error,
	omitNil, applyDefaultPolicy bool,
) (workspace.List, error) {
	if err != nil {
		return nil, err
	}

	if operator == nil {
		if omitNil {
			return nil, nil
		}
		return make([]*workspace.Workspace, len(workspaces)), nil
	}

	for i, ws := range workspaces {
		if ws == nil || !operator.IsReadableWorkspace(ws.ID()) {
			workspaces[i] = nil
		}
	}

	if omitNil {
		workspaces = lo.Filter(workspaces, func(t *workspace.Workspace, _ int) bool {
			return t != nil
		})
	}

	if applyDefaultPolicy && operator.DefaultPolicy != nil {
		for _, ws := range workspaces {
			if ws == nil {
				continue
			}
			if ws.Policy() == nil {
				ws.SetPolicy(operator.DefaultPolicy)
			}
		}
	}

	return workspaces, nil
}
