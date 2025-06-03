package interfaces

import (
	"context"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/thread"
	"github.com/reearth/reearthx/asset/usecase"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
)

var (
	ErrCommentAlreadyExist = rerror.NewE(i18n.T("Comment already exist in this thread"))
	ErrCommentDoesNotExist = rerror.NewE(i18n.T("Comment does not exist in this thread"))
)

type ResourceType string

const (
	ResourceTypeItem    ResourceType = "item"
	ResourceTypeAsset   ResourceType = "asset"
	ResourceTypeRequest ResourceType = "request"
)

type CreateThreadWithCommentInput struct {
	ResourceID   string
	ResourceType ResourceType
	Content      string
	WorkspaceID  accountdomain.WorkspaceID
}

type Thread interface {
	FindByID(context.Context, id.ThreadID, *usecase.Operator) (*thread.Thread, error)
	FindByIDs(context.Context, []id.ThreadID, *usecase.Operator) (thread.List, error)
	CreateThreadWithComment(context.Context, CreateThreadWithCommentInput, *usecase.Operator) (*thread.Thread, *thread.Comment, error)
	AddComment(context.Context, id.ThreadID, string, *usecase.Operator) (*thread.Thread, *thread.Comment, error)
	UpdateComment(context.Context, id.ThreadID, id.CommentID, string, *usecase.Operator) (*thread.Thread, *thread.Comment, error)
	DeleteComment(context.Context, id.ThreadID, id.CommentID, *usecase.Operator) (*thread.Thread, error)
}
