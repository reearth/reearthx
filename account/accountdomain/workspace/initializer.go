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
		p.UserID = user.NewID().Ref()
	}
	if p.WorkspaceID == nil {
		p.WorkspaceID = NewID().Ref()
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

	metadata := user.NewMetadata()
	metadata.LangFrom(p.Lang.String())
	metadata.SetTheme(*p.Theme)

	b := user.New().
		ID(*p.UserID).
		Name(p.Name).
		Email(p.Email).
		Metadata(metadata).
		Auths([]user.Auth{*p.Sub})
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
