package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func Test_AssetRepo_Filtered(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	id1 := id.NewAssetID()
	id2 := id.NewAssetID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().
		ID(id1).
		Project(pid1).
		CreatedByUser(uid1).
		Size(1000).
		Thread(id.NewThreadID().Ref()).
		ArchiveExtractionStatus(s).
		NewUUID().
		MustBuild()
	a2 := asset.New().
		ID(id2).
		Project(pid1).
		CreatedByUser(uid1).
		Size(1000).
		Thread(id.NewThreadID().Ref()).
		ArchiveExtractionStatus(s).
		NewUUID().
		MustBuild()

	tests := []struct {
		name    string
		seeds   asset.List
		arg     repo.ProjectFilter
		wantErr error
	}{
		{
			name: "no r/w workspaces operation denied",
			seeds: asset.List{
				a1,
				a2,
			},
			arg: repo.ProjectFilter{
				Readable: []id.ProjectID{},
				Writable: []id.ProjectID{},
			},
			wantErr: repo.ErrOperationDenied,
		},
		{
			name: "r/w workspaces operation success",
			seeds: asset.List{
				a1,
				a2,
			},
			arg: repo.ProjectFilter{
				Readable: []id.ProjectID{pid1},
				Writable: []id.ProjectID{pid1},
			},
			wantErr: nil,
		},
	}

	initDB := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := mongox.NewClientWithDatabase(initDB(t))

			r := NewAsset(client).Filtered(tc.arg)
			ctx := context.Background()
			for _, p := range tc.seeds {
				err := r.Save(ctx, p)
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestAssetRepo_FindByID(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	id1 := id.NewAssetID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	tim, _ := time.Parse(time.RFC3339, "2021-03-16T04:19:57.592Z")
	a1 := asset.New().ID(id1).Project(pid1).CreatedAt(tim).CreatedByUser(uid1).Size(1000).
		Thread(id.NewThreadID().Ref()).ArchiveExtractionStatus(s).NewUUID().MustBuild()
	a1f := a1.Clone()
	// a1f.SetFile(asset.NewFile().Name("aaa.txt").Path("/aaa.txt").Size(100).Build())

	tests := []struct {
		name    string
		seeds   asset.List
		arg     id.AssetID
		want    *asset.Asset
		wantErr error
	}{
		{
			name:    "Not found in empty db",
			seeds:   asset.List{},
			arg:     id.NewAssetID(),
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name: "Not found",
			seeds: asset.List{
				asset.New().
					ID(id1).
					Project(pid1).
					CreatedByUser(uid1).
					Size(1000).
					Thread(id.NewThreadID().Ref()).
					ArchiveExtractionStatus(s).
					NewUUID().
					MustBuild(),
			},
			arg:     id.NewAssetID(),
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name: "Found 1",
			seeds: asset.List{
				a1f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     id1,
			want:    a1,
			wantErr: nil,
		},
	}

	initDB := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := mongox.NewClientWithDatabase(initDB(t))

			r := NewAsset(client)
			ctx := context.Background()
			for _, p := range tc.seeds {
				err := r.Save(ctx, p)
				assert.NoError(t, err)
			}

			got, err := r.FindByID(ctx, tc.arg)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAssetRepo_FindByIDs(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	id1 := id.NewAssetID()
	id2 := id.NewAssetID()
	tim, _ := time.Parse(time.RFC3339, "2021-03-16T04:19:57.592Z")
	// c := asset.NewFile().Path("/").Build()
	// f := asset.NewFile().Path("/").Children([]*asset.File{c}).Build()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().ID(id1).Project(pid1).CreatedAt(tim).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	a2 := asset.New().ID(id2).Project(pid1).CreatedAt(tim).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	a1f, a2f := a1.Clone(), a2.Clone()
	// a1f.SetFile(f)
	// a2f.SetFile(f)

	tests := []struct {
		name    string
		seeds   asset.List
		arg     id.AssetIDList
		want    asset.List
		wantErr error
	}{
		{
			name:    "0 count in empty db",
			seeds:   asset.List{},
			arg:     []id.AssetID{},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "0 count with asset for another workspaces",
			seeds: asset.List{
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     []id.AssetID{},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "1 count with single asset",
			seeds: asset.List{
				a1f,
			},
			arg:     []id.AssetID{id1},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "1 count with multi assets",
			seeds: asset.List{
				a1f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     []id.AssetID{id1},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "2 count with multi assets",
			seeds: asset.List{
				a1f,
				a2f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     []id.AssetID{id1, id2},
			want:    asset.List{a1, a2},
			wantErr: nil,
		},
	}

	initDB := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := mongox.NewClientWithDatabase(initDB(t))

			r := NewAsset(client)
			ctx := context.Background()
			for _, a := range tc.seeds {
				err := r.Save(ctx, a)
				assert.NoError(t, err)
			}

			got, err := r.FindByIDs(ctx, tc.arg)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAssetRepo_Search(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	tim, _ := time.Parse(time.RFC3339, "2021-03-16T04:19:57.592Z")

	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().NewID().Project(pid1).CreatedAt(tim).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	a2 := asset.New().NewID().Project(pid1).CreatedAt(tim).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	a1f, a2f := a1.Clone(), a2.Clone()

	type args struct {
		tid          id.ProjectID
		pInfo        *usecasex.Pagination
		keyword      *string
		contentTypes []string
	}
	tests := []struct {
		name    string
		seeds   asset.List
		args    args
		filter  *repo.ProjectFilter
		want    asset.List
		wantErr error
	}{
		{
			name:    "0 count in empty db",
			seeds:   asset.List{},
			args:    args{tid: id.NewProjectID(), pInfo: nil},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "0 count with asset for another projects",
			seeds: asset.List{
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args:    args{tid: id.NewProjectID(), pInfo: nil},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "1 count with single asset",
			seeds: asset.List{
				a1,
			},
			args: args{
				tid:   pid1,
				pInfo: usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap(),
			},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "1 count with multi assets",
			seeds: asset.List{
				a1f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args: args{
				tid:   pid1,
				pInfo: usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap(),
			},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "2 count with multi assets",
			seeds: asset.List{
				a1f,
				a2f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args: args{
				tid:   pid1,
				pInfo: usecasex.CursorPagination{First: lo.ToPtr(int64(2))}.Wrap(),
			},
			want:    asset.List{a1, a2},
			wantErr: nil,
		},
		{
			name: "get 1st page of 2",
			seeds: asset.List{
				a1f,
				a2f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args: args{
				tid:   pid1,
				pInfo: usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap(),
			},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "get last page of 2",
			seeds: asset.List{
				a1f,
				a2f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args: args{
				tid:   pid1,
				pInfo: usecasex.CursorPagination{Last: lo.ToPtr(int64(1))}.Wrap(),
			},
			want:    asset.List{a2},
			wantErr: nil,
		},
		{
			name: "project filter operation success",
			seeds: asset.List{
				a1f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args: args{
				tid:   pid1,
				pInfo: usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap(),
			},
			filter: &repo.ProjectFilter{
				Readable: []id.ProjectID{pid1},
				Writable: []id.ProjectID{pid1},
			},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "project filter operation denied",
			seeds: asset.List{
				a1f,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args: args{
				tid:   pid1,
				pInfo: usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap(),
			},
			filter:  &repo.ProjectFilter{Readable: []id.ProjectID{}, Writable: []id.ProjectID{}},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "success content type filter",
			seeds: asset.List{
				a1f,
			},
			args: args{
				tid:          pid1,
				pInfo:        usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap(),
				contentTypes: []string{"application/json"},
			},
			want:    nil, // currently asset file data is inside the file object in the mongodoc / not a part of main Asset data
			wantErr: nil,
		},
	}

	initDB := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := mongox.NewClientWithDatabase(initDB(t))

			r := NewAsset(client)
			ctx := context.Background()
			for _, a := range tc.seeds {
				err := r.Save(ctx, a)
				assert.NoError(t, err)
			}

			if tc.filter != nil {
				r = r.Filtered(*tc.filter)
			}

			got, _, err := r.Search(ctx, tc.args.tid, repo.AssetFilter{
				Pagination:   tc.args.pInfo,
				Keyword:      tc.args.keyword,
				ContentTypes: tc.args.contentTypes,
			})
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAssetRepo_Delete(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	id1 := id.NewAssetID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().ID(id1).Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	tests := []struct {
		name  string
		seeds asset.List
		arg   id.AssetID

		wantErr error
	}{
		{
			name:    "Not found in empty db",
			seeds:   asset.List{},
			arg:     id.NewAssetID(),
			wantErr: rerror.ErrNotFound,
		},
		{
			name: "Not found",
			seeds: asset.List{
				asset.New().NewID().Project(pid1).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     id.NewAssetID(),
			wantErr: rerror.ErrNotFound,
		},
		{
			name: "Found 1",
			seeds: asset.List{
				a1,
			},
			arg:     id1,
			wantErr: nil,
		},
		{
			name: "Found 2",
			seeds: asset.List{
				a1,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     id1,
			wantErr: nil,
		},
	}

	initDB := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := mongox.NewClientWithDatabase(initDB(t))

			r := NewAsset(client)
			ctx := context.Background()
			for _, p := range tc.seeds {
				err := r.Save(ctx, p)
				assert.NoError(t, err)
			}

			err := r.Delete(ctx, tc.arg)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, err)
			_, err = r.FindByID(ctx, tc.arg)
			assert.ErrorIs(t, err, rerror.ErrNotFound)
		})
	}
}

func TestAssetRepo_BatchDelete(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	id1 := id.NewAssetID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().ID(id1).Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()

	pid2 := id.NewProjectID()
	uid2 := accountdomain.NewUserID()
	id2 := id.NewAssetID()

	a2 := asset.New().ID(id2).Project(pid2).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid2).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	initDB := mongotest.Connect(t)
	type args struct {
		ids []id.AssetID
	}
	tests := []struct {
		name  string
		seeds asset.List
		args  args
		want  error
	}{
		{
			name:  "success",
			seeds: asset.List{a1, a2},
			args: args{
				ids: []id.AssetID{id1, id2},
			},
			want: nil,
		},
		{
			name:  "success partial delete",
			seeds: asset.List{a1, a2},
			args: args{
				ids: []id.AssetID{id1},
			},
			want: nil,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := mongox.NewClientWithDatabase(initDB(t))

			r := NewAsset(client)
			ctx := context.Background()
			for _, p := range tc.seeds {
				err := r.Save(ctx, p)
				assert.NoError(t, err)
			}

			err := r.BatchDelete(ctx, tc.args.ids)
			if tc.want != nil {
				assert.Error(t, err, tc.want)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, err)
			assetsFinal, err := r.FindByIDs(ctx, tc.args.ids)
			assert.ErrorIs(t, err, nil)
			assert.Nil(t, assetsFinal)
		})
	}
}

func TestAssetRepo_Save(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	id1 := id.NewAssetID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().ID(id1).Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	tests := []struct {
		name    string
		seeds   asset.List
		arg     *asset.Asset
		want    *asset.Asset
		wantErr error
	}{
		{
			name: "Saved",
			seeds: asset.List{
				a1,
			},
			arg:     a1,
			want:    a1,
			wantErr: nil,
		},
	}

	initDB := mongotest.Connect(t)

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel()

			client := mongox.NewClientWithDatabase(initDB(t))

			r := NewAsset(client)
			ctx := context.Background()
			for _, p := range tc.seeds {
				err := r.Save(ctx, p)
				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
					return
				} else {
					assert.NoError(t, err)
				}
			}

			err := r.Save(ctx, tc.arg)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
