package accountinteractor

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/samber/lo"
	"golang.org/x/text/language"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmemory"
	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
	"github.com/reearth/reearthx/mailer"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"

	"github.com/stretchr/testify/assert"
)

func TestUser_Signup(t *testing.T) {
	user.DefaultPasswordEncoder = &user.NoopPasswordEncoder{}
	uid := accountdomain.NewUserID()
	tid := accountdomain.NewWorkspaceID()
	mocktime := time.Time{}
	mockcode := "CODECODE"

	tests := []struct {
		name             string
		signupSecret     string
		authSrvUIDomain  string
		createUserBefore *user.User
		args             accountinterfaces.SignupParam
		wantUser         func(u *user.User) *user.User
		wantWorkspace    *workspace.Workspace
		wantMailTo       []accountgateway.Contact
		wantMailSubject  string
		wantMailContent  string
		wantError        error
	}{
		{
			name:            "without secret",
			signupSecret:    "",
			authSrvUIDomain: "https://reearth.io",
			args: accountinterfaces.SignupParam{
				Email:       "aaa@bbb.com",
				Name:        "NAME",
				Password:    "PAss00!!",
				UserID:      &uid,
				WorkspaceID: &tid,
			},
			wantUser: func(u *user.User) *user.User {
				return user.New().
					ID(uid).
					Workspace(tid).
					Name("NAME").
					Auths(u.Auths()).
					Email("aaa@bbb.com").
					PasswordPlainText("PAss00!!").
					Verification(user.VerificationFrom(mockcode, mocktime.Add(24*time.Hour), false)).
					MustBuild()
			},
			wantWorkspace: workspace.NewWorkspace().
				ID(tid).
				Name("NAME").
				Members(map[user.ID]workspace.Member{uid: {Role: workspace.RoleOwner, Disabled: false, InvitedBy: uid}}).
				Personal(true).
				MustBuild(),
			wantMailTo:      []accountgateway.Contact{{Email: "aaa@bbb.com", Name: "NAME"}},
			wantMailSubject: "email verification",
			wantMailContent: "https://reearth.io/?user-verification-token=CODECODE",
			wantError:       nil,
		},
		{
			name:            "existing but not valdiated user",
			signupSecret:    "",
			authSrvUIDomain: "",
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Name("NAME").
				Email("aaa@bbb.com").
				MustBuild(),
			args: accountinterfaces.SignupParam{
				Email:       "aaa@bbb.com",
				Name:        "NAME",
				Password:    "PAss00!!",
				UserID:      &uid,
				WorkspaceID: &tid,
			},
			wantUser:      nil,
			wantWorkspace: nil,
			wantError:     accountinterfaces.ErrUserAlreadyExists,
		},
		{
			name:            "existing and valdiated user",
			signupSecret:    "",
			authSrvUIDomain: "",
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Email("aaa@bbb.com").
				Name("NAME").
				Verification(user.VerificationFrom(mockcode, mocktime, true)).
				MustBuild(),
			args: accountinterfaces.SignupParam{
				Email:       "aaa@bbb.com",
				Name:        "NAME",
				Password:    "PAss00!!",
				UserID:      &uid,
				WorkspaceID: &tid,
			},
			wantUser:      nil,
			wantWorkspace: nil,
			wantError:     accountinterfaces.ErrUserAlreadyExists,
		},
		{
			name:            "without secret 2",
			signupSecret:    "",
			authSrvUIDomain: "",
			args: accountinterfaces.SignupParam{
				Email:       "aaa@bbb.com",
				Name:        "NAME",
				Password:    "PAss00!!",
				Secret:      lo.ToPtr("hogehoge"),
				UserID:      &uid,
				WorkspaceID: &tid,
			},
			wantUser: func(u *user.User) *user.User {
				return user.New().
					ID(uid).
					Workspace(tid).
					Name("NAME").
					Auths(u.Auths()).
					Email("aaa@bbb.com").
					PasswordPlainText("PAss00!!").
					Verification(user.VerificationFrom(mockcode, mocktime.Add(24*time.Hour), false)).
					MustBuild()
			},
			wantWorkspace: workspace.NewWorkspace().
				ID(tid).
				Name("NAME").
				Members(map[user.ID]workspace.Member{uid: {Role: workspace.RoleOwner, Disabled: false, InvitedBy: uid}}).
				Personal(true).
				MustBuild(),
			wantMailTo:      []accountgateway.Contact{{Email: "aaa@bbb.com", Name: "NAME"}},
			wantMailSubject: "email verification",
			wantMailContent: "/?user-verification-token=CODECODE",
			wantError:       nil,
		},
		{
			name:            "with secret",
			signupSecret:    "SECRET",
			authSrvUIDomain: "",
			args: accountinterfaces.SignupParam{
				Email:       "aaa@bbb.com",
				Name:        "NAME",
				Password:    "PAss00!!",
				Secret:      lo.ToPtr("SECRET"),
				UserID:      &uid,
				WorkspaceID: &tid,
				Lang:        &language.Japanese,
				Theme:       user.ThemeDark.Ref(),
			},
			wantUser: func(u *user.User) *user.User {
				return user.New().
					ID(uid).
					Workspace(tid).
					Name("NAME").
					Auths(u.Auths()).
					Email("aaa@bbb.com").
					PasswordPlainText("PAss00!!").
					Lang(language.Japanese).
					Theme(user.ThemeDark).
					Verification(user.VerificationFrom(mockcode, mocktime.Add(24*time.Hour), false)).
					MustBuild()
			},
			wantWorkspace: workspace.NewWorkspace().
				ID(tid).
				Name("NAME").
				Members(map[user.ID]workspace.Member{uid: {Role: workspace.RoleOwner, Disabled: false, InvitedBy: uid}}).
				Personal(true).
				MustBuild(),
			wantMailTo:      []accountgateway.Contact{{Email: "aaa@bbb.com", Name: "NAME"}},
			wantMailSubject: "email verification",
			wantMailContent: "/?user-verification-token=CODECODE",
			wantError:       nil,
		},
		{
			name:            "invalid secret",
			signupSecret:    "SECRET",
			authSrvUIDomain: "",
			args: accountinterfaces.SignupParam{
				Email:    "aaa@bbb.com",
				Name:     "NAME",
				Password: "PAss00!!",
				Secret:   lo.ToPtr("SECRET!"),
			},
			wantError: accountinterfaces.ErrSignupInvalidSecret,
		},
		{
			name:            "invalid secret 2",
			signupSecret:    "SECRET",
			authSrvUIDomain: "",
			args: accountinterfaces.SignupParam{
				Email:    "aaa@bbb.com",
				Name:     "NAME",
				Password: "PAss00!!",
			},
			wantError: accountinterfaces.ErrSignupInvalidSecret,
		},
		{
			name: "invalid email",
			args: accountinterfaces.SignupParam{
				Email:    "aaa",
				Name:     "NAME",
				Password: "PAss00!!",
			},
			wantError: user.ErrInvalidEmail,
		},
		{
			name: "invalid password",
			args: accountinterfaces.SignupParam{
				Email:    "aaa@bbb.com",
				Name:     "NAME",
				Password: "PAss00",
			},
			wantError: user.ErrPasswordLength,
		},
		{
			name: "invalid name",
			args: accountinterfaces.SignupParam{
				Email:    "aaa@bbb.com",
				Name:     "",
				Password: "Ass00!!",
			},
			wantError: user.ErrInvalidName,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel() cannot be used because Now and GenerateVerificationCode are mocked

			defer util.MockNow(mocktime)()
			defer user.MockGenerateVerificationCode(mockcode)()

			ctx := context.Background()
			r := accountmemory.New()
			if tt.createUserBefore != nil {
				assert.NoError(t, r.User.Save(ctx, tt.createUserBefore))
			}

			m := mailer.NewMock()
			g := &accountgateway.Container{Mailer: m}
			uc := NewUser(r, g, tt.signupSecret, tt.authSrvUIDomain)
			u, err := uc.Signup(ctx, tt.args)

			if tt.wantUser != nil {
				assert.Equal(t, tt.wantUser(u), u)
			} else {
				assert.Nil(t, u)
			}

			var ws *workspace.Workspace
			if u != nil {
				ws, _ = r.Workspace.FindByID(ctx, u.Workspace())
			}
			assert.Equal(t, tt.wantWorkspace, ws)

			assert.Equal(t, tt.wantError, err)

			mails := m.Mails()
			if tt.wantMailSubject == "" {
				assert.Empty(t, mails)
			} else {
				assert.Equal(t, 1, len(mails))
				assert.Equal(t, tt.wantMailSubject, mails[0].Subject)
				assert.Equal(t, tt.wantMailTo, mails[0].To)
				assert.Contains(t, mails[0].PlainContent, tt.wantMailContent)
			}
		})
	}
}

