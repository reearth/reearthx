package memory

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAssetRepo_Filtered(t *testing.T) {
	tid1 := id.NewProjectID()
	id1 := id.NewAssetID()
	id2 := id.NewAssetID()
	uid1 := accountdomain.NewUserID()
	uid2 := accountdomain.NewUserID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	p1 := asset.New().
		ID(id1).
		Project(tid1).
		CreatedByUser(uid1).
		Size(1000).
		Thread(id.NewThreadID().Ref()).
		ArchiveExtractionStatus(s).
		NewUUID().
		MustBuild()
	p2 := asset.New().
		ID(id2).
		Project(tid1).
		CreatedByUser(uid2).
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
		mockErr bool
	}{
		{
			name: "project filter operation denied",
			seeds: asset.List{
				p1,
				p2,
			},
			arg: repo.ProjectFilter{
				Readable: []id.ProjectID{},
				Writable: []id.ProjectID{},
			},
			wantErr: repo.ErrOperationDenied,
		},
		{
			name: "project filter operation success",
			seeds: asset.List{
				p1,
				p2,
			},
			arg: repo.ProjectFilter{
				Readable: []id.ProjectID{tid1},
				Writable: []id.ProjectID{tid1},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := NewAsset().Filtered(tc.arg)
			ctx := context.Background()
			for _, p := range tc.seeds {
				err := r.Save(ctx, p.Clone())
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
	a1 := asset.New().ID(id1).
		Project(pid1).
		CreatedByUser(uid1).
		Size(1000).
		Thread(id.NewThreadID().Ref()).
		NewUUID().
		MustBuild()
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
				asset.New().NewID().Project(pid1).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     id.NewAssetID(),
			want:    nil,
			wantErr: rerror.ErrNotFound,
		},
		{
			name: "Found 1",
			seeds: asset.List{
				a1,
			},
			arg:     id1,
			want:    a1,
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
			want:    a1,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := NewAsset()
			ctx := context.Background()

			for _, a := range tc.seeds {
				err := r.Save(ctx, a.Clone())
				assert.NoError(t, err)
			}

			got, err := r.FindByID(ctx, tc.arg)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
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
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().ID(id1).Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	a2 := asset.New().ID(id2).Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()

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
			arg:     id.AssetIDList{},
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
			arg:     id.AssetIDList{},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "1 count with single asset",
			seeds: asset.List{
				a1,
			},
			arg:     id.AssetIDList{id1},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "1 count with multi assets",
			seeds: asset.List{
				a1,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     id.AssetIDList{id1},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "2 count with multi assets",
			seeds: asset.List{
				a1,
				a2,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			arg:     id.AssetIDList{id1, id2},
			want:    asset.List{a1, a2},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := NewAsset()
			ctx := context.Background()
			for _, a := range tc.seeds {
				err := r.Save(ctx, a.Clone())
				assert.NoError(t, err)
			}

			got, err := r.FindByIDs(ctx, tc.arg)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAssetRepo_Search(t *testing.T) {
	pid1 := id.NewProjectID()
	uid1 := accountdomain.NewUserID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().NewID().Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	a2 := asset.New().NewID().Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()

	type args struct {
		pid   id.ProjectID
		pInfo *usecasex.Pagination
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
			args:    args{id.NewProjectID(), nil},
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
			args:    args{id.NewProjectID(), nil},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "1 count with single asset",
			seeds: asset.List{
				a1,
			},
			args:    args{pid1, usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap()},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "1 count with multi assets",
			seeds: asset.List{
				a1,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args:    args{pid1, usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap()},
			want:    asset.List{a1},
			wantErr: nil,
		},
		{
			name: "2 count with multi assets",
			seeds: asset.List{
				a1,
				a2,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args:    args{pid1, usecasex.CursorPagination{First: lo.ToPtr(int64(2))}.Wrap()},
			want:    asset.List{a1, a2},
			wantErr: nil,
		},
		{
			name: "get 1st page of 2",
			seeds: asset.List{
				a1,
				a2,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args:    args{pid1, usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap()},
			want:    asset.List{a1, a2},
			wantErr: nil,
		},
		{
			name: "project filter operation succeed",
			seeds: asset.List{
				a1,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args: args{pid1, usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap()},
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
				a1,
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
				asset.New().NewID().Project(id.NewProjectID()).ArchiveExtractionStatus(s).NewUUID().
					CreatedByUser(accountdomain.NewUserID()).
					Size(1000).Thread(id.NewThreadID().Ref()).MustBuild(),
			},
			args:    args{pid1, usecasex.CursorPagination{First: lo.ToPtr(int64(1))}.Wrap()},
			filter:  &repo.ProjectFilter{Readable: []id.ProjectID{}, Writable: []id.ProjectID{}},
			want:    nil,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := NewAsset()
			ctx := context.Background()
			for _, a := range tc.seeds {
				err := r.Save(ctx, a.Clone())
				assert.NoError(t, err)
			}

			if tc.filter != nil {
				r = r.Filtered(*tc.filter)
			}

			got, _, err := r.Search(ctx, tc.args.pid, repo.AssetFilter{})
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAssetRepo_Delete(t *testing.T) {
	pid1 := id.NewProjectID()
	id1 := id.NewAssetID()
	uid1 := accountdomain.NewUserID()
	s := lo.ToPtr(asset.ArchiveExtractionStatusPending)
	a1 := asset.New().NewID().Project(pid1).ArchiveExtractionStatus(s).NewUUID().
		CreatedByUser(uid1).Size(1000).Thread(id.NewThreadID().Ref()).MustBuild()
	tests := []struct {
		name    string
		seeds   asset.List
		arg     id.AssetID
		wantErr error
	}{
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

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := NewAsset()
			ctx := context.Background()
			for _, a := range tc.seeds {
				err := r.Save(ctx, a.Clone())
				assert.NoError(t, err)
			}

			err := r.Delete(ctx, tc.arg)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			_, err = r.FindByID(ctx, tc.arg)
			assert.ErrorIs(t, err, rerror.ErrNotFound)
		})
	}
}
