package accountmemory

import (
	"context"
	"errors"
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