func TestUser_FindOrCreate(t *testing.T) {
	r := accountmemory.New()
	uc := NewUser(r, nil, "", "")

	httpmock.Activate()
	defer httpmock.Deactivate()

	httpmock.RegisterResponder("GET", "https://example.com/.well-known/openid-configuration", lo.Must(httpmock.NewJsonResponder(200, map[string]any{
		"userinfo_endpoint": "https://example.com/userinfo",
	})))

	httpmock.RegisterResponder("GET", "https://example.com/userinfo", lo.Must(httpmock.NewJsonResponder(200, map[string]any{
		"sub":   "auth0|SUB",
		"name":  "NAME",
		"email": "aaa@example.com",
	})))

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		go func() {
			_, err := uc.FindOrCreate(context.Background(), accountinterfaces.UserFindOrCreateParam{
				Sub:   "auth0|SUB",
				ISS:   "https://example.com",
				Token: "token",
			})
			assert.NoError(t, err)
			wg.Done()
		}()
		wg.Add(1)
	}
	wg.Wait()

	u, _ := r.User.FindBySub(context.Background(), "auth0|SUB")
	assert.Equal(
		t,
		user.New().
			ID(u.ID()).
			Workspace(u.Workspace()).
			Name("NAME").
			Email("aaa@example.com").
			Auths([]user.Auth{
				user.AuthFrom("auth0|SUB"),
			}).
			MustBuild(),
		u)
}

