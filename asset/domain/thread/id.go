package thread

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
)

type (
	ID          = id.ThreadID
	CommentID   = id.CommentID
	UserID      = accountdomain.UserID
	WorkspaceID = id.WorkspaceID
)

var (
	NewID          = id.NewThreadID
	NewCommentID   = id.NewCommentID
	NewUserID      = accountdomain.NewUserID
	NewWorkspaceID = accountdomain.NewWorkspaceID
)

var (
	MustID          = id.MustThreadID
	MustCommentID   = id.MustCommentID
	MustUserID      = id.MustUserID
	MustWorkspaceID = id.MustWorkspaceID
)

var (
	IDFrom          = id.ThreadIDFrom
	CommentIDFrom   = id.CommentIDFrom
	UserIDFrom      = accountdomain.UserIDFrom
	WorkspaceIDFrom = id.WorkspaceIDFrom
)

var (
	IDFromRef          = id.ThreadIDFromRef
	CommentIDFromRef   = id.CommentIDFromRef
	UserIDFromRef      = accountdomain.UserIDFromRef
	WorkspaceIDFromRef = id.WorkspaceIDFromRef
)

var ErrInvalidID = id.ErrInvalidID
