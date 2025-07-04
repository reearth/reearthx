package interactor

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/asset/infrastructure/memory"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/item"
	"github.com/reearth/reearthx/asset/domain/operator"
	"github.com/reearth/reearthx/asset/domain/thread"
	"github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/asset/usecase/interfaces"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestThread_FindByID(t *testing.T) {
	id1 := id.NewThreadID()
	wid1 := accountdomain.NewWorkspaceID()
	comments := []*thread.Comment{}
	th1 := thread.New().ID(id1).Workspace(wid1).Comments(comments).MustBuild()

	op := &usecase.Operator{}

	type args struct {
		id       id.ThreadID
		operator *usecase.Operator
	}

	tests := []struct {
		name    string
		seeds   []*thread.Thread
		args    args
		want    *thread.Thread
		wantErr error
	}{
		{
			name:  "Not found in empty db",
			seeds: []*thread.Thread{},
			args: args{
				id:       id.NewThreadID(),
				operator: op,
			},
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name: "Not found",
			seeds: []*thread.Thread{
				thread.New().ID(id1).Workspace(wid1).Comments(comments).MustBuild(),
			},
			args: args{
				id:       id.NewThreadID(),
				operator: op,
			},
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name: "Found 1",
			seeds: []*thread.Thread{
				th1,
			},
			args: args{
				id:       id1,
				operator: op,
			},
			want:    th1,
			wantErr: nil,
		},
		{
			name: "Found 2",
			seeds: []*thread.Thread{
				th1,
				thread.New().ID(id1).Workspace(wid1).Comments(comments).MustBuild(),
				thread.New().ID(id1).Workspace(wid1).Comments(comments).MustBuild(),
			},
			args: args{
				id:       id1,
				operator: op,
			},
			want:    th1,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()

			for _, a := range tc.seeds {
				err := db.Thread.Save(ctx, a.Clone())
				assert.NoError(t, err)
			}
			threadUC := NewThread(db, nil)

			got, err := threadUC.FindByID(ctx, tc.args.id, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestThread_FindByIDs(t *testing.T) {
	id1 := id.NewThreadID()
	wid1 := accountdomain.NewWorkspaceID()
	comments1 := []*thread.Comment{}
	th1 := thread.New().ID(id1).Workspace(wid1).Comments(comments1).MustBuild()

	id2 := id.NewThreadID()
	wid2 := accountdomain.NewWorkspaceID()
	comments2 := []*thread.Comment{}
	th2 := thread.New().ID(id2).Workspace(wid2).Comments(comments2).MustBuild()

	tests := []struct {
		name    string
		seeds   thread.List
		arg     id.ThreadIDList
		want    thread.List
		wantErr error
	}{
		{
			name:    "0 count in empty db",
			seeds:   thread.List{},
			arg:     []id.ThreadID{},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "0 count with thread for another workspaces",
			seeds: thread.List{
				thread.New().
					NewID().
					Workspace(accountdomain.NewWorkspaceID()).
					Comments([]*thread.Comment{}).
					MustBuild(),
			},
			arg:     []id.ThreadID{},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "1 count with single thread",
			seeds: thread.List{
				th1,
			},
			arg:     []id.ThreadID{id1},
			want:    thread.List{th1},
			wantErr: nil,
		},
		{
			name: "1 count with multi threads",
			seeds: thread.List{
				th1,
				thread.New().
					NewID().
					Workspace(accountdomain.NewWorkspaceID()).
					Comments([]*thread.Comment{}).
					MustBuild(),
				thread.New().
					NewID().
					Workspace(accountdomain.NewWorkspaceID()).
					Comments([]*thread.Comment{}).
					MustBuild(),
			},
			arg:     []id.ThreadID{id1},
			want:    thread.List{th1},
			wantErr: nil,
		},
		{
			name: "2 count with multi threads",
			seeds: thread.List{
				th1,
				th2,
				thread.New().
					NewID().
					Workspace(accountdomain.NewWorkspaceID()).
					Comments([]*thread.Comment{}).
					MustBuild(),
				thread.New().
					NewID().
					Workspace(accountdomain.NewWorkspaceID()).
					Comments([]*thread.Comment{}).
					MustBuild(),
			},
			arg:     []id.ThreadID{id1, id2},
			want:    thread.List{th1, th2},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()

			for _, a := range tc.seeds {
				err := db.Thread.Save(ctx, a.Clone())
				assert.NoError(t, err)
			}
			threadUC := NewThread(db, nil)

			got, err := threadUC.FindByIDs(
				ctx,
				tc.arg,
				&usecase.Operator{AcOperator: &accountusecase.Operator{}},
			)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestThreadRepo_CreateThreadWithComment(t *testing.T) {
	wid := accountdomain.NewWorkspaceID()
	wid2 := accountdomain.WorkspaceID{}
	pid := id.NewProjectID()
	uid := accountdomain.NewUserID()
	i := item.New().NewID().Schema(id.NewSchemaID()).Model(id.NewModelID()).Project(pid).MustBuild()
	rt := interfaces.ResourceTypeItem
	content := "content"
	op := &usecase.Operator{
		AcOperator: &accountusecase.Operator{
			User:               &uid,
			ReadableWorkspaces: nil,
			WritableWorkspaces: nil,
			OwningWorkspaces:   []accountdomain.WorkspaceID{wid},
		},
	}

	tests := []struct {
		name     string
		arg      interfaces.CreateThreadWithCommentInput
		operator *usecase.Operator
		wantErr  error
	}{
		{
			name: "Save succeed",
			arg: interfaces.CreateThreadWithCommentInput{
				WorkspaceID:  wid,
				ResourceID:   i.ID().String(),
				ResourceType: rt,
				Content:      content,
			},
			operator: op,
			wantErr:  nil,
		},
		{
			name: "Save error: invalid workspace id",
			arg: interfaces.CreateThreadWithCommentInput{
				WorkspaceID:  wid2,
				ResourceID:   i.ID().String(),
				ResourceType: rt,
				Content:      content,
			},
			operator: &usecase.Operator{
				AcOperator: &accountusecase.Operator{
					User:               &uid,
					ReadableWorkspaces: nil,
					WritableWorkspaces: nil,
					OwningWorkspaces:   []accountdomain.WorkspaceID{wid2},
				},
			},
			wantErr: thread.ErrNoWorkspaceID,
		},
		{
			name: "operator error",
			arg:  interfaces.CreateThreadWithCommentInput{},
			operator: &usecase.Operator{
				AcOperator: &accountusecase.Operator{},
			},
			wantErr: interfaces.ErrOperationDenied,
		},
		{
			name: "operator succeed",
			arg: interfaces.CreateThreadWithCommentInput{
				WorkspaceID:  wid,
				ResourceID:   i.ID().String(),
				ResourceType: rt,
				Content:      content,
			},
			operator: op,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db := memory.New()

			itemUC := NewItem(db, nil)
			err := itemUC.repos.Item.Save(ctx, i)
			assert.NoError(t, err)

			threadUC := NewThread(db, nil)
			th, _, err := threadUC.CreateThreadWithComment(
				ctx,
				interfaces.CreateThreadWithCommentInput{
					WorkspaceID:  tc.arg.WorkspaceID,
					ResourceID:   tc.arg.ResourceID,
					ResourceType: tc.arg.ResourceType,
					Content:      tc.arg.Content,
				},
				tc.operator,
			)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			} else {
				assert.NoError(t, err)
			}

			res, err := threadUC.FindByID(ctx, th.ID(), tc.operator)
			assert.NoError(t, err)
			assert.Equal(t, res, th)
		})
	}
}

func TestThread_AddComment(t *testing.T) {
	c1 := thread.NewComment(
		thread.NewCommentID(),
		operator.OperatorFromUser(accountdomain.NewUserID()),
		"aaa",
	)
	wid := accountdomain.NewWorkspaceID()
	th1 := thread.New().NewID().Workspace(wid).Comments([]*thread.Comment{}).MustBuild()
	uid := accountdomain.NewUserID()
	op := &usecase.Operator{
		AcOperator: &accountusecase.Operator{
			User:               &uid,
			ReadableWorkspaces: nil,
			WritableWorkspaces: nil,
			OwningWorkspaces:   []accountdomain.WorkspaceID{wid},
		},
	}

	type args struct {
		content  string
		operator *usecase.Operator
	}

	tests := []struct {
		name      string
		seed      *thread.Thread
		args      args
		wantErr   error
		mockError bool
	}{
		{
			name: "workspaces invalid operator",
			seed: th1,
			args: args{
				content: c1.Content(),
				operator: &usecase.Operator{
					AcOperator: &accountusecase.Operator{},
				},
			},
			wantErr: interfaces.ErrInvalidOperator,
		},
		{
			name: "workspaces operation success",
			seed: th1,
			args: args{
				content:  c1.Content(),
				operator: op,
			},
			wantErr: nil,
		},
		{
			name: "add comment success",
			seed: th1,
			args: args{
				content:  c1.Content(),
				operator: op,
			},
			wantErr: nil,
		},
		{
			name: "add comment fail",
			seed: th1,
			args: args{
				content:  c1.Content(),
				operator: op,
			},
			wantErr:   rerror.ErrNotFound,
			mockError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()

			thread := tc.seed.Clone()
			err := db.Thread.Save(ctx, thread)
			assert.NoError(t, err)

			threadUC := NewThread(db, nil)
			if tc.mockError && tc.wantErr != nil {
				thid := id.NewThreadID()
				_, _, err := threadUC.AddComment(ctx, thid, tc.args.content, tc.args.operator)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			th, c, err := threadUC.AddComment(ctx, thread.ID(), tc.args.content, tc.args.operator)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, th)
				assert.NotNil(t, c)
			}

			th, err = threadUC.FindByID(ctx, thread.ID(), tc.args.operator)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(th.Comments()))
			assert.True(t, th.HasComment(c.ID()))
		})
	}
}

func TestThread_UpdateComment(t *testing.T) {
	c1 := thread.NewComment(
		thread.NewCommentID(),
		operator.OperatorFromUser(accountdomain.NewUserID()),
		"aaa",
	)
	c2 := thread.NewComment(
		thread.NewCommentID(),
		operator.OperatorFromUser(accountdomain.NewUserID()),
		"test",
	)
	wid := accountdomain.NewWorkspaceID()
	th1 := thread.New().NewID().Workspace(wid).Comments([]*thread.Comment{c1, c2}).MustBuild()
	uid := accountdomain.NewUserID()
	op := &usecase.Operator{
		AcOperator: &accountusecase.Operator{
			User:               &uid,
			ReadableWorkspaces: nil,
			WritableWorkspaces: nil,
			OwningWorkspaces:   []accountdomain.WorkspaceID{wid},
		},
	}

	type args struct {
		comment  *thread.Comment
		content  string
		operator *usecase.Operator
	}

	tests := []struct {
		name      string
		seed      *thread.Thread
		args      args
		want      *thread.Comment
		wantErr   error
		mockError bool
	}{
		{
			name: "workspaces operation denied",
			seed: th1,
			args: args{
				comment: c1,
				content: "updated",
				operator: &usecase.Operator{
					AcOperator: &accountusecase.Operator{},
				},
			},
			wantErr: interfaces.ErrInvalidOperator,
		},
		{
			name: "workspaces operation success",
			args: args{
				comment:  c1,
				content:  "updated",
				operator: op,
			},
			seed:    th1,
			wantErr: nil,
		},
		{
			name: "update comment success",
			seed: th1,
			args: args{
				comment:  c1,
				content:  "updated",
				operator: op,
			},
			wantErr: nil,
		},
		{
			name: "update comment fail",
			seed: th1,
			args: args{
				comment:  c1,
				content:  "updated",
				operator: op,
			},
			wantErr:   rerror.ErrNotFound,
			mockError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()

			thread := tc.seed.Clone()
			err := db.Thread.Save(ctx, thread)
			assert.NoError(t, err)

			threadUC := NewThread(db, nil)
			if tc.mockError && tc.wantErr != nil {
				thid := id.NewThreadID()
				_, _, err := threadUC.UpdateComment(
					ctx,
					thid,
					tc.args.comment.ID(),
					tc.args.content,
					tc.args.operator,
				)
				assert.Equal(t, tc.wantErr, err)
				return
			}
			if _, _, err := threadUC.UpdateComment(ctx, thread.ID(), tc.args.comment.ID(), tc.args.content, tc.args.operator); tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			} else {
				assert.NoError(t, err)
			}

			thread2, _ := threadUC.FindByID(ctx, thread.ID(), tc.args.operator)
			comment := thread2.Comments()[0]
			assert.Equal(t, tc.args.content, comment.Content())
		})
	}
}

