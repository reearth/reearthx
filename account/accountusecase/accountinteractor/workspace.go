package accountinteractor

import (
	"context"
	"fmt"
	"strings"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/permittable"
	"github.com/reearth/reearthx/account/accountdomain/role"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/log"
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

	return Run1(ctx, operator, i.repos, Usecase().Transaction(), func(ctx context.Context) (*workspace.Workspace, error) {
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

		if err := i.repos.Workspace.Create(ctx, ws); err != nil {
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

	ul, err := i.userquery.FetchByID(ctx, maps.Keys(users))
	if err != nil {
		return nil, err
	}

	return Run1(ctx, operator, i.repos, Usecase().Transaction().WithOwnableWorkspaces(workspaceID), func(ctx context.Context) (*workspace.Workspace, error) {
		ws, err := i.repos.Workspace.FindByID(ctx, workspaceID)
		if err != nil {
			return nil, err
		}

		if ws.IsPersonal() {
			return nil, workspace.ErrCannotModifyPersonalWorkspace
		}

		if i.enforceMemberCount != nil {
			if err := i.enforceMemberCount(ctx, ws, ul, operator); err != nil {
				return nil, err
			}
		}

		// TODO: Delete this once the permission check migration is complete.
		maintainerRole, err := i.getMaintainerRole(ctx)
		if err != nil {
			return nil, err
		}

		for _, m := range ul {
			if m == nil {
				continue
			}

			// TODO: Delete this once the permission check migration is complete.
			if err := i.ensureUserHasMaintainerRole(ctx, m.ID(), maintainerRole.ID()); err != nil {
				return nil, err
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

func (i *Workspace) RemoveUserMember(ctx context.Context, id workspace.ID, u workspace.UserID, operator *accountusecase.Operator) (*workspace.Workspace, error) {
	return i.RemoveMultipleUserMembers(ctx, id, workspace.UserIDList{u}, operator)
}

func (i *Workspace) RemoveMultipleUserMembers(ctx context.Context, id workspace.ID, userIds workspace.UserIDList, operator *accountusecase.Operator) (_ *workspace.Workspace, err error) {
	if operator.User == nil {
		return nil, accountinterfaces.ErrInvalidOperator
	}

	if userIds.Len() == 0 {
		return nil, workspace.ErrNoSpecifiedUsers
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

		for _, uId := range userIds {
			isSelfLeave := *operator.User == uId

			if !isOwner && !isSelfLeave {
				return nil, accountinterfaces.ErrOperationDenied
			}
			if isSelfLeave && ws.Members().IsOnlyOwner(uId) {
				return nil, accountinterfaces.ErrOwnerCannotLeaveTheWorkspace
			}

			err := ws.Members().Leave(uId)
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

// TODO: Delete this once the permission check migration is complete.
func (i *Workspace) getMaintainerRole(ctx context.Context) (*role.Role, error) {
	// check and create maintainer role
	if i.repos.Role == nil {
		return nil, fmt.Errorf("repos not found Role %v", i.repos.Role)
	}
	roles, err := i.repos.Role.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	var maintainerRole *role.Role
	for _, r := range roles {
		if r.Name() == "maintainer" {
			maintainerRole = r
			log.Info("Found maintainer role")
			break
		}
	}

	if maintainerRole == nil {
		r, err := role.New().
			NewID().
			Name("maintainer").
			Build()
		if err != nil {
			return nil, fmt.Errorf("failed to create maintainer role domain: %w", err)
		}

		err = i.repos.Role.Save(ctx, *r)
		if err != nil {
			return nil, fmt.Errorf("failed to save maintainer role: %w", err)
		}

		maintainerRole = r
		log.Info("Created maintainer role")
	}

	return maintainerRole, nil
}

// TODO: Delete this once the permission check migration is complete.
func (i *Workspace) ensureUserHasMaintainerRole(ctx context.Context, userID user.ID, maintainerRoleID accountdomain.RoleID) error {
	var p *permittable.Permittable
	var err error

	p, err = i.repos.Permittable.FindByUserID(ctx, userID)
	if err != nil && err != rerror.ErrNotFound {
		return err
	}

	if hasRole(p, maintainerRoleID) {
		return nil
	}

	if p == nil {
		p, err = permittable.New().
			NewID().
			UserID(userID).
			RoleIDs([]accountdomain.RoleID{maintainerRoleID}).
			Build()
		if err != nil {
			return err
		}
	} else {
		p.EditRoleIDs(append(p.RoleIDs(), maintainerRoleID))
	}

	err = i.repos.Permittable.Save(ctx, *p)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Delete this once the permission check migration is complete.
func hasRole(p *permittable.Permittable, roleID role.ID) bool {
	if p == nil {
		return false
	}
	for _, r := range p.RoleIDs() {
		if r == roleID {
			return true
		}
	}
	return false
}
