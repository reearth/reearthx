package accountinteractor

import (
	"context"

	"github.com/reearth/reearthx/account/accountusecase"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
)

type uc struct {
	tx                     bool
	readableWorkspaces     accountdomain.WorkspaceIDList
	writableWorkspaces     accountdomain.WorkspaceIDList
	maintainableWorkspaces accountdomain.WorkspaceIDList
	ownableWorkspaces      accountdomain.WorkspaceIDList
}

func Usecase() *uc {
	return &uc{}
}

func (u *uc) WithReadableWorkspaces(ids ...accountdomain.WorkspaceID) *uc {
	u.readableWorkspaces = accountdomain.WorkspaceIDList(ids).Clone()
	return u
}

func (u *uc) WithWritableWorkspaces(ids ...accountdomain.WorkspaceID) *uc {
	u.writableWorkspaces = accountdomain.WorkspaceIDList(ids).Clone()
	return u
}

func (u *uc) WithMaintainableWorkspaces(ids ...accountdomain.WorkspaceID) *uc {
	u.maintainableWorkspaces = accountdomain.WorkspaceIDList(ids).Clone()
	return u
}

func (u *uc) WithOwnableWorkspaces(ids ...accountdomain.WorkspaceID) *uc {
	u.ownableWorkspaces = accountdomain.WorkspaceIDList(ids).Clone()
	return u
}

func (u *uc) Transaction() *uc {
	u.tx = true
	return u
}

func Run0(ctx context.Context, op *accountusecase.Operator, r *accountrepo.Container, e *uc, f func() error) (err error) {
	_, _, _, err = Run3(
		ctx, op, r, e,
		func() (_, _, _ any, err error) {
			err = f()
			return
		})
	return
}

func Run1[A any](ctx context.Context, op *accountusecase.Operator, r *accountrepo.Container, e *uc, f func() (A, error)) (a A, err error) {
	a, _, _, err = Run3(
		ctx, op, r, e,
		func() (a A, _, _ any, err error) {
			a, err = f()
			return
		})
	return
}

func Run2[A, B any](ctx context.Context, op *accountusecase.Operator, r *accountrepo.Container, e *uc, f func() (A, B, error)) (a A, b B, err error) {
	a, b, _, err = Run3(
		ctx, op, r, e,
		func() (a A, b B, _ any, err error) {
			a, b, err = f()
			return
		})
	return
}

func Run3[A, B, C any](ctx context.Context, op *accountusecase.Operator, r *accountrepo.Container, e *uc, f func() (A, B, C, error)) (_ A, _ B, _ C, err error) {
	if err = e.checkPermission(op); err != nil {
		return
	}
	if e.tx && r.Transaction != nil {
		tx, err2 := r.Transaction.Begin()
		if err2 != nil {
			err = err2
			return
		}
		defer func() {
			if err == nil {
				tx.Commit()
			}
			if err2 := tx.End(ctx); err == nil && err2 != nil {
				err = err2
			}
		}()
	}

	return f()
}

func (u *uc) checkPermission(op *accountusecase.Operator) error {
	ok := true
	if u.readableWorkspaces != nil {
		ok = op.IsReadableWorkspace(u.readableWorkspaces...)
	}
	if ok && u.writableWorkspaces != nil {
		ok = op.IsWritableWorkspace(u.writableWorkspaces...)
	}
	if ok && u.maintainableWorkspaces != nil {
		ok = op.IsMaintainingWorkspace(u.maintainableWorkspaces...)
	}
	if ok && u.ownableWorkspaces != nil {
		ok = op.IsOwningWorkspace(u.ownableWorkspaces...)
	}
	if !ok {
		return accountinterfaces.ErrOperationDenied
	}
	return nil
}
