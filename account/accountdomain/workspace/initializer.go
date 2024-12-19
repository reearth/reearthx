package workspace

import (
	"github.com/reearth/reearthx/account/accountdomain/user"
	"golang.org/x/text/language"
)

type InitParams struct {
	Email       string
	Name        string
	Sub         *user.Auth
	Password    *string
	Lang        *language.Tag
	Theme       *user.Theme
	UserID      *user.ID
	WorkspaceID *ID
}

func Init(p InitParams) (*user.User, *Workspace, error) {
	if p.UserID == nil {
		newID := user.NewID()
		p.UserID = newID.Ref()
	}
	if p.WorkspaceID == nil {
		newWorkspaceID := NewID()
		p.WorkspaceID = newWorkspaceID.Ref()
	}
	if p.Lang == nil {
		p.Lang = &language.Tag{}
	}
	if p.Theme == nil {
		t := user.ThemeDefault
		p.Theme = &t
	}
	if p.Sub == nil {
		p.Sub = user.GenReearthSub(p.UserID.String())
	}

	b := user.New().
		ID(*p.UserID).
		Name(p.Name).
		Email(p.Email).
		Auths([]user.Auth{*p.Sub}).
		Lang(*p.Lang).
		Theme(*p.Theme)
	if p.Password != nil {
		b = b.PasswordPlainText(*p.Password)
	}
	u, err := b.Build()
	if err != nil {
		return nil, nil, err
	}

	// create a user's own workspace
	t, err := New().
		ID(*p.WorkspaceID).
		Name(p.Name).
		Members(map[user.ID]Member{u.ID(): {Role: RoleOwner, Disabled: false, InvitedBy: u.ID()}}).
		Personal(true).
		Build()
	if err != nil {
		return nil, nil, err
	}
	u.UpdateWorkspace(t.ID())

	return u, t, err
}