func TestIssToURL(t *testing.T) {
	assert.Nil(t, issToURL("", ""))
	assert.Equal(t, &url.URL{Scheme: "https", Host: "iss.com"}, issToURL("iss.com", ""))
	assert.Equal(t, &url.URL{Scheme: "https", Host: "iss.com"}, issToURL("https://iss.com", ""))
	assert.Equal(t, &url.URL{Scheme: "http", Host: "iss.com"}, issToURL("http://iss.com", ""))
	assert.Equal(t, &url.URL{Scheme: "https", Host: "iss.com", Path: ""}, issToURL("https://iss.com/", ""))
	assert.Equal(t, &url.URL{Scheme: "https", Host: "iss.com", Path: "/hoge"}, issToURL("https://iss.com/hoge", ""))
	assert.Equal(t, &url.URL{Scheme: "https", Host: "iss.com", Path: "/hoge/foobar"}, issToURL("https://iss.com/hoge", "foobar"))
}

func TestUser_CreateVerification(t *testing.T) {
	user.DefaultPasswordEncoder = &user.NoopPasswordEncoder{}
	uid := accountdomain.NewUserID()
	tid := accountdomain.NewWorkspaceID()
	r := accountmemory.New()

	m := mailer.NewMock()
	g := &accountgateway.Container{Mailer: m}
	uc := NewUser(r, g, "", "")
	mocktime := time.Time{}
	mockcode := "CODECODE"

	tests := []struct {
		name             string
		createUserBefore *user.User
		email            string
		wantError        error
	}{
		{
			name: "ok",
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Email("aaa@bbb.com").
				Name("NAME").
				Verification(user.VerificationFrom(mockcode, mocktime, false)).
				MustBuild(),
			email:     "aaa@bbb.com",
			wantError: nil,
		},
		{
			name: "verified user",
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Email("aaa@bbb.com").
				Name("NAME").
				Verification(user.VerificationFrom(mockcode, mocktime, true)).
				MustBuild(),
			email:     "aaa@bbb.com",
			wantError: nil,
		},
		{
			name:      "not found",
			email:     "ccc@bbb.com",
			wantError: rerror.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			if tt.createUserBefore != nil {
				assert.NoError(t, r.User.Save(ctx, tt.createUserBefore))
			}
			err := uc.CreateVerification(ctx, tt.email)

			if err != nil {
				assert.Equal(t, tt.wantError, err)
			} else {
				user, err := r.User.FindByEmail(ctx, tt.email)
				assert.NoError(t, err)
				assert.NotNil(t, user.Verification())
			}
		})
	}
}
