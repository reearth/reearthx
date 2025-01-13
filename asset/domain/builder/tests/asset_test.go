package builder_test

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/domain/builder"
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/stretchr/testify/assert"
)

func TestAssetBuilder_Build(t *testing.T) {
	assetID := id.NewID()
	workspaceID := id.NewWorkspaceID()

	tests := []struct {
		name    string
		builder func() *builder.AssetBuilder
		want    *entity.Asset
		wantErr error
	}{
		{
			name: "success",
			builder: func() *builder.AssetBuilder {
				return builder.NewAssetBuilder().
					ID(assetID).
					Name("test.jpg").
					Size(1024).
					ContentType("image/jpeg").
					WorkspaceID(workspaceID).
					URL("https://example.com/test.jpg")
			},
			want: func() *entity.Asset {
				asset := entity.NewAsset(assetID, "test.jpg", 1024, "image/jpeg")
				asset.MoveToWorkspace(workspaceID)
				asset.UpdateMetadata("test.jpg", "https://example.com/test.jpg", "image/jpeg")
				return asset
			}(),
			wantErr: nil,
		},
		{
			name: "missing ID",
			builder: func() *builder.AssetBuilder {
				return builder.NewAssetBuilder().
					Name("test.jpg").
					Size(1024).
					ContentType("image/jpeg").
					WorkspaceID(workspaceID).
					URL("https://example.com/test.jpg")
			},
			wantErr: id.ErrInvalidID,
		},
		{
			name: "missing workspace ID",
			builder: func() *builder.AssetBuilder {
				return builder.NewAssetBuilder().
					ID(assetID).
					Name("test.jpg").
					Size(1024).
					ContentType("image/jpeg").
					URL("https://example.com/test.jpg")
			},
			wantErr: domain.ErrEmptyWorkspaceID,
		},
		{
			name: "missing URL",
			builder: func() *builder.AssetBuilder {
				return builder.NewAssetBuilder().
					ID(assetID).
					Name("test.jpg").
					Size(1024).
					ContentType("image/jpeg").
					WorkspaceID(workspaceID)
			},
			wantErr: domain.ErrEmptyURL,
		},
		{
			name: "invalid size",
			builder: func() *builder.AssetBuilder {
				return builder.NewAssetBuilder().
					ID(assetID).
					Name("test.jpg").
					Size(0).
					ContentType("image/jpeg").
					WorkspaceID(workspaceID).
					URL("https://example.com/test.jpg")
			},
			wantErr: domain.ErrEmptySize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.builder().Build()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			if tt.want != nil {
				assert.Equal(t, tt.want.ID(), got.ID())
				assert.Equal(t, tt.want.Name(), got.Name())
				assert.Equal(t, tt.want.Size(), got.Size())
				assert.Equal(t, tt.want.ContentType(), got.ContentType())
				assert.Equal(t, tt.want.URL(), got.URL())
				assert.Equal(t, tt.want.WorkspaceID(), got.WorkspaceID())
			}
		})
	}
}

func TestAssetBuilder_MustBuild(t *testing.T) {
	assetID := id.NewID()
	workspaceID := id.NewWorkspaceID()

	// Test successful build
	assert.NotPanics(t, func() {
		asset := builder.NewAssetBuilder().
			ID(assetID).
			Name("test.jpg").
			Size(1024).
			ContentType("image/jpeg").
			WorkspaceID(workspaceID).
			URL("https://example.com/test.jpg").
			MustBuild()
		assert.NotNil(t, asset)
	})

	// Test panic on invalid build
	assert.Panics(t, func() {
		_ = builder.NewAssetBuilder().MustBuild()
	})
}

func TestAssetBuilder_Setters(t *testing.T) {
	assetID := id.NewID()
	workspaceID := id.NewWorkspaceID()
	projectID := id.NewProjectID()
	groupID := id.NewGroupID()
	now := time.Now()

	b := builder.NewAssetBuilder().
		CreatedAt(now).
		ID(assetID).
		Name("test.jpg").
		Size(1024).
		ContentType("image/jpeg").
		WorkspaceID(workspaceID).
		ProjectID(projectID).
		GroupID(groupID).
		URL("https://example.com/test.jpg").
		Status(entity.StatusActive).
		Error("test error")

	asset, err := b.Build()
	assert.NoError(t, err)
	assert.NotNil(t, asset)

	assert.Equal(t, assetID, asset.ID())
	assert.Equal(t, workspaceID, asset.WorkspaceID())
	assert.Equal(t, projectID, asset.ProjectID())
	assert.Equal(t, groupID, asset.GroupID())
	assert.Equal(t, "test.jpg", asset.Name())
	assert.Equal(t, int64(1024), asset.Size())
	assert.Equal(t, "https://example.com/test.jpg", asset.URL())
	assert.Equal(t, "image/jpeg", asset.ContentType())
	assert.Equal(t, entity.StatusActive, asset.Status())
	assert.Equal(t, "test error", asset.Error())
	assert.Equal(t, now.Unix(), asset.CreatedAt().Unix())
}

func TestAssetBuilder_NewID(t *testing.T) {
	b := builder.NewAssetBuilder().NewID()
	// Add required fields to make the build succeed
	b = b.
		Name("test.jpg").
		Size(1024).
		ContentType("image/jpeg").
		WorkspaceID(id.NewWorkspaceID()).
		URL("https://example.com/test.jpg")

	asset, err := b.Build()
	assert.NoError(t, err)
	assert.NotNil(t, asset)
	assert.NotEqual(t, id.ID{}, asset.ID()) // ID should be set
}
