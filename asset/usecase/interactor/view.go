package interactor

import (
	"context"
	"time"

	"github.com/reearth/reearthx/asset/domain/item/view"
	"github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/asset/usecase/gateway"
	"github.com/reearth/reearthx/asset/usecase/interfaces"
	"github.com/reearth/reearthx/asset/usecase/repo"

	"github.com/reearth/reearthx/rerror"
)

type View struct {
	repos    *repo.Container
	gateways *gateway.Container
}

func NewView(r *repo.Container, g *gateway.Container) interfaces.View {
	return &View{
		repos:    r,
		gateways: g,
	}
}

func (i View) FindByID(ctx context.Context, id view.ID, _ *usecase.Operator) (*view.View, error) {
	return i.repos.View.FindByID(ctx, id)
}

func (i View) FindByIDs(
	ctx context.Context,
	ids view.IDList,
	_ *usecase.Operator,
) (view.List, error) {
	return i.repos.View.FindByIDs(ctx, ids)
}

func (i View) FindByModel(
	ctx context.Context,
	mID view.ModelID,
	_ *usecase.Operator,
) (view.List, error) {
	v, err := i.repos.View.FindByModel(ctx, mID)
	if err != nil {
		return nil, err
	}
	return v.Ordered(), nil
}

func (i View) Create(
	ctx context.Context,
	param interfaces.CreateViewParam,
	op *usecase.Operator,
) (*view.View, error) {
	if op.AcOperator.User == nil {
		return nil, interfaces.ErrInvalidOperator
	}
	return Run1(ctx, op, i.repos, Usecase().Transaction(),
		func(ctx context.Context) (_ *view.View, err error) {
			if !op.IsMaintainingProject(param.Project) {
				return nil, interfaces.ErrOperationDenied
			}

			m, err := i.repos.Model.FindByID(ctx, param.Model)
			if err != nil {
				return nil, err
			}

			if m == nil || m.Project() != param.Project {
				return nil, rerror.ErrNotFound
			}

			vb := view.
				New().
				NewID().
				Project(param.Project).
				Model(param.Model).
				Schema(m.Schema()).
				Name(param.Name).
				Sort(param.Sort).
				Filter(param.Filter).
				Columns(param.Columns).
				User(*op.Operator().User())

			views, err := i.repos.View.FindByModel(ctx, param.Model)
			if err != nil {
				return nil, err
			}
			if len(views) > 0 {
				vb = vb.Order(len(views))
			}

			v, err := vb.Build()
			if err != nil {
				return nil, err
			}

			err = i.repos.View.Save(ctx, v)
			if err != nil {
				return nil, err
			}
			return v, nil
		})
}

func (i View) Update(
	ctx context.Context,
	id view.ID,
	param interfaces.UpdateViewParam,
	op *usecase.Operator,
) (*view.View, error) {
	return Run1(ctx, op, i.repos, Usecase().Transaction(),
		func(ctx context.Context) (_ *view.View, err error) {
			v, err := i.repos.View.FindByID(ctx, id)
			if err != nil {
				return nil, err
			}

			if !op.IsMaintainingProject(v.Project()) {
				return nil, interfaces.ErrOperationDenied
			}

			if param.Name != nil {
				v.SetName(*param.Name)
			}
			v.SetFilter(param.Filter)
			v.SetSort(param.Sort)
			v.SetColumns(param.Columns)
			v.SetUpdatedAt(time.Now())

			if err := i.repos.View.Save(ctx, v); err != nil {
				return nil, err
			}
			return v, nil
		})
}

func (i View) UpdateOrder(
	ctx context.Context,
	ids view.IDList,
	operator *usecase.Operator,
) (view.List, error) {
	return Run1(ctx, operator, i.repos, Usecase().Transaction(),
		func(ctx context.Context) (_ view.List, err error) {
			if len(ids) == 0 {
				return nil, nil
			}
			v, err := i.repos.View.FindByIDs(ctx, ids)
			if err != nil {
				return nil, err
			}
			if !v.AreViewsInTheSameModel() {
				return nil, interfaces.ErrViewsAreNotInTheSameModel
			}
			if !operator.IsMaintainingProject(v[0].Project()) {
				return nil, interfaces.ErrOperationDenied
			}

			views, err := i.repos.View.FindByModel(ctx, v[0].Model())
			if err != nil {
				return nil, err
			}
			if len(views) != len(ids) {
				return nil, interfaces.ErrViewsLengthMismatch
			}

			ordered := views.OrderByIDs(ids)
			if err := i.repos.View.SaveAll(ctx, ordered); err != nil {
				return nil, err
			}
			return ordered, nil
		})
}

func (i View) Delete(ctx context.Context, id view.ID, op *usecase.Operator) error {
	return Run0(ctx, op, i.repos, Usecase().Transaction(),
		func(ctx context.Context) error {
			m, err := i.repos.View.FindByID(ctx, id)
			if err != nil {
				return err
			}
			if !op.IsMaintainingProject(m.Project()) {
				return interfaces.ErrOperationDenied
			}

			views, err := i.repos.View.FindByModel(ctx, m.Model())
			if err != nil {
				return err
			}
			if len(views) <= 1 {
				return interfaces.ErrLastView
			}

			if err := i.repos.View.Remove(ctx, id); err != nil {
				return err
			}
			return nil
		})
}
