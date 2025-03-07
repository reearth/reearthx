package accountrepo

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
)

type Container struct {
	User        User
	Workspace   Workspace
	Role        Role        // TODO: Delete this once the permission check migration is complete.
	Permittable Permittable // TODO: Delete this once the permission check migration is complete.
	Transaction usecasex.Transaction
	Users       []User
}

var (
	ErrOperationDenied = rerror.NewE(i18n.T("operation denied"))
)

func (c *Container) Filtered(workspace WorkspaceFilter) *Container {
	if c == nil {
		return c
	}
	return &Container{
		Workspace: c.Workspace.Filtered(workspace),
		User:      c.User,
		Users:     c.Users,
	}
}

type WorkspaceFilter struct {
	Readable accountdomain.WorkspaceIDList
	Writable accountdomain.WorkspaceIDList
}

func WorkspaceFilterFromOperator(o *accountusecase.Operator) WorkspaceFilter {
	return WorkspaceFilter{
		Readable: o.AllReadableWorkspaces(),
		Writable: o.AllWritableWorkspaces(),
	}
}

func (f WorkspaceFilter) Clone() WorkspaceFilter {
	return WorkspaceFilter{
		Readable: f.Readable.Clone(),
		Writable: f.Writable.Clone(),
	}
}

func (f WorkspaceFilter) Merge(g WorkspaceFilter) WorkspaceFilter {
	var r, w accountdomain.WorkspaceIDList
	if f.Readable != nil || g.Readable != nil {
		if f.Readable == nil {
			r = g.Readable.Clone()
		} else {
			r = append(f.Readable, g.Readable...)
		}
	}
	if f.Writable != nil || g.Writable != nil {
		if f.Writable == nil {
			w = g.Writable.Clone()
		} else {
			w = append(f.Writable, g.Writable...)
		}
	}
	return WorkspaceFilter{
		Readable: r,
		Writable: w,
	}
}

func (f WorkspaceFilter) CanRead(id accountdomain.WorkspaceID) bool {
	return f.Readable == nil || f.Readable.Has(id) || f.CanWrite(id)
}

func (f WorkspaceFilter) CanWrite(id accountdomain.WorkspaceID) bool {
	return f.Writable == nil || f.Writable.Has(id)
}

func (f WorkspaceFilter) Filter(q any) any {
	if f.Readable == nil {
		return q
	}

	return bson.M{
		"$and": bson.A{
			bson.M{"id": bson.M{"$in": f.Readable.Strings()}},
			q,
		},
	}
}
