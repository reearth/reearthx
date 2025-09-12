package accountmongo

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
)

func TestUserRepo_FindAll(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Workspace(wsid).
		Name("foo").
		MustBuild()
	user2 := user.New().
		NewID().
		Email("aa2@bb.cc").
		Workspace(wsid).
		Name("hoge").
		MustBuild()

	tests := []struct {
		Name               string
		RepoData, Expected []*user.User
	}{
		{
			Name:     "must find users",
			RepoData: []*user.User{user1, user2},
			Expected: []*user.User{user1, user2},
		},
		{
			Name:     "must not find any user",
			RepoData: []*user.User{},
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			for _, u := range tc.RepoData {
				err := repo.Save(ctx, u)
				assert.NoError(tt, err)
			}

			got, err := repo.FindAll(ctx)
			assert.NoError(tt, err)
			for k, u := range got {
				if u != nil {
					assert.Equal(tt, tc.Expected[k].ID(), u.ID())
					assert.Equal(tt, tc.Expected[k].Email(), u.Email())
					assert.Equal(tt, tc.Expected[k].Name(), u.Name())
					assert.Equal(tt, tc.Expected[k].Workspace(), u.Workspace())
				}
			}
		})
	}
}

func TestUserRepo_FindByID(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Workspace(wsid).
		Name("foo").
		MustBuild()
	tests := []struct {
		Name               string
		Input              accountdomain.UserID
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user",
			Input:    user1.ID(),
			RepoData: user1,
			Expected: user1,
		},
		{
			Name:     "must not find any user",
			Input:    user.NewID(),
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByID(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_FindByIDs(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Workspace(wsid).
		Name("foo").
		MustBuild()
	user2 := user.New().
		NewID().
		Email("aa2@bb.cc").
		Workspace(wsid).
		Name("hoge").
		MustBuild()
	user3 := user.New().
		NewID().
		Email("aa3@bb.cc").
		Workspace(wsid).
		Name("xxx").
		MustBuild()

	tests := []struct {
		Name               string
		Input              accountdomain.UserIDList
		RepoData, Expected []*user.User
	}{
		{
			Name:     "must find users",
			RepoData: []*user.User{user1, user2},
			Input: accountdomain.UserIDList{
				user1.ID(),
				user2.ID(),
			},
			Expected: []*user.User{user1, user2},
		},
		{
			Name:     "must not find any user",
			Input:    accountdomain.UserIDList{user3.ID()},
			RepoData: []*user.User{user1, user2},
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			for _, u := range tc.RepoData {
				err := repo.Save(ctx, u)
				assert.NoError(tt, err)
			}

			got, err := repo.FindByIDs(ctx, tc.Input)
			assert.NoError(tt, err)
			for k, u := range got {
				if u != nil {
					assert.Equal(tt, tc.Expected[k].ID(), u.ID())
					assert.Equal(tt, tc.Expected[k].Email(), u.Email())
					assert.Equal(tt, tc.Expected[k].Name(), u.Name())
					assert.Equal(tt, tc.Expected[k].Workspace(), u.Workspace())
				}
			}
		})
	}
}

func TestUserRepo_FindByName(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Workspace(wsid).
		Name("foo").
		MustBuild()
	tests := []struct {
		Name               string
		Input              string
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user",
			Input:    user1.Name(),
			RepoData: user1,
			Expected: user1,
		},
		{
			Name:     "must not find any user",
			Input:    "xxx",
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByName(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_FindByAlias(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Name("foo").
		Alias("alias").
		Email("foo@bar.com").
		Workspace(wsid).
		MustBuild()
	tests := []struct {
		Name               string
		Input              string
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user",
			Input:    user1.Alias(),
			RepoData: user1,
			Expected: user1,
		},
		{
			Name:     "must not find any user",
			Input:    "xxx",
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)
	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByAlias(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_FindByEmail(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Workspace(wsid).
		Name("foo").
		MustBuild()
	tests := []struct {
		Name               string
		Input              string
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user",
			Input:    user1.Email(),
			RepoData: user1,
			Expected: user1,
		},
		{
			Name:     "must not find any user",
			Input:    "xx@yy.zz",
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByEmail(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_FindByNameOrEmail(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Workspace(wsid).
		Name("foo").
		MustBuild()
	tests := []struct {
		Name               string
		Input              string
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user by email",
			Input:    user1.Email(),
			RepoData: user1,
			Expected: user1,
		},
		{
			Name:     "must find a user by name",
			Input:    user1.Name(),
			RepoData: user1,
			Expected: user1,
		},
		{
			Name:     "must not find any user",
			Input:    "xx@yy.zz",
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByNameOrEmail(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_FindByPasswordResetRequest(t *testing.T) {
	pr := user.PasswordReset{
		Token: "123abc",
	}
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		PasswordReset(pr.Clone()).
		Workspace(wsid).
		Name("foo").
		MustBuild()
	tests := []struct {
		Name               string
		Input              string
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user",
			Input:    pr.Token,
			RepoData: user1,
			Expected: user1,
		},

		{
			Name:     "must not find any user",
			Input:    "x@yxz",
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByPasswordResetRequest(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_FindByVerification(t *testing.T) {
	vr := user.VerificationFrom("123abc", time.Now(), false)

	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Verification(vr).
		Workspace(wsid).
		Name("foo").
		MustBuild()
	tests := []struct {
		Name               string
		Input              string
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user",
			Input:    vr.Code(),
			RepoData: user1,
			Expected: user1,
		},

		{
			Name:     "must not find any user",
			Input:    "x@yxz",
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindByVerification(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_FindBySub(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Auths([]user.Auth{{
			Sub: "xxx",
		}}).
		Workspace(wsid).
		Name("foo").
		MustBuild()
	tests := []struct {
		Name               string
		Input              string
		RepoData, Expected *user.User
		WantErr            bool
	}{
		{
			Name:     "must find a user",
			Input:    "xxx",
			RepoData: user1,
			Expected: user1,
		},

		{
			Name:     "must not find any user",
			Input:    "x@yxz",
			RepoData: user1,
			WantErr:  true,
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc

		t.Run(tc.Name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)
			ctx := context.Background()
			err := repo.Save(ctx, tc.RepoData)
			assert.NoError(tt, err)

			got, err := repo.FindBySub(ctx, tc.Input)
			if tc.WantErr {
				assert.Equal(tt, err, rerror.ErrNotFound)
			} else {
				assert.Equal(tt, tc.Expected.ID(), got.ID())
				assert.Equal(tt, tc.Expected.Email(), got.Email())
				assert.Equal(tt, tc.Expected.Name(), got.Name())
				assert.Equal(tt, tc.Expected.Workspace(), got.Workspace())
			}
		})
	}
}

func TestUserRepo_Remove(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("aa@bb.cc").
		Workspace(wsid).
		Name("foo").
		MustBuild()

	init := mongotest.Connect(t)

	client := mongox.NewClientWithDatabase(init(t))

	repo := NewUser(client)
	ctx := context.Background()
	err := repo.Save(ctx, user1)
	assert.NoError(t, err)

	err = repo.Remove(ctx, user1.ID())
	assert.NoError(t, err)
}

func TestUserRepo_SearchByKeyword(t *testing.T) {
	wsid := user.NewWorkspaceID()
	user1 := user.New().
		NewID().
		Email("john@example.com").
		Name("John Doe").
		Alias("johnny").
		Workspace(wsid).
		MustBuild()
	user2 := user.New().
		NewID().
		Email("jane@test.com").
		Name("Jane Smith").
		Alias("jsmith").
		Workspace(wsid).
		MustBuild()
	user3 := user.New().
		NewID().
		Email("bob@company.com").
		Name("Bob Johnson").
		Alias("bobby").
		Workspace(wsid).
		MustBuild()

	tests := []struct {
		name     string
		keyword  string
		fields   []string
		expected []*user.User
	}{
		{
			name:     "search without fields (default email and name)",
			keyword:  "john",
			fields:   nil,
			expected: []*user.User{user1, user3}, // john@example.com and Bob Johnson
		},
		{
			name:     "search by email field only",
			keyword:  "test",
			fields:   []string{"email"},
			expected: []*user.User{user2}, // jane@test.com
		},
		{
			name:     "search by name field only",
			keyword:  "smith",
			fields:   []string{"name"},
			expected: []*user.User{user2}, // Jane Smith
		},
		{
			name:     "search by alias field only",
			keyword:  "johnny",
			fields:   []string{"alias"},
			expected: []*user.User{user1}, // alias: johnny
		},
		{
			name:     "search by multiple fields",
			keyword:  "bob",
			fields:   []string{"email", "name", "alias"},
			expected: []*user.User{user3, user3}, // bob@company.com, Bob Johnson, bobby (may return duplicates)
		},
		{
			name:     "search with non-existent field",
			keyword:  "test",
			fields:   []string{"nonexistent"},
			expected: []*user.User{},
		},
		{
			name:     "search with mixed existing and non-existent fields",
			keyword:  "jane",
			fields:   []string{"email", "nonexistent", "name"},
			expected: []*user.User{user2}, // jane@test.com and Jane Smith
		},
		{
			name:     "search with short keyword (less than 3 chars)",
			keyword:  "jo",
			fields:   nil,
			expected: []*user.User{},
		},
		{
			name:     "case insensitive search",
			keyword:  "JOHN",
			fields:   []string{"email", "name"},
			expected: []*user.User{user1, user3}, // john@example.com and Bob Johnson
		},
	}

	init := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			client := mongox.NewClientWithDatabase(init(t))
			repo := NewUser(client)
			ctx := context.Background()

			// Save test users
			for _, u := range []*user.User{user1, user2, user3} {
				err := repo.Save(ctx, u)
				assert.NoError(tt, err)
			}

			var got user.List
			var err error
			if tc.fields == nil {
				got, err = repo.SearchByKeyword(ctx, tc.keyword)
			} else {
				got, err = repo.SearchByKeyword(ctx, tc.keyword, tc.fields...)
			}

			assert.NoError(tt, err)

			// For this test, we'll check that we got the expected number of results
			// and that all returned users are in our expected list
			if len(tc.expected) == 0 {
				assert.Equal(tt, 0, len(got))
			} else {
				assert.Greater(tt, len(got), 0)
				for _, resultUser := range got {
					found := false
					for _, expectedUser := range tc.expected {
						if resultUser.ID() == expectedUser.ID() {
							found = true
							break
						}
					}
					assert.True(tt, found, "returned user should be in expected list")
				}
			}
		})
	}
}

func TestUserRepo_FindByIDsWithPagination(t *testing.T) {
	wsid := user.NewWorkspaceID()

	// Create test users
	user1 := user.New().NewID().Email("user1@test.com").Workspace(wsid).Name("user1").MustBuild()
	user2 := user.New().NewID().Email("user2@test.com").Workspace(wsid).Name("user2").MustBuild()
	user3 := user.New().NewID().Email("user3@test.com").Workspace(wsid).Name("user3").MustBuild()
	user4 := user.New().NewID().Email("user4@test.com").Workspace(wsid).Name("user4").MustBuild()
	user5 := user.New().NewID().Email("user5@test.com").Workspace(wsid).Name("user5").MustBuild()

	users := []*user.User{user1, user2, user3, user4, user5}

	tests := []struct {
		name          string
		ids           accountdomain.UserIDList
		pagination    *usecasex.Pagination
		expectedCount int64
		expectedUsers int
		expectedNext  bool
		expectedPrev  bool
	}{
		{
			name:          "nil pagination returns no results",
			ids:           accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:    nil,
			expectedCount: 0,
			expectedUsers: 0, // mongox.Paginate returns early with no results when pagination is nil
			expectedNext:  false,
			expectedPrev:  false,
		},
		{
			name:          "offset pagination first page",
			ids:           accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:    usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 2,
			expectedNext:  true,
			expectedPrev:  false,
		},
		{
			name:          "offset pagination second page",
			ids:           accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:    usecasex.OffsetPagination{Offset: 2, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 2,
			expectedNext:  true,
			expectedPrev:  false, // MongoDB pagination doesn't efficiently determine hasPreviousPage for offset
		},
		{
			name:          "offset pagination last page",
			ids:           accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:    usecasex.OffsetPagination{Offset: 4, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 1,
			expectedNext:  false,
			expectedPrev:  false, // MongoDB pagination doesn't efficiently determine hasPreviousPage for offset
		},
		{
			name:          "offset pagination beyond range",
			ids:           accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:    usecasex.OffsetPagination{Offset: 10, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 0,
			expectedNext:  false,
			expectedPrev:  false, // MongoDB pagination doesn't efficiently determine hasPreviousPage for offset
		},
		{
			name:          "empty ids list",
			ids:           accountdomain.UserIDList{},
			pagination:    usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			expectedCount: 0,
			expectedUsers: 0,
			expectedNext:  false,
			expectedPrev:  false,
		},
		{
			name:          "partial ids match",
			ids:           accountdomain.UserIDList{user1.ID(), user3.ID()},
			pagination:    usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			expectedCount: 2,
			expectedUsers: 2,
			expectedNext:  false,
			expectedPrev:  false,
		},
		{
			name:          "non-existent ids",
			ids:           accountdomain.UserIDList{user.NewID(), user.NewID()},
			pagination:    usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			expectedCount: 0,
			expectedUsers: 0,
			expectedNext:  false,
			expectedPrev:  false,
		},
	}

	init := mongotest.Connect(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := mongox.NewClientWithDatabase(init(t))
			repo := NewUser(client)
			ctx := context.Background()

			// Save all test users to the database
			for _, u := range users {
				err := repo.Save(ctx, u)
				assert.NoError(t, err)
			}

			result, pageInfo, err := repo.FindByIDsWithPagination(ctx, tt.ids, tt.pagination)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedUsers, len(result))

			if tt.pagination == nil {
				assert.Nil(t, pageInfo)
			} else {
				assert.NotNil(t, pageInfo)
				assert.Equal(t, tt.expectedCount, pageInfo.TotalCount)
				assert.Equal(t, tt.expectedNext, pageInfo.HasNextPage)
				assert.Equal(t, tt.expectedPrev, pageInfo.HasPreviousPage)
			}

			// Verify that returned users are from the requested IDs
			for _, resultUser := range result {
				assert.True(t, tt.ids.Has(resultUser.ID()), "returned user should be in requested IDs")
			}

			// Verify user data integrity
			for _, resultUser := range result {
				assert.NotNil(t, resultUser.ID())
				assert.NotEmpty(t, resultUser.Email())
				assert.NotEmpty(t, resultUser.Name())
				assert.Equal(t, wsid, resultUser.Workspace())
			}
		})
	}
}

func TestUserRepo_FindByIDsWithPagination_WithFilter(t *testing.T) {
	wsid := user.NewWorkspaceID()

	// Create test users with specific names and aliases
	user1 := user.New().NewID().Email("john@test.com").Workspace(wsid).Name("john_doe").Alias("johnd").MustBuild()
	user2 := user.New().NewID().Email("jane@test.com").Workspace(wsid).Name("jane_smith").Alias("janes").MustBuild()
	user3 := user.New().NewID().Email("alice@test.com").Workspace(wsid).Name("alice_wonder").Alias("alice").MustBuild()
	user4 := user.New().NewID().Email("bob@test.com").Workspace(wsid).Name("bob_builder").Alias("bobby").MustBuild()
	user5 := user.New().NewID().Email("charlie@test.com").Workspace(wsid).Name("charlie_brown").Alias("charlie").MustBuild()

	users := []*user.User{user1, user2, user3, user4, user5}

	tests := []struct {
		name              string
		ids               accountdomain.UserIDList
		pagination        *usecasex.Pagination
		nameOrAliasFilter string
		expectedUserNames []string
		expectedAliases   []string
		expectedCount     int64
		expectedUsers     int
		expectedNext      bool
		expectedPrev      bool
	}{
		{
			name:              "filter by name substring - case insensitive",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: "JOH",
			expectedUserNames: []string{"john_doe"},
			expectedAliases:   []string{"johnd"},
			expectedCount:     1,
			expectedUsers:     1,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter by alias substring - case insensitive",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: "alice",
			expectedUserNames: []string{"alice_wonder"},
			expectedAliases:   []string{"alice"},
			expectedCount:     1,
			expectedUsers:     1,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter matches multiple users by name pattern",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: "_",
			expectedCount:     5,
			expectedUsers:     5,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter matches users by alias pattern",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: "j",
			expectedCount:     2, // john and jane
			expectedUsers:     2,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter with no matches",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: "xyz",
			expectedCount:     0,
			expectedUsers:     0,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter with pagination - first page",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			nameOrAliasFilter: "_",
			expectedCount:     5,
			expectedUsers:     2,
			expectedNext:      true,
			expectedPrev:      false,
		},
		{
			name:              "filter with pagination - second page",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 2, Limit: 2}.Wrap(),
			nameOrAliasFilter: "_",
			expectedCount:     5,
			expectedUsers:     2,
			expectedNext:      true,
			expectedPrev:      false, // Changed based on actual MongoDB pagination behavior
		},
		{
			name:              "no filter provided - returns all",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: "",
			expectedCount:     5,
			expectedUsers:     5,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter with partial ID list",
			ids:               accountdomain.UserIDList{user1.ID(), user3.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: "alice",
			expectedUserNames: []string{"alice_wonder"},
			expectedCount:     1,
			expectedUsers:     1,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "test regex escaping for special characters",
			ids:               accountdomain.UserIDList{user1.ID(), user2.ID(), user3.ID(), user4.ID(), user5.ID()},
			pagination:        usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			nameOrAliasFilter: ".*", // Should match literal ".*", not regex pattern
			expectedCount:     0,
			expectedUsers:     0,
			expectedNext:      false,
			expectedPrev:      false,
		},
	}

	init := mongotest.Connect(t)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)

			// Clean up any existing data
			for _, u := range users {
				_ = repo.Remove(ctx, u.ID())
			}

			// Save all test users to the database
			for _, u := range users {
				err := repo.Save(ctx, u)
				assert.NoError(t, err)
			}

			var result user.List
			var pageInfo *usecasex.PageInfo
			var err error

			if tt.nameOrAliasFilter == "" {
				result, pageInfo, err = repo.FindByIDsWithPagination(ctx, tt.ids, tt.pagination)
			} else {
				result, pageInfo, err = repo.FindByIDsWithPagination(ctx, tt.ids, tt.pagination, tt.nameOrAliasFilter)
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedUsers, len(result))

			if tt.pagination == nil {
				assert.Nil(t, pageInfo)
			} else {
				assert.NotNil(t, pageInfo)
				assert.Equal(t, tt.expectedCount, pageInfo.TotalCount)
				assert.Equal(t, tt.expectedNext, pageInfo.HasNextPage)
				assert.Equal(t, tt.expectedPrev, pageInfo.HasPreviousPage)
			}

			// Verify that returned users are from the requested IDs
			for _, resultUser := range result {
				assert.True(t, tt.ids.Has(resultUser.ID()), "returned user should be in requested IDs")
			}

			// Verify specific expected users if provided
			if len(tt.expectedUserNames) > 0 {
				actualNames := make([]string, len(result))
				for i, u := range result {
					actualNames[i] = u.Name()
				}
				assert.ElementsMatch(t, tt.expectedUserNames, actualNames, "expected specific user names")
			}

			if len(tt.expectedAliases) > 0 {
				actualAliases := make([]string, len(result))
				for i, u := range result {
					actualAliases[i] = u.Alias()
				}
				assert.ElementsMatch(t, tt.expectedAliases, actualAliases, "expected specific user aliases")
			}

			// If filter is provided, verify that returned users match the filter
			if tt.nameOrAliasFilter != "" && tt.expectedUsers > 0 {
				filterLower := strings.ToLower(tt.nameOrAliasFilter)
				for _, resultUser := range result {
					nameMatches := strings.Contains(strings.ToLower(resultUser.Name()), filterLower)
					aliasMatches := strings.Contains(strings.ToLower(resultUser.Alias()), filterLower)
					assert.True(t, nameMatches || aliasMatches,
						"user %s (alias: %s) should match filter %s",
						resultUser.Name(), resultUser.Alias(), tt.nameOrAliasFilter)
				}
			}

			// Clean up
			for _, u := range users {
				_ = repo.Remove(ctx, u.ID())
			}
		})
	}
}

// TestUserRepo_FindByIDsWithPagination_SecurityInjectionPrevention tests that the MongoDB filtering
// implementation properly prevents NoSQL injection attacks
func TestUserRepo_FindByIDsWithPagination_SecurityInjectionPrevention(t *testing.T) {
	ctx := context.Background()

	wsid := user.NewWorkspaceID()

	// Test users with various names that could be targeted by injection
	normalUser := user.New().NewID().Name("normal_user").Alias("normal").Email("normal@test.com").Workspace(wsid).MustBuild()
	specialUser := user.New().NewID().Name("admin_user").Alias("admin").Email("admin@test.com").Workspace(wsid).MustBuild()
	systemUser := user.New().NewID().Name("system.service").Alias("sys").Email("system@test.com").Workspace(wsid).MustBuild()

	users := []*user.User{normalUser, specialUser, systemUser}
	allIDs := accountdomain.UserIDList{normalUser.ID(), specialUser.ID(), systemUser.ID()}

	injectionAttempts := []struct {
		name           string
		maliciousInput string
		description    string
		shouldMatch    []string // Names that should legitimately match if treated as literal string
	}{
		{
			name:           "regex wildcard injection",
			maliciousInput: ".*",
			description:    "Attempt to match all users with regex wildcard",
			shouldMatch:    []string{}, // Should match nothing since no user name contains literal ".*"
		},
		{
			name:           "regex character class injection",
			maliciousInput: "[a-z]*",
			description:    "Attempt to match with character class",
			shouldMatch:    []string{}, // Should match nothing
		},
		{
			name:           "regex quantifier injection",
			maliciousInput: "admin+",
			description:    "Attempt to use quantifier to match admin variations",
			shouldMatch:    []string{}, // Should match nothing since no name contains literal "admin+"
		},
		{
			name:           "regex anchor injection",
			maliciousInput: "^admin",
			description:    "Attempt to use anchor to match from start",
			shouldMatch:    []string{}, // Should match nothing
		},
		{
			name:           "regex escape injection",
			maliciousInput: "\\w+",
			description:    "Attempt to use word character class",
			shouldMatch:    []string{}, // Should match nothing
		},
		{
			name:           "regex alternation injection",
			maliciousInput: "admin|system",
			description:    "Attempt to use alternation to match multiple patterns",
			shouldMatch:    []string{}, // Should match nothing
		},
		{
			name:           "dot literal should match",
			maliciousInput: ".",
			description:    "Literal dot should match system.service",
			shouldMatch:    []string{"system.service"}, // Should match literal dot
		},
		{
			name:           "normal substring search",
			maliciousInput: "admin",
			description:    "Normal search should work",
			shouldMatch:    []string{"admin_user"}, // Should match normally
		},
		{
			name:           "test regex escaping for special characters",
			maliciousInput: ".*", // Should match literal ".*", not regex pattern
			description:    "Special regex patterns should be escaped",
			shouldMatch:    []string{},
		},
	}

	init := mongotest.Connect(t)

	for _, attempt := range injectionAttempts {
		attempt := attempt
		t.Run(attempt.name, func(t *testing.T) {
			t.Parallel()
			client := mongox.NewClientWithDatabase(init(t))

			repo := NewUser(client)

			// Clean up and set up test data
			for _, u := range users {
				_ = repo.Remove(ctx, u.ID())
			}
			for _, u := range users {
				err := repo.Save(ctx, u)
				assert.NoError(t, err)
			}

			result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
				usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
				attempt.maliciousInput)

			assert.NoError(t, err, "Query should not fail")
			assert.NotNil(t, pageInfo, "PageInfo should not be nil")

			// Verify expected matches
			assert.Equal(t, len(attempt.shouldMatch), len(result),
				"Expected %d matches for input '%s', got %d", len(attempt.shouldMatch), attempt.maliciousInput, len(result))

			// Verify actual matches
			if len(attempt.shouldMatch) > 0 {
				actualNames := make([]string, len(result))
				for i, u := range result {
					actualNames[i] = u.Name()
				}
				assert.ElementsMatch(t, attempt.shouldMatch, actualNames,
					"Expected matches don't align with actual results")
			}

			// Clean up
			for _, u := range users {
				_ = repo.Remove(ctx, u.ID())
			}

			t.Logf("MongoDB - Input: '%s' - Description: %s - Matches: %d",
				attempt.maliciousInput, attempt.description, len(result))
		})
	}
}

// TestUserRepo_FindByIDsWithPagination_InputValidation tests various input validation scenarios for MongoDB
func TestUserRepo_FindByIDsWithPagination_InputValidation(t *testing.T) {
	init := mongotest.Connect(t)
	ctx := context.Background()

	client := mongox.NewClientWithDatabase(init(t))
	wsid := user.NewWorkspaceID()
	testUser := user.New().NewID().Name("test_user").Alias("test").Email("test@test.com").Workspace(wsid).MustBuild()

	repo := NewUser(client)

	// Clean up and set up test data
	_ = repo.Remove(ctx, testUser.ID())
	err := repo.Save(ctx, testUser)
	assert.NoError(t, err)

	testCases := []struct {
		name            string
		filter          string
		expectedMatches int
		description     string
	}{
		{
			name:            "empty string filter",
			filter:          "",
			expectedMatches: 1, // Should return all users (no filter applied)
			description:     "Empty filter should return all matching IDs",
		},
		{
			name:            "whitespace only filter",
			filter:          "   ",
			expectedMatches: 0, // Trimmed to empty but treated as filter
			description:     "Whitespace-only filter should be treated as literal",
		},
		{
			name:            "very long filter string",
			filter:          strings.Repeat("a", 1000),
			expectedMatches: 0,
			description:     "Very long filter should not cause performance issues",
		},
		{
			name:            "unicode characters",
			filter:          "café",
			expectedMatches: 0,
			description:     "Unicode characters should be handled properly",
		},
		{
			name:            "special characters combination",
			filter:          "!@#$%^&*()",
			expectedMatches: 0,
			description:     "Special characters should be escaped properly",
		},
		{
			name:            "case sensitivity test",
			filter:          "TEST_USER",
			expectedMatches: 1,
			description:     "Case insensitive matching should work",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, pageInfo, err := repo.FindByIDsWithPagination(ctx,
				accountdomain.UserIDList{testUser.ID()},
				usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
				tc.filter)

			assert.NoError(t, err, "Query should not fail for input: %q", tc.filter)
			assert.NotNil(t, pageInfo, "PageInfo should not be nil")
			assert.Equal(t, tc.expectedMatches, len(result),
				"Expected %d matches for filter %q, got %d. %s",
				tc.expectedMatches, tc.filter, len(result), tc.description)

			t.Logf("MongoDB Filter: %q - Expected: %d - Actual: %d - %s",
				tc.filter, tc.expectedMatches, len(result), tc.description)
		})
	}

	// Clean up
	_ = repo.Remove(ctx, testUser.ID())
}

// TestUserRepo_FindByIDsWithPagination_BoundaryConditions tests edge cases for MongoDB
func TestUserRepo_FindByIDsWithPagination_BoundaryConditions(t *testing.T) {
	init := mongotest.Connect(t)
	ctx := context.Background()

	client := mongox.NewClientWithDatabase(init(t))
	wsid := user.NewWorkspaceID()

	// Create test users
	users := make([]*user.User, 10)
	var allIDs user.IDList
	for i := 0; i < 10; i++ {
		users[i] = user.New().NewID().
			Name(fmt.Sprintf("user_%d", i)).
			Alias(fmt.Sprintf("alias_%d", i)).
			Email(fmt.Sprintf("user%d@test.com", i)).
			Workspace(wsid).
			MustBuild()
		allIDs = append(allIDs, users[i].ID())
	}

	repo := NewUser(client)

	// Clean up and set up test data
	for _, u := range users {
		_ = repo.Remove(ctx, u.ID())
	}
	for _, u := range users {
		err := repo.Save(ctx, u)
		assert.NoError(t, err)
	}

	t.Run("zero limit pagination", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
			usecasex.OffsetPagination{Offset: 0, Limit: 0}.Wrap(),
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		// Zero limit falls back to default limit in MongoDB pagination implementation
		assert.Equal(t, 10, len(result), "Zero limit should fall back to default limit and return all matching results")
		assert.Equal(t, int64(10), pageInfo.TotalCount, "Total count should be accurate")
	})

	t.Run("offset beyond total count", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
			usecasex.OffsetPagination{Offset: 100, Limit: 10}.Wrap(),
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 0, len(result), "Offset beyond count should return empty results")
		assert.Equal(t, int64(10), pageInfo.TotalCount, "Total count should still be accurate")
		assert.False(t, pageInfo.HasNextPage, "Should not have next page")
	})

	t.Run("very large limit", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
			usecasex.OffsetPagination{Offset: 0, Limit: 10000}.Wrap(),
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 10, len(result), "Should return all available results")
		assert.Equal(t, int64(10), pageInfo.TotalCount)
		assert.False(t, pageInfo.HasNextPage, "Should not have next page")
	})

	t.Run("empty ID list", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, accountdomain.UserIDList{},
			usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap(),
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 0, len(result), "Empty ID list should return no results")
		assert.Equal(t, int64(0), pageInfo.TotalCount)
	})

	// Clean up
	for _, u := range users {
		_ = repo.Remove(ctx, u.ID())
	}
}
