package accountmemory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/reearth/reearthx/util"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	expected := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}

	got := NewUser()
	assert.Equal(t, expected, got)
}

func TestNewUserWith(t *testing.T) {
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").Auths([]user.Auth{{
		Sub: "xxx",
	}}).MustBuild()

	got, err := NewUserWith(u).FindByID(context.Background(), u.ID())
	assert.NoError(t, err)
	assert.Equal(t, u, got)
}

func TestUser_FindAll(t *testing.T) {
	ctx := context.Background()
	u1 := user.New().NewID().Name("hoge").Email("abc@bb.cc").MustBuild()
	u2 := user.New().NewID().Name("foo").Email("cba@bb.cc").MustBuild()
	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}

	out, err := r.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(out))

	r.data.Store(u1.ID(), u1)
	r.data.Store(u2.ID(), u2)

	out, err = r.FindAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(out))

	wantErr := errors.New("test")
	SetUserError(r, wantErr)
	_, err = r.FindAll(ctx)
	assert.Same(t, wantErr, err)
}

func TestUser_FindBySub(t *testing.T) {
	ctx := context.Background()
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").Auths([]user.Auth{{
		Sub: "xxx",
	}}).MustBuild()

	tests := []struct {
		name     string
		auth0sub string
		want     *user.User
		wantErr  error
		mockErr  bool
	}{
		{
			name:     "must find user by auth",
			auth0sub: "xxx",
			want:     u,
		},
		{
			name:     "must return ErrInvalidParams",
			auth0sub: "",
			wantErr:  rerror.ErrInvalidParams,
		},
		{
			name:    "must mock error",
			wantErr: errors.New("test"),
			mockErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			r := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}
			r.data.Store(u.ID(), u)
			if tc.mockErr {
				SetUserError(r, tc.wantErr)
			}
			got, err := r.FindBySub(ctx, tc.auth0sub)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
			} else {
				assert.Equal(tt, tc.want, got)
			}
		})
	}
}

func TestUser_FindByEmail(t *testing.T) {
	ctx := context.Background()
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").MustBuild()
	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}
	r.data.Store(u.ID(), u)
	out, err := r.FindByEmail(ctx, "aa@bb.cc")
	assert.NoError(t, err)
	assert.Equal(t, u, out)

	out, err = r.FindByEmail(ctx, "abc@bb.cc")
	assert.Same(t, rerror.ErrNotFound, err)
	assert.Nil(t, out)

	wantErr := errors.New("test")
	SetUserError(r, wantErr)
	_, err = r.FindByEmail(ctx, "")
	assert.Same(t, wantErr, err)
}

func TestUser_FindByIDs(t *testing.T) {
	ctx := context.Background()
	u1 := user.New().NewID().Name("hoge").Email("abc@bb.cc").MustBuild()
	u2 := user.New().NewID().Name("foo").Email("cba@bb.cc").MustBuild()
	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}
	r.data.Store(u1.ID(), u1)
	r.data.Store(u2.ID(), u2)

	ids := accountdomain.UserIDList{
		u1.ID(),
		u2.ID(),
	}
	out, err := r.FindByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(out))

	wantErr := errors.New("test")
	SetUserError(r, wantErr)
	_, err = r.FindByIDs(ctx, ids)
	assert.Same(t, wantErr, err)
}

func TestUser_FindByName(t *testing.T) {
	ctx := context.Background()
	pr := user.PasswordReset{
		Token: "123abc",
	}
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").PasswordReset(pr.Clone()).MustBuild()

	tests := []struct {
		name    string
		seeds   []*user.User
		uName   string
		want    *user.User
		wantErr error
		mockErr bool
	}{
		{
			name:  "must find user by name",
			seeds: []*user.User{u},
			uName: "hoge",
			want:  u,
		},
		{
			name:    "must return ErrInvalidParams",
			wantErr: rerror.ErrInvalidParams,
		},
		{
			name:    "must return ErrNotFound",
			uName:   "xxx",
			wantErr: rerror.ErrNotFound,
		},
		{
			name:    "must mock error",
			wantErr: errors.New("test"),
			mockErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			r := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}
			r.data.Store(u.ID(), u)
			if tc.mockErr {
				SetUserError(r, tc.wantErr)
			}
			for _, u := range tc.seeds {
				_ = r.Save(ctx, u.Clone())
			}
			got, err := r.FindByName(ctx, tc.uName)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
			} else {
				assert.Equal(tt, tc.want, got)
			}
		})
	}
}

