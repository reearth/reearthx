package accountinteractor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmemory"
	"github.com/reearth/reearthx/rerror"

	"github.com/stretchr/testify/assert"
)

func TestUser_VerifyUser(t *testing.T) {
	user.DefaultPasswordEncoder = &user.NoopPasswordEncoder{}
	uid := accountdomain.NewUserID()
	tid := accountdomain.NewWorkspaceID()
	r := accountmemory.New()
	uc := NewUser(r, nil, "", "")
	expired := time.Now().Add(24 * time.Hour)
	tests := []struct {
		name             string
		code             string
		createUserBefore *user.User
		wantUser         func(u *user.User) *user.User
		wantError        error
	}{
		{
			name: "ok",
			code: "code",
			wantUser: func(u *user.User) *user.User {
				return user.New().
					ID(uid).
					Workspace(tid).
					Name("NAME").
					Email("aaa@bbb.com").
					PasswordPlainText("PAss00!!").
					Verification(user.VerificationFrom("code", expired, true)).
					MustBuild()
			},
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Name("NAME").
				Email("aaa@bbb.com").
				PasswordPlainText("PAss00!!").
				Verification(user.VerificationFrom("code", expired, false)).
				MustBuild(),
			wantError: nil,
		},
		{
			name:     "expired",
			code:     "code",
			wantUser: nil,
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Name("NAME").
				Email("aaa@bbb.com").
				PasswordPlainText("PAss00!!").
				Verification(user.VerificationFrom("code", time.Now(), false)).
				MustBuild(),
			wantError: errors.New("verification expired"),
		},
		{
			name:     "not found",
			code:     "codesss",
			wantUser: nil,
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Name("NAME").
				Email("aaa@bbb.com").
				PasswordPlainText("PAss00!!").
				Verification(user.VerificationFrom("code", expired, false)).
				MustBuild(),
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
			u, err := uc.VerifyUser(ctx, tt.code)

			if tt.wantUser != nil {
				assert.Equal(t, tt.wantUser(u), u)
			} else {
				assert.Nil(t, u)
			}
			assert.Equal(t, tt.wantError, err)
		})
	}
}
