package mongodoc

import (
	"time"

	"github.com/google/uuid"
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/request"
	"github.com/reearth/reearthx/asset/domain/version"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
)

type RequestDocument struct {
	UpdatedAt   time.Time
	ApprovedAt  *time.Time
	ClosedAt    *time.Time
	Thread      *string
	ID          string
	Workspace   string
	Project     string
	Title       string
	Description string
	CreatedBy   string
	State       string
	Items       []RequestItem
	Reviewers   []string
}

type RequestItem struct {
	Version *string
	Ref     *string
	Item    string
}

type RequestConsumer = mongox.SliceFuncConsumer[*RequestDocument, *request.Request]

func NewRequestConsumer() *RequestConsumer {
	return NewConsumer[*RequestDocument, *request.Request]()
}

func NewRequest(r *request.Request) (*RequestDocument, string) {
	rid := r.ID().String()
	items := lo.Map(r.Items(), func(i *request.Item, _ int) RequestItem {
		return version.MatchVersionOrRef(
			i.Pointer(),
			func(v version.Version) RequestItem {
				return RequestItem{
					Item:    i.Item().String(),
					Version: lo.ToPtr(v.String()),
				}
			},
			func(r version.Ref) RequestItem {
				return RequestItem{
					Item: i.Item().String(),
					Ref:  lo.ToPtr(r.String()),
				}
			},
		)
	})

	doc, id := &RequestDocument{
		ID:          rid,
		Workspace:   r.Workspace().String(),
		Project:     r.Project().String(),
		Items:       items,
		Title:       r.Title(),
		Description: r.Description(),
		CreatedBy:   r.CreatedBy().String(),
		Reviewers: lo.Map(r.Reviewers(), func(u accountdomain.UserID, i int) string {
			return u.String()
		}),
		State:      r.State().String(),
		UpdatedAt:  r.UpdatedAt(),
		ApprovedAt: r.ApprovedAt(),
		ClosedAt:   r.ClosedAt(),
		Thread:     r.Thread().StringRef(),
	}, rid

	return doc, id
}

func NewRequests(requests request.List) ([]*RequestDocument, []string) {
	res := make([]*RequestDocument, 0, len(requests))
	ids := make([]string, 0, len(requests))
	for _, d := range requests {
		if d == nil {
			continue
		}
		r, rid := NewRequest(d)
		res = append(res, r)
		ids = append(ids, rid)
	}
	return res, ids
}

func (d *RequestDocument) Model() (*request.Request, error) {
	rid, err := id.RequestIDFrom(d.ID)
	if err != nil {
		return nil, err
	}
	pid, err := id.ProjectIDFrom(d.Project)
	if err != nil {
		return nil, err
	}
	wid, err := accountdomain.WorkspaceIDFrom(d.Workspace)
	if err != nil {
		return nil, err
	}
	uid, err := accountdomain.UserIDFrom(d.CreatedBy)
	if err != nil {
		return nil, err
	}
	reviewers, err := id.UserIDListFrom(d.Reviewers)
	if err != nil {
		return nil, err
	}
	items, err := util.TryMap(d.Items, func(ri RequestItem) (*request.Item, error) {
		iid, err := id.ItemIDFrom(ri.Item)
		if err != nil {
			return nil, err
		}
		var vor version.VersionOrRef
		if ri.Version != nil {
			v, err := uuid.Parse(*ri.Version)
			if err != nil {
				return nil, err
			}
			vor = version.Version(v).OrRef()
		} else if ri.Ref != nil {
			vor = version.Ref(*ri.Ref).OrVersion()
		}
		return request.NewItemWithVersion(iid, vor)
	})
	if err != nil {
		return nil, err
	}

	builder := request.New().
		ID(rid).
		Project(pid).
		Workspace(wid).
		CreatedBy(uid).
		Items(items).
		Title(d.Title).
		Description(d.Description).
		State(request.StateFrom(d.State)).
		UpdatedAt(d.UpdatedAt).
		ClosedAt(d.ClosedAt).
		ApprovedAt(d.ApprovedAt).
		Reviewers(reviewers).
		Thread(id.ThreadIDFromRef(d.Thread))

	return builder.Build()
}