func TestUser_FindByNameOrEmail(t *testing.T) {
	ctx := context.Background()
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").MustBuild()
	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}
	r.data.Store(u.ID(), u)

	out, err := r.FindByNameOrEmail(ctx, "hoge")
	assert.NoError(t, err)
	assert.Equal(t, u, out)

	out2, err := r.FindByNameOrEmail(ctx, "aa@bb.cc")
	assert.NoError(t, err)
	assert.Equal(t, u, out2)

	out3, err := r.FindByNameOrEmail(ctx, "xxx")
	assert.Nil(t, out3)
	assert.Same(t, rerror.ErrNotFound, err)

	wantErr := errors.New("test")
	SetUserError(r, wantErr)
	_, err = r.FindByID(ctx, u.ID())
	assert.Same(t, wantErr, err)
}

func TestUser_FindByPasswordResetRequest(t *testing.T) {
	ctx := context.Background()
	pr := user.PasswordReset{
		Token: "123abc",
	}
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").PasswordReset(pr.Clone()).MustBuild()

	tests := []struct {
		name    string
		seeds   []*user.User
		token   string
		want    *user.User
		wantErr error
		mockErr bool
	}{
		{
			name:  "must find user by password reset",
			seeds: []*user.User{u},
			token: "123abc",
			want:  u,
		},
		{
			name:    "must return ErrInvalidParams",
			wantErr: rerror.ErrInvalidParams,
		},
		{
			name:    "must return ErrNotFound",
			token:   "xxx",
			wantErr: rerror.ErrNotFound,
		},
		{
			name:    "must mock error",
			wantErr: errors.New("test"),
			mockErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			r := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}
			if tc.mockErr {
				SetUserError(r, tc.wantErr)
			}
			for _, uu := range tc.seeds {
				_ = r.Save(ctx, uu.Clone())
			}
			got, err := r.FindByPasswordResetRequest(ctx, tc.token)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr, err)
			} else {
				assert.Equal(tt, tc.want, got)
			}
		})
	}
}

func TestUser_FindByVerification(t *testing.T) {
	ctx := context.Background()
	vr := user.VerificationFrom("123abc", time.Now(), false)
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").Verification(vr).MustBuild()

	tests := []struct {
		name    string
		seeds   []*user.User
		code    string
		want    *user.User
		wantErr error
		mockErr bool
	}{
		{
			name:    "must find user by verification",
			code:    "123abc",
			want:    u,
			wantErr: nil,
		},
		{
			name:    "must return ErrInvalidParams",
			wantErr: rerror.ErrInvalidParams,
		},
		{
			name:    "must return ErrNotFound",
			code:    "xxx",
			wantErr: rerror.ErrNotFound,
		},
		{
			name:    "must mock error",
			wantErr: errors.New("test"),
			mockErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			r := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}
			r.data.Store(u.ID(), u)
			if tc.mockErr {
				SetUserError(r, tc.wantErr)
			}
			for _, u := range tc.seeds {
				_ = r.Save(ctx, u.Clone())
			}
			got, err := r.FindByVerification(ctx, tc.code)
			if tc.wantErr != nil {
				assert.Equal(tt, tc.wantErr, err)
			} else {
				assert.Equal(tt, tc.want, got)
			}
		})
	}
}

func TestUser_FindByID(t *testing.T) {
	ctx := context.Background()
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").MustBuild()
	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}
	r.data.Store(u.ID(), u)

	out, err := r.FindByID(ctx, u.ID())
	assert.NoError(t, err)
	assert.Equal(t, u, out)

	out2, err := r.FindByID(ctx, accountdomain.UserID{})
	assert.Nil(t, out2)
	assert.Same(t, rerror.ErrNotFound, err)

	wantErr := errors.New("test")
	SetUserError(r, wantErr)
	_, err = r.FindByID(ctx, u.ID())
	assert.Same(t, wantErr, err)
}

func TestUser_FindBySubOrCreate(t *testing.T) {
	ctx := context.Background()
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").Auths([]user.Auth{{Sub: "auth0|aaa", Provider: "auth0"}}).MustBuild()

	r := &User{data: &util.SyncMap[accountdomain.UserID, *user.User]{}}

	_, err := r.FindBySubOrCreate(ctx, u, "auth0|aaa")
	assert.NoError(t, err)
	assert.Equal(t, 1, r.data.Len())

	// if same sub, it returns existing data in stead of inserting new data
	_, err = r.FindBySubOrCreate(ctx, u, "auth0|aaa")
	assert.NoError(t, err)
	assert.Equal(t, 1, r.data.Len())
}

