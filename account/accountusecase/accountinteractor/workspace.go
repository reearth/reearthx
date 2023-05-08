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

type Workspace struct {
	repos *accountrepo.Container
}

func NewWorkspace(r *accountrepo.Container) accountinterfaces.Workspace {
	return &Workspace{
		repos: r,
	}
}

func (i *Workspace) Fetch(ctx context.Context, ids accountdomain.WorkspaceIDList, operator *accountusecase.Operator) ([]*workspace.Workspace, error) {
	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) ([]*workspace.Workspace, error) {
		res, err := i.repos.Workspace.FindByIDs(ctx, ids)
		res2, err := i.filterWorkspaces(res, operator, err)
		return res2, err
	})
}

func (i *Workspace) FindByUser(ctx context.Context, id accountdomain.UserID, operator *accountusecase.Operator) ([]*workspace.Workspace, error) {
	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) ([]*workspace.Workspace, error) {
		res, err := i.repos.Workspace.FindByUser(ctx, id)
		res2, err := i.filterWorkspaces(res, operator, err)
		return res2, err
	})
}

func (i *Workspace) FetchPolicy(ctx context.Context, ids []workspace.PolicyID, operator *accountusecase.Operator) ([]*workspace.Policy, error) {
	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) ([]*workspace.Policy, error) {
		res, err := i.repos.Policy.FindByIDs(ctx, ids)
		return res, err
	})
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

		return ws, nil
	})
}

func (i *Workspace) AddIntegrationMember(ctx context.Context, wId accountdomain.WorkspaceID, iId accountdomain.IntegrationID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}
	return Run1(ctx, operator, i.repos, Usecase().Transaction().WithOwnableWorkspaces(wId), func(ctx context.Context) (*workspace.Workspace, error) {
		workspace, err := i.repos.Workspace.FindByID(ctx, wId)
		if err != nil {
			return nil, err
		}

		err = workspace.Members().AddIntegration(iId, role, *operator.User)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, workspace)
		if err != nil {
			return nil, err
		}

		return workspace, nil
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

		return ws, nil
	})
}

func (i *Workspace) RemoveIntegration(ctx context.Context, wId accountdomain.WorkspaceID, iId accountdomain.IntegrationID, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}
	return Run1(ctx, operator, i.repos, Usecase().WithOwnableWorkspaces(wId).Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		workspace, err := i.repos.Workspace.FindByID(ctx, wId)
		if err != nil {
			return nil, err
		}

		err = workspace.Members().DeleteIntegration(iId)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, workspace)
		if err != nil {
			return nil, err
		}

		return workspace, nil
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

		return ws, nil
	})
}

func (i *Workspace) UpdateIntegration(ctx context.Context, wId accountdomain.WorkspaceID, iId accountdomain.IntegrationID, role workspace.Role, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}
	return Run1(ctx, operator, i.repos, Usecase().WithOwnableWorkspaces(wId).Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
		workspace, err := i.repos.Workspace.FindByID(ctx, wId)
		if err != nil {
			return nil, err
		}

		err = workspace.Members().UpdateIntegrationRole(iId, role)
		if err != nil {
			return nil, err
		}

		err = i.repos.Workspace.Save(ctx, workspace)
		if err != nil {
			return nil, err
		}

		return workspace, nil
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

func (i *Workspace) filterWorkspaces(workspaces []*workspace.Workspace, operator *accountusecase.Operator, err error) ([]*workspace.Workspace, error) {
	if err != nil {
		return nil, err
	}
	if operator == nil {
		return make([]*workspace.Workspace, len(workspaces)), nil
	}
	for i, t := range workspaces {
		if t == nil || !operator.IsReadableWorkspace(t.ID()) {
			workspaces[i] = nil
		}
	}
	return workspaces, nil
}
