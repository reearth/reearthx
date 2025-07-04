package mongodoc

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/operator"
	"github.com/reearth/reearthx/asset/domain/thread"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/util"
)

type ThreadDocument struct {
	ID        string
	Workspace string
	Comments  []*CommentDocument
}

type CommentDocument struct {
	ID          string
	User        *string
	Integration *string
	Content     string
}

type ThreadConsumer = mongox.SliceFuncConsumer[*ThreadDocument, *thread.Thread]

func NewThreadConsumer() *ThreadConsumer {
	return NewConsumer[*ThreadDocument, *thread.Thread]()
}

func NewThread(a *thread.Thread) (*ThreadDocument, string) {
	thid := a.ID().String()
	comments := util.Map(a.Comments(), NewComment)
	thd, id := &ThreadDocument{
		ID:        thid,
		Workspace: a.Workspace().String(),
		Comments:  comments,
	}, thid

	return thd, id
}

func NewThreads(a thread.List) ([]ThreadDocument, []string) {
	res := make([]ThreadDocument, 0, len(a))
	ids := make([]string, 0, len(a))
	for _, th := range a {
		if th == nil {
			continue
		}
		thDoc, thId := NewThread(th)
		res = append(res, *thDoc)
		ids = append(ids, thId)
	}
	return res, ids
}

func (d *ThreadDocument) Model() (*thread.Thread, error) {
	thid, err := id.ThreadIDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	wid, err := accountdomain.WorkspaceIDFrom(d.Workspace)
	if err != nil {
		return nil, err
	}

	comments := util.Map(d.Comments, func(c *CommentDocument) *thread.Comment {
		return c.Model()
	})

	return thread.New().
		ID(thid).
		Workspace(wid).
		Comments(comments).
		Build()
}

func NewComment(c *thread.Comment) *CommentDocument {
	if c == nil {
		return nil
	}

	return &CommentDocument{
		ID:          c.ID().String(),
		User:        c.Author().User().StringRef(),
		Integration: c.Author().Integration().StringRef(),
		Content:     c.Content(),
	}
}

func (c *CommentDocument) Model() *thread.Comment {
	if c == nil {
		return nil
	}

	cid, err := id.CommentIDFrom(c.ID)
	if err != nil {
		return nil
	}

	var author operator.Operator
	if c.User != nil {
		if uid := accountdomain.UserIDFromRef(c.User); uid != nil {
			author = operator.OperatorFromUser(*uid)
		}
	} else if c.Integration != nil {
		if iid := id.IntegrationIDFromRef(c.Integration); iid != nil {
			author = operator.OperatorFromIntegration(*iid)
		}
	}

	return thread.NewComment(cid, author, c.Content)
}
