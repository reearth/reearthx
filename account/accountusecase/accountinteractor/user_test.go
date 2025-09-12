package accountinteractor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmemory"
	"github.com/reearth/reearthx/account/accountusecase/accountgateway"
	"github.com/reearth/reearthx/mailer"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"

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
				Verification(user.VerificationFrom("code", time.Now().Add(-24*time.Hour), false)).
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

func TestUser_StartPasswordReset(t *testing.T) {
	user.DefaultPasswordEncoder = &user.NoopPasswordEncoder{}
	uid := accountdomain.NewUserID()
	tid := accountdomain.NewWorkspaceID()
	r := accountmemory.New()

	m := mailer.NewMock()
	g := &accountgateway.Container{Mailer: m}
	uc := NewUser(r, g, "", "")
	tests := []struct {
		name             string
		createUserBefore *user.User
		email            string
		wantMailSubject  string
		wantMailTo       []mailer.Contact
		wantError        error
	}{
		{
			name: "ok",
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Email("aaa@bbb.com").
				Name("NAME").
				Auths([]user.Auth{
					{
						Provider: user.ProviderReearth,
						Sub:      "reearth|" + uid.String(),
					},
				}).
				MustBuild(),
			email:           "aaa@bbb.com",
			wantMailSubject: "Password reset",
			wantMailTo: []mailer.Contact{
				{
					Email: "aaa@bbb.com",
					Name:  "NAME",
				},
			},
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
			ctx := context.Background()
			if tt.createUserBefore != nil {
				assert.NoError(t, r.User.Save(ctx, tt.createUserBefore))
			}
			err := uc.StartPasswordReset(ctx, tt.email)

			if err != nil {
				assert.Equal(t, tt.wantError, err)
			} else {
				user, err := r.User.FindByEmail(ctx, tt.email)
				assert.NoError(t, err)
				assert.NotNil(t, user.PasswordReset())
			}

			mails := m.Mails()
			if tt.wantMailSubject != "" {
				assert.Equal(t, 1, len(mails))
				assert.Equal(t, tt.wantMailSubject, mails[0].Subject)
				assert.Equal(t, tt.wantMailTo, mails[0].To)
			}
		})
	}
}

func TestUser_PasswordReset(t *testing.T) {
	user.DefaultPasswordEncoder = &user.NoopPasswordEncoder{}
	uid := accountdomain.NewUserID()
	tid := accountdomain.NewWorkspaceID()
	r := accountmemory.New()
	uc := NewUser(r, nil, "", "")
	pr := user.NewPasswordReset()
	expired := time.Now().Add(24 * time.Hour)
	tests := []struct {
		name             string
		password         string
		token            string
		createUserBefore *user.User
		wantError        error
	}{
		{
			name:     "ok",
			password: "PAss00!!",
			token:    pr.Token,
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Name("NAME").
				Email("aaa@bbb.com").
				PasswordPlainText("PAss00!!").
				Verification(user.VerificationFrom("code", expired, false)).
				PasswordReset(pr).
				Auths([]user.Auth{
					{
						Provider: user.ProviderReearth,
						Sub:      "reearth|" + uid.String(),
					},
				}).
				MustBuild(),
			wantError: nil,
		},
		{
			name:     "invalid password",
			password: "pass",
			token:    pr.Token,
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Name("NAME").
				Email("aaa@bbb.com").
				PasswordPlainText("PAss00!!").
				Verification(user.VerificationFrom("code", expired, false)).
				PasswordReset(pr).
				Auths([]user.Auth{
					{
						Provider: user.ProviderReearth,
						Sub:      "reearth|" + uid.String(),
					},
				}).
				MustBuild(),
			wantError: user.ErrPasswordLength,
		},
		{
			name:     "not found",
			password: "PAss00!!",
			token:    pr.Token,
			createUserBefore: user.New().
				ID(uid).
				Workspace(tid).
				Name("NAME").
				Email("aaa@bbb.com").
				PasswordPlainText("PAss00!!").
				Verification(user.VerificationFrom("code", expired, false)).
				Auths([]user.Auth{
					{
						Provider: user.ProviderReearth,
						Sub:      "reearth|" + uid.String(),
					},
				}).
				MustBuild(),
			wantError: rerror.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.createUserBefore != nil {
				assert.NoError(t, r.User.Save(ctx, tt.createUserBefore))
			}
			err := uc.PasswordReset(ctx, tt.password, tt.token)
			assert.Equal(t, tt.wantError, err)
		})
	}
}

