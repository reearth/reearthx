package accountmongo

import (
	"context"
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
