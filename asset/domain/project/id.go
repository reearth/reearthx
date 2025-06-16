package project

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/samber/lo"
)

type (
	ID          = id.ProjectID
	WorkspaceID = id.WorkspaceID
)

type IDList = id.ProjectIDList

var (
	NewID          = id.NewProjectID
	NewWorkspaceID = accountdomain.NewWorkspaceID
)

var (
	MustID          = id.MustProjectID
	MustWorkspaceID = id.MustWorkspaceID
)

var (
	IDFrom          = id.ProjectIDFrom
	WorkspaceIDFrom = id.WorkspaceIDFrom
)

var (
	IDFromRef          = id.ProjectIDFromRef
	WorkspaceIDFromRef = id.WorkspaceIDFromRef
)

var ErrInvalidID = id.ErrInvalidID

type IDOrAlias string

func (i IDOrAlias) ID() *ID {
	return IDFromRef(lo.ToPtr(string(i)))
}

func (i IDOrAlias) Alias() *string {
	if string(i) != "" && i.ID() == nil {
		return lo.ToPtr(string(i))
	}
	return nil
}