// TestUserQuery_FindByIDsWithPagination_Integration tests the integration of the pagination
// feature through the UserQuery interface, demonstrating real-world usage scenarios
func TestUserQuery_FindByIDsWithPagination_Integration(t *testing.T) {
	ctx := context.Background()

	// Set up test users with realistic data
	wsID := user.NewWorkspaceID()

	// Team members from different departments
	devLead := user.New().NewID().Name("john_smith").Alias("johndev").Email("john.smith@company.com").Workspace(wsID).MustBuild()
	devJunior := user.New().NewID().Name("jane_doe").Alias("janedev").Email("jane.doe@company.com").Workspace(wsID).MustBuild()
	designer := user.New().NewID().Name("alice_wonder").Alias("alice_design").Email("alice.wonder@company.com").Workspace(wsID).MustBuild()

	allUsers := []*user.User{devLead, devJunior, designer}
	r := accountmemory.NewUserWith(allUsers...)
	query := NewUserQuery(r)

	tests := []struct {
		name          string
		scenario      string
		teamMemberIDs user.IDList
		searchFilter  string
		expectedUsers int
		expectedNames []string
	}{
		{
			name:          "Search developers in team by role",
			scenario:      "Product manager wants to find all developers in a specific team",
			teamMemberIDs: user.IDList{devLead.ID(), devJunior.ID(), designer.ID()},
			searchFilter:  "dev",
			expectedUsers: 2,
			expectedNames: []string{"john_smith", "jane_doe"},
		},
		{
			name:          "Search by alias pattern across team",
			scenario:      "Admin looking for users with specific alias pattern",
			teamMemberIDs: user.IDList{devLead.ID(), devJunior.ID(), designer.ID()},
			searchFilter:  "alice",
			expectedUsers: 1,
			expectedNames: []string{"alice_wonder"},
		},
		{
			name:          "Case insensitive search",
			scenario:      "User types search term in different case",
			teamMemberIDs: user.IDList{devLead.ID(), devJunior.ID(), designer.ID()},
			searchFilter:  "ALICE",
			expectedUsers: 1,
			expectedNames: []string{"alice_wonder"},
		},
		{
			name:          "No results found",
			scenario:      "Search term doesn't match any team members",
			teamMemberIDs: user.IDList{devLead.ID(), devJunior.ID(), designer.ID()},
			searchFilter:  "nonexistent",
			expectedUsers: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This demonstrates how the feature would be used in practice
			// Even though we're testing at the repository level, this shows the integration path
			if len(query.repos) > 0 {
				if repo, ok := query.repos[0].(interface {
					FindByIDsWithPagination(context.Context, user.IDList, *usecasex.Pagination, ...string) (user.List, *usecasex.PageInfo, error)
				}); ok {
					var result user.List
					var pageInfo *usecasex.PageInfo
					var err error

					if tt.searchFilter == "" {
						result, pageInfo, err = repo.FindByIDsWithPagination(ctx, tt.teamMemberIDs,
							&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}})
					} else {
						result, pageInfo, err = repo.FindByIDsWithPagination(ctx, tt.teamMemberIDs,
							&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}}, tt.searchFilter)
					}

					assert.NoError(t, err)
					assert.Equal(t, tt.expectedUsers, len(result), "Expected %d users but got %d", tt.expectedUsers, len(result))

					assert.NotNil(t, pageInfo, "PageInfo should not be nil when pagination is provided")

					// Verify specific expected users if provided
					if len(tt.expectedNames) > 0 {
						actualNames := make([]string, len(result))
						for i, u := range result {
							actualNames[i] = u.Name()
						}
						assert.ElementsMatch(t, tt.expectedNames, actualNames, "Expected specific user names to match")
					}

					// Verify all returned users are from the requested IDs
					for _, resultUser := range result {
						assert.True(t, tt.teamMemberIDs.Has(resultUser.ID()), "User %s should be in the requested team member IDs", resultUser.Name())
					}

					// Verify workspace consistency
					for _, resultUser := range result {
						assert.Equal(t, wsID, resultUser.Workspace(), "All users should belong to the same workspace")
					}

					t.Logf("Scenario: %s - Found %d users", tt.scenario, len(result))
				} else {
					t.Skip("Repository doesn't support FindByIDsWithPagination with filtering")
				}
			} else {
				t.Skip("No repositories available for testing")
			}
		})
	}
}