func TestUser_Create(t *testing.T) {
	uid := accountdomain.NewUserID()
	ctx := context.Background()
	u := user.New().ID(uid).Name("hoge").Email("aa@bb.cc").Auths([]user.Auth{{Sub: "auth0|aaa", Provider: "auth0"}}).MustBuild()

	r := &User{data: &util.SyncMap[accountdomain.UserID, *user.User]{}}

	err := r.Create(ctx, u)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.data.Len())

	err = r.Create(ctx, u)
	assert.Equal(t, accountrepo.ErrDuplicatedUser, err)
}

func TestUser_Save(t *testing.T) {
	ctx := context.Background()
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").MustBuild()

	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}
	_ = r.Save(ctx, u)

	assert.Equal(t, 1, r.data.Len())

	wantErr := errors.New("test")
	SetUserError(r, wantErr)
	assert.Same(t, wantErr, r.Save(ctx, u))
}

func TestUser_Remove(t *testing.T) {
	ctx := context.Background()
	u := user.New().NewID().Name("hoge").Email("aa@bb.cc").MustBuild()
	u2 := user.New().NewID().Name("xxx").Email("abc@bb.cc").MustBuild()
	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}
	r.data.Store(u.ID(), u)
	r.data.Store(u2.ID(), u2)

	_ = r.Remove(ctx, u2.ID())
	assert.Equal(t, 1, r.data.Len())

	wantErr := errors.New("test")
	SetUserError(r, wantErr)
	assert.Same(t, wantErr, r.Remove(ctx, u.ID()))
}

func TestUser_FindByIDsWithPagination(t *testing.T) {
	ctx := context.Background()

	// Create test users
	u1 := user.New().NewID().Name("user1").Email("user1@test.com").MustBuild()
	u2 := user.New().NewID().Name("user2").Email("user2@test.com").MustBuild()
	u3 := user.New().NewID().Name("user3").Email("user3@test.com").MustBuild()
	u4 := user.New().NewID().Name("user4").Email("user4@test.com").MustBuild()
	u5 := user.New().NewID().Name("user5").Email("user5@test.com").MustBuild()

	users := []*user.User{u1, u2, u3, u4, u5}

	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}

	// Store all users
	for _, u := range users {
		r.data.Store(u.ID(), u)
	}

	ids := accountdomain.UserIDList{u1.ID(), u2.ID(), u3.ID(), u4.ID(), u5.ID()}

	tests := []struct {
		name          string
		ids           accountdomain.UserIDList
		pagination    *usecasex.Pagination
		expectedCount int64
		expectedUsers int
		expectedNext  bool
		expectedPrev  bool
		mockErr       bool
		wantErr       error
	}{
		{
			name:          "nil pagination returns all users",
			ids:           ids,
			pagination:    nil,
			expectedCount: 0,
			expectedUsers: 5,
			expectedNext:  false,
			expectedPrev:  false,
		},
		{
			name:          "offset pagination first page",
			ids:           ids,
			pagination:    usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 2,
			expectedNext:  true,
			expectedPrev:  false,
		},
		{
			name:          "offset pagination second page",
			ids:           ids,
			pagination:    usecasex.OffsetPagination{Offset: 2, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 2,
			expectedNext:  true,
			expectedPrev:  true,
		},
		{
			name:          "offset pagination last page",
			ids:           ids,
			pagination:    usecasex.OffsetPagination{Offset: 4, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 1,
			expectedNext:  false,
			expectedPrev:  true,
		},
		{
			name:          "offset pagination beyond range",
			ids:           ids,
			pagination:    usecasex.OffsetPagination{Offset: 10, Limit: 2}.Wrap(),
			expectedCount: 5,
			expectedUsers: 0,
			expectedNext:  false,
			expectedPrev:  true,
		},
		{
			name:          "cursor pagination with first",
			ids:           ids,
			pagination:    usecasex.CursorPagination{First: &[]int64{3}[0]}.Wrap(),
			expectedCount: 5,
			expectedUsers: 3,
			expectedNext:  true,
			expectedPrev:  false,
		},
		{
			name:          "cursor pagination with last",
			ids:           ids,
			pagination:    usecasex.CursorPagination{Last: &[]int64{2}[0]}.Wrap(),
			expectedCount: 5,
			expectedUsers: 2,
			expectedNext:  false,
			expectedPrev:  true,
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
			ids:           accountdomain.UserIDList{u1.ID(), u3.ID()},
			pagination:    usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			expectedCount: 2,
			expectedUsers: 2,
			expectedNext:  false,
			expectedPrev:  false,
		},
		{
			name:       "mock error",
			ids:        ids,
			pagination: usecasex.OffsetPagination{Offset: 0, Limit: 2}.Wrap(),
			mockErr:    true,
			wantErr:    errors.New("test error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh repository for each test
			testRepo := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}

			// Store all users
			for _, u := range users {
				testRepo.data.Store(u.ID(), u)
			}

			if tt.mockErr {
				SetUserError(testRepo, tt.wantErr)
			}

			result, pageInfo, err := testRepo.FindByIDsWithPagination(ctx, tt.ids, tt.pagination)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, result)
				assert.Nil(t, pageInfo)
				return
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
		})
	}
}

