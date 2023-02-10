package workspace

import (
	"errors"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/idx"
)

var ErrMembersRequired = errors.New("members required")

type WorkspaceBuilder struct {
	w            *Workspace
	members      map[UserID]Member
	integrations map[IntegrationID]Member
	personal     bool
}

func NewWorkspace() *WorkspaceBuilder {
	return &WorkspaceBuilder{w: &Workspace{}}
}

func (b *WorkspaceBuilder) Build() (*Workspace, error) {
	if b.w.id.IsEmpty() {
		return nil, ErrInvalidID
	}

	if b.members == nil {
		b.w.members = NewMembers(
			map[idx.ID[*accountdomain.UserIDType]]Member{},
			map[idx.ID[*accountdomain.IntegrationIDType]]Member{},
			false,
		)
	} else {
		b.w.members = NewMembersWith(b.members, b.integrations)
	}
	b.w.members.fixed = b.personal
	return b.w, nil
}

func (b *WorkspaceBuilder) MustBuild() *Workspace {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *WorkspaceBuilder) ID(id ID) *WorkspaceBuilder {
	b.w.id = id
	return b
}

func (b *WorkspaceBuilder) NewID() *WorkspaceBuilder {
	b.w.id = NewID()
	return b
}

func (b *WorkspaceBuilder) Name(name string) *WorkspaceBuilder {
	b.w.name = name
	return b
}

func (b *WorkspaceBuilder) Members(members map[UserID]Member) *WorkspaceBuilder {
	b.members = members
	return b
}

func (b *WorkspaceBuilder) Integrations(integrations map[IntegrationID]Member) *WorkspaceBuilder {
	b.integrations = integrations
	return b
}

func (b *WorkspaceBuilder) Personal(p bool) *WorkspaceBuilder {
	b.personal = p
	return b
}
