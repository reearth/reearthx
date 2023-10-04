package accountinteractor

import (
	"context"
	"strings"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"golang.org/x/exp/maps"
)

type WorkspaceMemberCountEnforcer func(context.Context, *workspace.Workspace, user.List, *accountusecase.Operator) error

type Workspace struct {
	repos              *accountrepo.Container
	enforceMemberCount WorkspaceMemberCountEnforcer
}

func NewWorkspace(r *accountrepo.Container, enforceMemberCount WorkspaceMemberCountEnforcer) accountinterfaces.Workspace {
	return &Workspace{
		repos:              r,
		enforceMemberCount: enforceMemberCount,
	}
}

func (i *Workspace) Fetch(ctx context.Context, ids accountdomain.WorkspaceIDList, operator *accountusecase.Operator) ([]*workspace.Workspace, error) {
	res, err := i.repos.Workspace.FindByIDs(ctx, ids)
	res2, err := accountinterfaces.FilterWorkspaces(res, operator, err, false)
	return res2, err
}

func (i *Workspace) FindByUser(ctx context.Context, id accountdomain.UserID, operator *accountusecase.Operator) ([]*workspace.Workspace, error) {
	res, err := i.repos.Workspace.FindByUser(ctx, id)
	res2, err := accountinterfaces.FilterWorkspaces(res, operator, err, true)
	return res2, err
}

func (i *Workspace) Create(ctx context.Context, name string, firstUser accountdomain.UserID, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		if len(strings.TrimSpace(name)) == 0 {
			return nil, user.ErrInvalidName
		}

		ws, err := workspace.New().
			NewID().
			Name(name).
			Build()
		if err != nil {
			return nil, err
		}

		if err := ws.Members().Join(firstUser, workspace.RoleOwner, *operator.User); err != nil {
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

func (i *Workspace) Update(ctx context.Context, id accountdomain.WorkspaceID, name string, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
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

func (i *Workspace) AddUserMember(ctx context.Context, workspaceID accountdomain.WorkspaceID, users map[accountdomain.UserID]workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
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

		ul, err := i.repos.User.FindByIDs(ctx, maps.Keys(users))
		if err != nil {
			return nil, err
		}

		if i.enforceMemberCount != nil {
			if err := i.enforceMemberCount(ctx, ws, ul, operator); err != nil {
				return nil, err
			}
		}

		for _, m := range ul {
			err = ws.Members().Join(m.ID(), users[m.ID()], *operator.User)
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

func (i *Workspace) AddIntegrationMember(ctx context.Context, wId accountdomain.WorkspaceID, iId accountdomain.IntegrationID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
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

func (i *Workspace) RemoveUserMember(ctx context.Context, id accountdomain.WorkspaceID, u accountdomain.UserID, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
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

func (i *Workspace) RemoveIntegration(ctx context.Context, wId accountdomain.WorkspaceID, iId accountdomain.IntegrationID, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
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

func (i *Workspace) UpdateUserMember(ctx context.Context, id accountdomain.WorkspaceID, u accountdomain.UserID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
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

func (i *Workspace) UpdateIntegration(ctx context.Context, wId accountdomain.WorkspaceID, iId accountdomain.IntegrationID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
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

func (i *Workspace) Remove(ctx context.Context, id accountdomain.WorkspaceID, operator *accountusecase.Operator) error {
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
