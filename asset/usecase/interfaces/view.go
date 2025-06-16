package interfaces

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/item/view"
	"github.com/reearth/reearthx/asset/usecase"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

type CreateViewParam struct {
	Filter  *view.Condition
	Sort    *view.Sort
	Columns *view.ColumnList
	Name    string
	Project view.ProjectID
	Model   view.ModelID
}

type UpdateViewParam struct {
	Name    *string
	Filter  *view.Condition
	Sort    *view.Sort
	Columns *view.ColumnList
	ID      view.ID
}

var (
	ErrLastView                  = rerror.NewE(i18n.T("model should have at least one view"))
	ErrViewsAreNotInTheSameModel = rerror.NewE(i18n.T("views are not in the same model"))
	ErrViewsLengthMismatch       = rerror.NewE(i18n.T("views length mismatch"))
)

type View interface {
	FindByIDs(context.Context, view.IDList, *usecase.Operator) (view.List, error)
	FindByModel(context.Context, view.ModelID, *usecase.Operator) (view.List, error)
	Create(context.Context, CreateViewParam, *usecase.Operator) (*view.View, error)
	Update(context.Context, view.ID, UpdateViewParam, *usecase.Operator) (*view.View, error)
	UpdateOrder(context.Context, view.IDList, *usecase.Operator) (view.List, error)
	Delete(context.Context, view.ID, *usecase.Operator) error
}
