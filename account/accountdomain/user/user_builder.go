package user

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"golang.org/x/text/language"
)

var ErrInvalidName = rerror.NewE(i18n.T("invalid user name"))

type Builder struct {
	u            *User
	err          error
	passwordText string
	email        string
}

func New() *Builder {
	return &Builder{u: &User{}}
}

func (b *Builder) Build() (*User, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.u.id.IsEmpty() {
		return nil, ErrInvalidID
	}
	if b.u.name == "" {
		return nil, ErrInvalidName
	}
	if !b.u.theme.Valid() {
		b.u.theme = ThemeDefault
	}
	if b.passwordText != "" {
		if err := b.u.SetPassword(b.passwordText); err != nil {
			return nil, err
		}
	}
	if err := b.u.UpdateEmail(b.email); err != nil {
		return nil, err
	}
	return b.u, nil
}

func (b *Builder) MustBuild() *User {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *Builder) ID(id accountdomain.UserID) *Builder {
	b.u.id = id
	return b
}

func (b *Builder) ParseID(id string) *Builder {
	b.u.id, b.err = IDFrom(id)
	return b
}

func (b *Builder) NewID() *Builder {
	b.u.id = NewID()
	return b
}

func (b *Builder) Name(name string) *Builder {
	b.u.name = name
	return b
}

func (b *Builder) DisplayName(displayName string) *Builder {
	b.u.displayName = displayName
	return b
}

func (b *Builder) Email(email string) *Builder {
	b.email = email
	return b
}

func (b *Builder) EncodedPassword(p EncodedPassword) *Builder {
	b.u.password = p.Clone()
	return b
}

func (b *Builder) PasswordPlainText(p string) *Builder {
	b.passwordText = p
	return b
}

func (b *Builder) Workspace(workspace WorkspaceID) *Builder {
	b.u.workspace = workspace
	return b
}

func (b *Builder) Lang(lang language.Tag) *Builder {
	b.u.lang = lang
	return b
}

func (b *Builder) Theme(t Theme) *Builder {
	b.u.theme = t
	return b
}

func (b *Builder) LangFrom(lang string) *Builder {
	if lang == "" {
		b.u.lang = language.Und
	} else if l, err := language.Parse(lang); err == nil {
		b.u.lang = l
	}
	return b
}

func (b *Builder) Auths(auths []Auth) *Builder {
	for _, a := range auths {
		b.u.AddAuth(a)
	}
	return b
}

func (b *Builder) PasswordReset(pr *PasswordReset) *Builder {
	b.u.passwordReset = pr
	return b
}

func (b *Builder) Verification(v *Verification) *Builder {
	b.u.verification = v
	return b
}