func TestUser_FindByIDsWithPagination_WithFilter(t *testing.T) {
	ctx := context.Background()

	// Create test users with specific names and aliases
	u1 := user.New().NewID().Name("john_doe").Alias("johnd").Email("john@test.com").MustBuild()
	u2 := user.New().NewID().Name("jane_smith").Alias("janes").Email("jane@test.com").MustBuild()
	u3 := user.New().NewID().Name("alice_wonder").Alias("alice").Email("alice@test.com").MustBuild()
	u4 := user.New().NewID().Name("bob_builder").Alias("bobby").Email("bob@test.com").MustBuild()
	u5 := user.New().NewID().Name("charlie_brown").Alias("charlie").Email("charlie@test.com").MustBuild()

	users := []*user.User{u1, u2, u3, u4, u5}

	r := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}

	// Store all users
	for _, u := range users {
		r.data.Store(u.ID(), u)
	}

	ids := accountdomain.UserIDList{u1.ID(), u2.ID(), u3.ID(), u4.ID(), u5.ID()}

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
		mockErr           bool
		wantErr           error
	}{
		{
			name:              "filter by name substring - case insensitive",
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
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
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
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
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			nameOrAliasFilter: "_",
			expectedCount:     5,
			expectedUsers:     5,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter matches users by alias pattern",
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			nameOrAliasFilter: "j",
			expectedCount:     2, // john and jane
			expectedUsers:     2,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter with no matches",
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			nameOrAliasFilter: "xyz",
			expectedCount:     0,
			expectedUsers:     0,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter with pagination - first page",
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 2}},
			nameOrAliasFilter: "_",
			expectedCount:     5,
			expectedUsers:     2,
			expectedNext:      true,
			expectedPrev:      false,
		},
		{
			name:              "filter with pagination - second page",
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 2, Limit: 2}},
			nameOrAliasFilter: "_",
			expectedCount:     5,
			expectedUsers:     2,
			expectedNext:      true,
			expectedPrev:      true,
		},
		{
			name:              "filter with nil pagination",
			ids:               ids,
			pagination:        nil,
			nameOrAliasFilter: "charlie",
			expectedUserNames: []string{"charlie_brown"},
			expectedAliases:   []string{"charlie"},
			expectedCount:     0, // nil pagination doesn't return count
			expectedUsers:     1,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "no filter provided - returns all",
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			nameOrAliasFilter: "",
			expectedCount:     5,
			expectedUsers:     5,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "filter with partial ID list",
			ids:               accountdomain.UserIDList{u1.ID(), u3.ID()},
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			nameOrAliasFilter: "alice",
			expectedUserNames: []string{"alice_wonder"},
			expectedCount:     1,
			expectedUsers:     1,
			expectedNext:      false,
			expectedPrev:      false,
		},
		{
			name:              "error handling",
			ids:               ids,
			pagination:        &usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			nameOrAliasFilter: "test",
			mockErr:           true,
			wantErr:           errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}

			// Store test users
			for _, u := range users {
				testRepo.data.Store(u.ID(), u)
			}

			if tt.mockErr {
				SetUserError(testRepo, tt.wantErr)
			}

			var result user.List
			var pageInfo *usecasex.PageInfo
			var err error

			if tt.nameOrAliasFilter == "" {
				result, pageInfo, err = testRepo.FindByIDsWithPagination(ctx, tt.ids, tt.pagination)
			} else {
				result, pageInfo, err = testRepo.FindByIDsWithPagination(ctx, tt.ids, tt.pagination, tt.nameOrAliasFilter)
			}

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, result)
				assert.Nil(t, pageInfo)
				return
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
		})
	}
}

