package user

import (
	"errors"
	"net/mail"
	"slices"

	"github.com/reearth/reearthx/util"
	"golang.org/x/text/language"
)

var (
	ErrInvalidEmail = errors.New("invalid email")
)

type User struct {
	id            ID
	name          string
	displayName   string
	email         string
	password      EncodedPassword
	workspace     WorkspaceID
	auths         []Auth
	lang          language.Tag
	theme         Theme
	verification  *Verification
	passwordReset *PasswordReset
	host          string
}

func (u *User) ID() ID {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) DisplayName() string {
	if u.displayName == "" {
		return u.name
	}
	return u.displayName
}

func (u *User) Email() string {
	return u.email
}

func (u *User) Workspace() WorkspaceID {
	return u.workspace
}

func (u *User) Lang() language.Tag {
	return u.lang
}

func (u *User) Theme() Theme {
	return u.theme
}

func (u *User) Password() []byte {
	return u.password
}

func (u *User) UpdateName(name string) {
	u.name = name
}

func (u *User) UpdateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	u.email = email
	return nil
}

func (u *User) UpdateWorkspace(workspace WorkspaceID) {
	u.workspace = workspace
}

func (u *User) UpdateLang(lang language.Tag) {
	u.lang = lang
}

func (u *User) UpdateTheme(t Theme) {
	u.theme = t
}

func (u *User) UpdateDisplayName(displayName string) {
	u.displayName = displayName
}

func (u *User) Verification() *Verification {
	return u.verification
}

func (u *User) Auths() Auths {
	if u == nil {
		return nil
	}
	return slices.Clone(u.auths)
}

func (u *User) ContainAuth(a Auth) bool {
	if u == nil {
		return false
	}
	for _, b := range u.auths {
		if a == b || a.Provider == b.Provider {
			return true
		}
	}
	return false
}

func (u *User) HasAuthProvider(p string) bool {
	if u == nil {
		return false
	}
	for _, b := range u.auths {
		if b.Provider == p {
			return true
		}
	}
	return false
}

func (u *User) AddAuth(a Auth) bool {
	if u == nil {
		return false
	}
	if !u.ContainAuth(a) {
		u.auths = append(u.auths, a)
		return true
	}
	return false
}

func (u *User) RemoveAuth(a Auth) bool {
	if u == nil || a.IsAuth0() {
		return false
	}
	for i, b := range u.auths {
		if a == b {
			u.auths = append(u.auths[:i], u.auths[i+1:]...)
			return true
		}
	}
	return false
}

func (u *User) GetAuthByProvider(provider string) *Auth {
	if u == nil || u.auths == nil {
		return nil
	}
	for _, b := range u.auths {
		if provider == b.Provider {
			return &b
		}
	}
	return nil
}

func (u *User) RemoveAuthByProvider(provider string) bool {
	if u == nil || provider == "auth0" {
		return false
	}
	for i, b := range u.auths {
		if provider == b.Provider {
			u.auths = append(u.auths[:i], u.auths[i+1:]...)
			return true
		}
	}
	return false
}

func (u *User) ClearAuths() {
	u.auths = nil
}

func (u *User) SetPassword(pass string) error {
	p, err := NewEncodedPassword(pass)
	if err != nil {
		return err
	}
	u.password = p
	return nil
}

func (u *User) MatchPassword(pass string) (bool, error) {
	if u == nil {
		return false, nil
	}
	return u.password.Verify(pass)
}

func (u *User) PasswordReset() *PasswordReset {
	return u.passwordReset
}

func (u *User) SetPasswordReset(pr *PasswordReset) {
	u.passwordReset = pr.Clone()
}

func (u *User) SetVerification(v *Verification) {
	u.verification = v
}

func (u *User) Host() string {
	return u.host
}

func (u *User) Clone() *User {
	return &User{
		id:            u.id,
		name:          u.name,
		displayName:   u.displayName,
		email:         u.email,
		password:      u.password,
		workspace:     u.workspace,
		auths:         slices.Clone(u.auths),
		lang:          u.lang,
		theme:         u.theme,
		verification:  util.CloneRef(u.verification),
		passwordReset: util.CloneRef(u.passwordReset),
	}
}

func (u *User) WithHost(host string) *User {
	d := u.Clone()
	d.host = host
	return d
}

type List []*User
