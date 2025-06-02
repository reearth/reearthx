package repo

import (
	"context"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/item/view"
)

type View interface {
	Filtered(ProjectFilter) View
	FindByIDs(context.Context, id.ViewIDList) (view.List, error)
	FindByModel(context.Context, id.ModelID) (view.List, error)
	FindByID(context.Context, id.ViewID) (*view.View, error)
	Save(context.Context, *view.View) error
	SaveAll(context.Context, view.List) error
	Remove(context.Context, id.ViewID) error
}