func TestThread_DeleteComment(t *testing.T) {
	c1 := thread.NewComment(
		thread.NewCommentID(),
		operator.OperatorFromUser(accountdomain.NewUserID()),
		"aaa",
	)
	c2 := thread.NewComment(
		thread.NewCommentID(),
		operator.OperatorFromUser(accountdomain.NewUserID()),
		"test",
	)
	wid := accountdomain.NewWorkspaceID()
	th1 := thread.New().NewID().Workspace(wid).Comments([]*thread.Comment{c1, c2}).MustBuild()
	uid := accountdomain.NewUserID()
	op := &usecase.Operator{
		AcOperator: &accountusecase.Operator{
			User:               &uid,
			ReadableWorkspaces: nil,
			WritableWorkspaces: nil,
			OwningWorkspaces:   []accountdomain.WorkspaceID{wid},
		},
	}

	type args struct {
		commentId id.CommentID
		operator  *usecase.Operator
	}

	tests := []struct {
		name      string
		seed      *thread.Thread
		args      args
		want      *thread.Comment
		wantErr   error
		mockError bool
	}{
		{
			name: "workspaces operation denied",
			seed: th1,
			args: args{commentId: c1.ID(), operator: &usecase.Operator{
				AcOperator: &accountusecase.Operator{},
			}},
			wantErr: interfaces.ErrInvalidOperator,
		},
		{
			name:    "workspaces operation success",
			seed:    th1,
			args:    args{commentId: c1.ID(), operator: op},
			wantErr: nil,
		},
		{
			name:    "delete comment success",
			seed:    th1,
			args:    args{commentId: c1.ID(), operator: op},
			wantErr: nil,
		},
		{
			name:      "delete comment fail",
			seed:      th1,
			args:      args{commentId: c1.ID(), operator: op},
			wantErr:   rerror.ErrNotFound,
			mockError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := memory.New()

			thread1 := tc.seed.Clone()
			err := db.Thread.Save(ctx, thread1)
			assert.NoError(t, err)

			threadUC := NewThread(db, nil)
			if tc.mockError && tc.wantErr != nil {
				thid := id.NewThreadID()
				_, err := threadUC.DeleteComment(ctx, thid, tc.args.commentId, tc.args.operator)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			if _, err := threadUC.DeleteComment(ctx, tc.seed.ID(), tc.args.commentId, tc.args.operator); tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
				return
			} else {
				assert.NoError(t, err)
			}

			commentID := tc.seed.Comments()[0].ID()
			thread2, err := threadUC.FindByID(ctx, tc.seed.ID(), tc.args.operator)
			assert.NoError(t, err)
			assert.Equal(t, len(tc.seed.Comments())-1, len(thread2.Comments()))
			assert.False(
				t,
				lo.ContainsBy(
					thread2.Comments(),
					func(cc *thread.Comment) bool { return cc.ID() == commentID },
				),
			)
		})
	}
}
