package interactor

import (
	"context"
	"errors"
	"testing"

	"github.com/reearth/reearthx/asset/infrastructure/memory"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/item"
	"github.com/reearth/reearthx/asset/domain/model"
	"github.com/reearth/reearthx/asset/domain/project"
	"github.com/reearth/reearthx/asset/domain/request"
	"github.com/reearth/reearthx/asset/domain/schema"
	"github.com/reearth/reearthx/asset/domain/version"
	"github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/asset/usecase/interfaces"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestRequest_FindByID(t *testing.T) {
	pid := id.NewProjectID()
	item, _ := request.NewItemWithVersion(id.NewItemID(), version.New().OrRef())
	wid := accountdomain.NewWorkspaceID()

	req1 := request.New().
		NewID().
		Workspace(wid).
		Project(pid).
		CreatedBy(accountdomain.NewUserID()).
		Thread(id.NewThreadID().Ref()).
		Items(request.ItemList{item}).
		Title("foo").
		MustBuild()
	req2 := request.New().
		NewID().
		Workspace(wid).
		Project(pid).
		CreatedBy(accountdomain.NewUserID()).
		Thread(id.NewThreadID().Ref()).
		Items(request.ItemList{item}).
		Title("hoge").
		MustBuild()
	u := user.New().Name("aaa").NewID().Email("aaa@bbb.com").Workspace(wid).MustBuild()
	op := &usecase.Operator{
		AcOperator: &accountusecase.Operator{
			User: lo.ToPtr(u.ID()),
		},
	}

	tests := []struct {
		name  string
		seeds request.List
		args  struct {
			id       id.RequestID
			operator *usecase.Operator
		}
		want           *request.Request
		mockRequestErr bool
		wantErr        error
	}{
		{
			name:  "find 1 of 2",
			seeds: request.List{req1, req2},
			args: struct {
				id       id.RequestID
				operator *usecase.Operator
			}{
				id:       req1.ID(),
				operator: op,
			},
			want:    req1,
			wantErr: nil,
		},
		{
			name:  "find 1 of 0",
			seeds: request.List{},
			args: struct {
				id       id.RequestID
				operator *usecase.Operator
			}{
				id:       req1.ID(),
				operator: op,
			},
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()
			if tc.mockRequestErr {
				memory.SetRequestError(db.Request, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Request.Save(ctx, p)
				assert.NoError(t, err)
			}
			requestUC := NewRequest(db, nil)

			got, err := requestUC.FindByID(ctx, tc.args.id, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRequest_FindByIDs(t *testing.T) {
	pid := id.NewProjectID()
	item, _ := request.NewItemWithVersion(id.NewItemID(), version.New().OrRef())
	wid := accountdomain.NewWorkspaceID()

	req1 := request.New().
		NewID().
		Workspace(wid).
		Project(pid).
		CreatedBy(accountdomain.NewUserID()).
		Thread(id.NewThreadID().Ref()).
		Items(request.ItemList{item}).
		Title("foo").
		MustBuild()
	req2 := request.New().
		NewID().
		Workspace(wid).
		Project(pid).
		CreatedBy(accountdomain.NewUserID()).
		Thread(id.NewThreadID().Ref()).
		Items(request.ItemList{item}).
		Title("hoge").
		MustBuild()
	req3 := request.New().
		NewID().
		Workspace(wid).
		Project(pid).
		CreatedBy(accountdomain.NewUserID()).
		Thread(id.NewThreadID().Ref()).
		Items(request.ItemList{item}).
		Title("xxx").
		MustBuild()
	u := user.New().Name("aaa").NewID().Email("aaa@bbb.com").Workspace(wid).MustBuild()
	op := &usecase.Operator{
		AcOperator: &accountusecase.Operator{
			User: lo.ToPtr(u.ID()),
		},
	}

	tests := []struct {
		name  string
		seeds request.List
		args  struct {
			ids      id.RequestIDList
			operator *usecase.Operator
		}
		want           int
		mockRequestErr bool
		wantErr        error
	}{
		{
			name:  "find 2 of 3",
			seeds: request.List{req1, req2, req3},
			args: struct {
				ids      id.RequestIDList
				operator *usecase.Operator
			}{
				ids:      id.RequestIDList{req1.ID(), req2.ID()},
				operator: op,
			},
			want: 2,
		},
		{
			name:  "find 0 of 3",
			seeds: request.List{req1, req2, req3},
			args: struct {
				ids      id.RequestIDList
				operator *usecase.Operator
			}{
				ids:      id.RequestIDList{id.NewRequestID()},
				operator: op,
			},
			want: 0,
		},
		{
			name:           "mock error",
			mockRequestErr: true,
			wantErr:        errors.New("test"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()
			if tc.mockRequestErr {
				memory.SetRequestError(db.Request, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Request.Save(ctx, p)
				assert.NoError(t, err)
			}
			requestUC := NewRequest(db, nil)

			got, err := requestUC.FindByIDs(ctx, tc.args.ids, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, len(got))
		})
	}
}

func TestRequest_FindByProject(t *testing.T) {
	pid := id.NewProjectID()
	item, _ := request.NewItemWithVersion(id.NewItemID(), version.New().OrRef())
	wid := accountdomain.NewWorkspaceID()

	req1 := request.New().
		NewID().
		Workspace(accountdomain.NewWorkspaceID()).
		Project(pid).
		CreatedBy(accountdomain.NewUserID()).
		Thread(id.NewThreadID().Ref()).
		Items(request.ItemList{item}).
		Title("foo").
		MustBuild()
	req2 := request.New().
		NewID().
		Workspace(accountdomain.NewWorkspaceID()).
		Project(pid).
		CreatedBy(accountdomain.NewUserID()).
		Thread(id.NewThreadID().Ref()).
		Items(request.ItemList{item}).
		State(request.StateDraft).
		Title("hoge").
		MustBuild()
	u := user.New().Name("aaa").NewID().Email("aaa@bbb.com").Workspace(wid).MustBuild()
	op := &usecase.Operator{
		AcOperator: &accountusecase.Operator{
			User: lo.ToPtr(u.ID()),
		},
	}

	tests := []struct {
		name  string
		seeds request.List
		args  struct {
			pid      id.ProjectID
			filter   interfaces.RequestFilter
			operator *usecase.Operator
		}
		want           int
		mockRequestErr bool
		wantErr        error
	}{
		{
			name:  "must find 2",
			seeds: request.List{req1, req2},
			args: struct {
				pid      id.ProjectID
				filter   interfaces.RequestFilter
				operator *usecase.Operator
			}{
				pid:      pid,
				operator: op,
			},
			want: 2,
		},
		{
			name:  "must find 1",
			seeds: request.List{req1, req2},
			args: struct {
				pid      id.ProjectID
				filter   interfaces.RequestFilter
				operator *usecase.Operator
			}{
				pid: pid,
				filter: interfaces.RequestFilter{
					Keyword: lo.ToPtr("foo"),
				},
				operator: op,
			},
			want: 1,
		},
		{
			name:  "must find 1",
			seeds: request.List{req1, req2},
			args: struct {
				pid      id.ProjectID
				filter   interfaces.RequestFilter
				operator *usecase.Operator
			}{
				pid: pid,
				filter: interfaces.RequestFilter{
					State: []request.State{request.StateDraft},
				},
				operator: op,
			},
			want: 1,
		},
		{
			name:           "mock error",
			mockRequestErr: true,
			wantErr:        errors.New("test"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()
			if tc.mockRequestErr {
				memory.SetRequestError(db.Request, tc.wantErr)
			}
			for _, p := range tc.seeds {
				err := db.Request.Save(ctx, p)
				assert.NoError(t, err)
			}
			requestUC := NewRequest(db, nil)

			got, _, err := requestUC.FindByProject(
				ctx,
				tc.args.pid,
				tc.args.filter,
				nil,
				nil,
				tc.args.operator,
			)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, len(got))
		})
	}
}

func TestRequest_Approve(t *testing.T) {
	now := util.Now()
	defer util.MockNow(now)()

	wid := accountdomain.NewWorkspaceID()
	prj := project.New().NewID().MustBuild()
	s := schema.New().
		NewID().
		Workspace(accountdomain.NewWorkspaceID()).
		Project(prj.ID()).
		MustBuild()
	m := model.New().NewID().Schema(s.ID()).RandomKey().MustBuild()
	i := item.New().
		NewID().
		Schema(s.ID()).
		Model(m.ID()).
		Project(prj.ID()).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	u := user.New().Name("aaa").NewID().Email("aaa@bbb.com").Workspace(wid).MustBuild()

	tests := []struct {
		name           string
		mockRequestErr bool
		mockItemErr    bool
		wantErr        error
		setupData      bool
		operator       *usecase.Operator
	}{
		{
			name:      "successful approval",
			setupData: true,
			operator: &usecase.Operator{
				AcOperator: &accountusecase.Operator{
					User:             lo.ToPtr(u.ID()),
					OwningWorkspaces: accountdomain.WorkspaceIDList{wid},
				},
			},
			wantErr: nil,
		},
		{
			name:           "request not found",
			setupData:      false,
			mockRequestErr: true,
			operator: &usecase.Operator{
				AcOperator: &accountusecase.Operator{
					User:             lo.ToPtr(u.ID()),
					OwningWorkspaces: accountdomain.WorkspaceIDList{wid},
				},
			},
			wantErr: rerror.ErrNotFound,
		},
		{
			name:      "unauthorized user",
			setupData: true,
			operator: &usecase.Operator{
				AcOperator: &accountusecase.Operator{
					User:             lo.ToPtr(accountdomain.NewUserID()),
					OwningWorkspaces: accountdomain.WorkspaceIDList{},
				},
			},
			wantErr: interfaces.ErrInvalidOperator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := memory.New()

			if tt.mockRequestErr && !tt.setupData {
				memory.SetRequestError(db.Request, tt.wantErr)
			}

			if tt.setupData {
				err := db.Project.Save(ctx, prj)
				assert.NoError(t, err)
				err = db.Schema.Save(ctx, s)
				assert.NoError(t, err)
				err = db.Model.Save(ctx, m)
				assert.NoError(t, err)
				err = db.Item.Save(ctx, i)
				assert.NoError(t, err)

				vi, err := db.Item.FindByID(ctx, i.ID(), nil)
				assert.NoError(t, err)
				ri, _ := request.NewItem(i.ID(), lo.ToPtr(vi.Version().String()))
				req1 := request.New().
					NewID().
					Workspace(wid).
					Project(prj.ID()).
					Reviewers(accountdomain.UserIDList{u.ID()}).
					CreatedBy(accountdomain.NewUserID()).
					Thread(id.NewThreadID().Ref()).
					Items(request.ItemList{ri}).
					Title("foo").
					MustBuild()

				err = db.Request.Save(ctx, req1)
				assert.NoError(t, err)

				requestUC := NewRequest(db, nil)
				_, err = requestUC.Approve(ctx, req1.ID(), tt.operator)

				if tt.wantErr != nil {
					assert.ErrorIs(t, err, tt.wantErr)
					return
				}

				assert.NoError(t, err)

				itemUC := NewItem(db, nil)
				itm, err := itemUC.FindByID(ctx, i.ID(), tt.operator)
				assert.NoError(t, err)
				expected := version.MustBeValue(
					itm.Version(),
					nil,
					version.NewRefs(version.Public, version.Latest),
					now,
					i,
				)
				assert.Equal(t, expected, itm)
			} else {
				requestUC := NewRequest(db, nil)
				_, err := requestUC.Approve(ctx, id.NewRequestID(), tt.operator)
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
