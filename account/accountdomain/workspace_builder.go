package accountdomain

import "errors"

var ErrMembersRequired = errors.New("members required")

type WorkspaceBuilder struct {
	w *Workspace
}

func NewWorkspace() *WorkspaceBuilder {
	return &WorkspaceBuilder{w: &Workspace{}}
}

func (b *WorkspaceBuilder) Build() (*Workspace, error) {
	if b.w.id.IsEmpty() {
		return nil, ErrInvalidID
	}
	if b.w.members == nil || b.w.members.IsEmpty() {
		return nil, ErrMembersRequired
	}
	return b.w, nil
}

func (b *WorkspaceBuilder) MustBuild() *Workspace {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *WorkspaceBuilder) ID(id WorkspaceID) *WorkspaceBuilder {
	b.w.id = id
	return b
}

func (b *WorkspaceBuilder) NewID(domain string) *WorkspaceBuilder {
	b.w.id = GenerateWorkspaceID(domain)
	return b
}

func (b *WorkspaceBuilder) Name(name string) *WorkspaceBuilder {
	b.w.name = name
	return b
}

func (b *WorkspaceBuilder) Members(members *Members) *WorkspaceBuilder {
	b.w.members = members
	return b
}