// TestUser_FindByIDsWithPagination_SecurityInjectionPrevention tests that the filtering
// implementation properly prevents NoSQL injection attacks
func TestUser_FindByIDsWithPagination_SecurityInjectionPrevention(t *testing.T) {
	ctx := context.Background()

	// Test users with various names that could be targeted by injection
	normalUser := user.New().NewID().Name("normal_user").Alias("normal").Email("normal@test.com").MustBuild()
	specialUser := user.New().NewID().Name("admin_user").Alias("admin").Email("admin@test.com").MustBuild()
	systemUser := user.New().NewID().Name("system.service").Alias("sys").Email("system@test.com").MustBuild()

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
	}

	for _, attempt := range injectionAttempts {
		t.Run(attempt.name, func(t *testing.T) {
			repo := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}

			// Store test users
			for _, u := range users {
				repo.data.Store(u.ID(), u)
			}

			result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
				&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
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

			t.Logf("Input: '%s' - Description: %s - Matches: %d", attempt.maliciousInput, attempt.description, len(result))
		})
	}
}

// TestUser_FindByIDsWithPagination_InputValidation tests various input validation scenarios
func TestUser_FindByIDsWithPagination_InputValidation(t *testing.T) {
	ctx := context.Background()

	testUser := user.New().NewID().Name("test_user").Alias("test").Email("test@test.com").MustBuild()

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
			name:            "null character attempt",
			filter:          "test\x00user",
			expectedMatches: 0,
			description:     "Null characters should be handled safely",
		},
		{
			name:            "newline characters",
			filter:          "test\nuser",
			expectedMatches: 0,
			description:     "Newline characters should be treated literally",
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
			repo := &User{
				data: &util.SyncMap[accountdomain.UserID, *user.User]{},
			}
			repo.data.Store(testUser.ID(), testUser)

			result, pageInfo, err := repo.FindByIDsWithPagination(ctx,
				accountdomain.UserIDList{testUser.ID()},
				&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
				tc.filter)

			assert.NoError(t, err, "Query should not fail for input: %q", tc.filter)
			assert.NotNil(t, pageInfo, "PageInfo should not be nil")
			assert.Equal(t, tc.expectedMatches, len(result),
				"Expected %d matches for filter %q, got %d. %s",
				tc.expectedMatches, tc.filter, len(result), tc.description)

			t.Logf("Filter: %q - Expected: %d - Actual: %d - %s",
				tc.filter, tc.expectedMatches, len(result), tc.description)
		})
	}
}

// TestUser_FindByIDsWithPagination_BoundaryConditions tests edge cases and boundary conditions
func TestUser_FindByIDsWithPagination_BoundaryConditions(t *testing.T) {
	ctx := context.Background()

	// Create test users
	users := make([]*user.User, 10)
	var allIDs accountdomain.UserIDList
	for i := 0; i < 10; i++ {
		users[i] = user.New().NewID().
			Name(fmt.Sprintf("user_%d", i)).
			Alias(fmt.Sprintf("alias_%d", i)).
			Email(fmt.Sprintf("user%d@test.com", i)).
			MustBuild()
		allIDs = append(allIDs, users[i].ID())
	}

	repo := &User{
		data: &util.SyncMap[accountdomain.UserID, *user.User]{},
	}
	for _, u := range users {
		repo.data.Store(u.ID(), u)
	}

	t.Run("zero limit pagination", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
			&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 0}},
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 0, len(result), "Zero limit should return no results")
		assert.Equal(t, int64(10), pageInfo.TotalCount, "Total count should still be accurate")
	})

	t.Run("offset beyond total count", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
			&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 100, Limit: 10}},
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 0, len(result), "Offset beyond count should return empty results")
		assert.Equal(t, int64(10), pageInfo.TotalCount, "Total count should still be accurate")
		assert.False(t, pageInfo.HasNextPage, "Should not have next page")
	})

	t.Run("very large limit", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
			&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10000}},
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 10, len(result), "Should return all available results")
		assert.Equal(t, int64(10), pageInfo.TotalCount)
		assert.False(t, pageInfo.HasNextPage, "Should not have next page")
	})

	t.Run("empty ID list", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, accountdomain.UserIDList{},
			&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			"user")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 0, len(result), "Empty ID list should return no results")
		assert.Equal(t, int64(0), pageInfo.TotalCount)
	})

	t.Run("single character filter", func(t *testing.T) {
		result, pageInfo, err := repo.FindByIDsWithPagination(ctx, allIDs,
			&usecasex.Pagination{Offset: &usecasex.OffsetPagination{Offset: 0, Limit: 10}},
			"_")

		assert.NoError(t, err)
		assert.NotNil(t, pageInfo)
		assert.Equal(t, 10, len(result), "Single underscore should match all users")
		assert.Equal(t, int64(10), pageInfo.TotalCount)
	})
}
