package accountusecase

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/util"
)

type Operator struct {
	User                   *accountdomain.UserID
	ReadableWorkspaces     accountdomain.WorkspaceIDList
	WritableWorkspaces     accountdomain.WorkspaceIDList
	OwningWorkspaces       accountdomain.WorkspaceIDList
	MaintainableWorkspaces accountdomain.WorkspaceIDList
	DefaultPolicy          *workspace.PolicyID
}

func (o *Operator) Workspaces(r workspace.Role) accountdomain.WorkspaceIDList {
	if o == nil {
		return nil
	}
	if r == workspace.RoleReader {
		return o.ReadableWorkspaces
	}
	if r == workspace.RoleWriter {
		return o.WritableWorkspaces
	}
	if r == workspace.RoleOwner {
		return o.OwningWorkspaces
	}
	return nil
}

func (o *Operator) AllReadableWorkspaces() accountdomain.WorkspaceIDList {
	return append(o.ReadableWorkspaces, o.AllWritableWorkspaces()...)
}

func (o *Operator) AllWritableWorkspaces() accountdomain.WorkspaceIDList {
	return append(o.WritableWorkspaces, o.AllMaintainingWorkspaces()...)
}

func (o *Operator) AllMaintainingWorkspaces() accountdomain.WorkspaceIDList {
	return append(o.MaintainableWorkspaces, o.AllOwningWorkspaces()...)
}

func (o *Operator) AllOwningWorkspaces() accountdomain.WorkspaceIDList {
	return o.OwningWorkspaces
}

func (o *Operator) IsReadableWorkspace(ws ...accountdomain.WorkspaceID) bool {
	return o.AllReadableWorkspaces().Intersect(ws).Len() > 0
}

func (o *Operator) IsWritableWorkspace(ws ...accountdomain.WorkspaceID) bool {
	return o.AllWritableWorkspaces().Intersect(ws).Len() > 0
}

func (o *Operator) IsMaintainingWorkspace(workspace ...accountdomain.WorkspaceID) bool {
	return o.AllMaintainingWorkspaces().Intersect(workspace).Len() > 0
}

func (o *Operator) IsOwningWorkspace(ws ...accountdomain.WorkspaceID) bool {
	return o.AllOwningWorkspaces().Intersect(ws).Len() > 0
}

func (o *Operator) AddNewWorkspace(ws accountdomain.WorkspaceID) {
	o.OwningWorkspaces = append(o.OwningWorkspaces, ws)
}

func (o *Operator) Policy(p *workspace.PolicyID) *workspace.PolicyID {
	if p == nil && o.DefaultPolicy != nil && *o.DefaultPolicy != "" {
		return util.CloneRef(o.DefaultPolicy)
	}
	if p != nil && *p == "" {
		return nil
	}
	return p
}
