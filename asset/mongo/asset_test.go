package mongo

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset"
	"github.com/reearth/reearthx/asset/mongo/mongodoc"
	"github.com/reearth/reearthx/idx"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	mongotest.Env = "REEARTH_DB"
}

func TestAssetRepository_Init(t *testing.T) {
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	err := repo.Init()
	assert.NoError(t, err)
}

func TestAssetRepository_Save(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("test.txt")
	a.SetUUID("test-uuid")
	a.SetURL("http://example.com/test.txt")

	err := repo.Save(ctx, a)
	assert.NoError(t, err)

	found, err := repo.FindByID(ctx, assetID)
	assert.NoError(t, err)
	assert.Equal(t, assetID, found.ID())
	assert.Equal(t, groupID, *found.GroupID())
	assert.Equal(t, "test.txt", found.FileName())
	assert.Equal(t, "test-uuid", found.UUID())
	assert.Equal(t, "http://example.com/test.txt", found.URL())
}

func TestAssetRepository_SaveCMS(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("test-cms.txt")
	a.SetUUID("test-cms-uuid")

	err := repo.SaveCMS(ctx, a)
	assert.NoError(t, err)

	found, err := repo.FindByID(ctx, assetID)
	assert.NoError(t, err)
	assert.Equal(t, assetID, found.ID())
	assert.Equal(t, groupID, *found.GroupID())
	assert.Equal(t, "test-cms.txt", found.FileName())
}

func TestAssetRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	notFoundID := asset.NewAssetID()
	_, err := repo.FindByID(ctx, notFoundID)
	assert.Error(t, err)

	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("test.txt")
	a.SetUUID("test-uuid")

	err = repo.Save(ctx, a)
	require.NoError(t, err)

	// Test found
	found, err := repo.FindByID(ctx, assetID)
	assert.NoError(t, err)
	assert.Equal(t, assetID, found.ID())
	assert.Equal(t, groupID, *found.GroupID())
	assert.Equal(t, "test.txt", found.FileName())
	assert.Equal(t, "test-uuid", found.UUID())
}

func TestAssetRepository_FindByUUID(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Test not found
	_, err := repo.FindByUUID(ctx, "not-found-uuid")
	assert.Error(t, err)

	// Create and save test asset
	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()
	uuid := "test-find-uuid"

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("test.txt")
	a.SetUUID(uuid)

	err = repo.Save(ctx, a)
	require.NoError(t, err)

	// Test found
	found, err := repo.FindByUUID(ctx, uuid)
	assert.NoError(t, err)
	assert.Equal(t, assetID, found.ID())
	assert.Equal(t, uuid, found.UUID())
}

func TestAssetRepository_FindByURL(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create and save test asset
	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()
	testURL := "http://example.com/test-url.txt"

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("test-url.txt")
	a.SetURL(testURL)

	err := repo.Save(ctx, a)
	require.NoError(t, err)

	// Test found
	found, err := repo.FindByURL(ctx, testURL)
	assert.NoError(t, err)
	assert.Equal(t, assetID, found.ID())
	assert.Equal(t, testURL, found.URL())
}

func TestAssetRepository_FindByIDs(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Test empty IDs
	assets, err := repo.FindByIDs(ctx, asset.AssetIDList{})
	assert.NoError(t, err)
	assert.Nil(t, assets)

	// Create test assets
	groupID := asset.NewGroupID()
	now := time.Now()

	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &groupID, now, 1024, "text/plain")
	asset1.SetFileName("test1.txt")

	asset2ID := asset.NewAssetID()
	asset2 := asset.NewAsset(asset2ID, &groupID, now, 2048, "text/plain")
	asset2.SetFileName("test2.txt")

	// Save assets
	err = repo.Save(ctx, asset1)
	require.NoError(t, err)
	err = repo.Save(ctx, asset2)
	require.NoError(t, err)

	// Test finding by IDs
	ids := asset.AssetIDList{asset1ID, asset2ID}
	found, err := repo.FindByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Len(t, found, 2)

	// Verify order is maintained (same as input order)
	foundIDs := make([]asset.AssetID, len(found))
	for i, a := range found {
		if a != nil {
			foundIDs[i] = a.ID()
		}
	}

	// Should contain both assets
	assert.Contains(t, foundIDs, asset1ID)
	assert.Contains(t, foundIDs, asset2ID)
}

func TestAssetRepository_FindByIDList(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Test empty IDs
	assets, err := repo.FindByIDList(ctx, asset.AssetIDList{})
	assert.NoError(t, err)
	assert.Nil(t, assets)

	// Create and save test assets
	groupID := asset.NewGroupID()
	now := time.Now()

	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &groupID, now, 1024, "text/plain")
	asset1.SetFileName("test1.txt")

	err = repo.Save(ctx, asset1)
	require.NoError(t, err)

	// Test finding by ID list
	ids := asset.AssetIDList{asset1ID}
	var found asset.List
	found, err = repo.FindByIDList(ctx, ids)
	assert.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, asset1ID, found[0].ID())
}

func TestAssetRepository_Search(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create test assets
	groupID := asset.NewGroupID()
	now := time.Now()

	// Asset 1 - PDF
	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &groupID, now, 1024, "application/pdf")
	asset1.SetFileName("document.pdf")

	// Asset 2 - Image
	asset2ID := asset.NewAssetID()
	asset2 := asset.NewAsset(asset2ID, &groupID, now, 2048, "image/jpeg")
	asset2.SetFileName("image.jpg")

	// Save assets
	err := repo.Save(ctx, asset1)
	require.NoError(t, err)
	err = repo.Save(ctx, asset2)
	require.NoError(t, err)

	// Test search with keyword filter
	keyword := "document"
	filter := asset.AssetFilter{
		Keyword: &keyword,
	}

	results, pageInfo, err := repo.Search(ctx, groupID, filter)
	assert.NoError(t, err)
	assert.NotNil(t, pageInfo)
	assert.Len(t, results, 1)
	assert.Equal(t, asset1ID, results[0].ID())

	// Test search with content type filter
	filter = asset.AssetFilter{
		ContentTypes: []string{"image/jpeg"},
	}

	results, pageInfo, err = repo.Search(ctx, groupID, filter)
	assert.NoError(t, err)
	assert.NotNil(t, pageInfo)
	assert.Len(t, results, 1)
	assert.Equal(t, asset2ID, results[0].ID())

	// Test search without filters
	filter = asset.AssetFilter{}
	results, pageInfo, err = repo.Search(ctx, groupID, filter)
	assert.NoError(t, err)
	assert.NotNil(t, pageInfo)
	assert.Len(t, results, 2)
}

func TestAssetRepository_FindByGroup(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create test assets
	groupID := asset.NewGroupID()
	now := time.Now()

	// Create multiple assets with different sizes and names
	assets := []*asset.Asset{}
	for i := 0; i < 3; i++ {
		assetID := asset.NewAssetID()
		a := asset.NewAsset(assetID, &groupID, now.Add(time.Duration(i)*time.Hour), int64(1024*(i+1)), "text/plain")
		a.SetFileName(map[int]string{0: "alpha.txt", 1: "beta.txt", 2: "gamma.txt"}[i])
		assets = append(assets, a)

		err := repo.Save(ctx, a)
		require.NoError(t, err)
	}

	// Test with no filter
	filter := asset.AssetFilter{}
	sort := asset.AssetSort{By: asset.AssetSortTypeDate, Direction: asset.SortDirectionDesc}
	pagination := asset.Pagination{Limit: 10, Offset: 0}

	found, count, err := repo.FindByGroup(ctx, groupID, filter, sort, pagination)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.Len(t, found, 3)

	// Test with keyword filter
	keyword := "alpha"
	filter = asset.AssetFilter{Keyword: &keyword}
	found, count, err = repo.FindByGroup(ctx, groupID, filter, sort, pagination)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Len(t, found, 1)
	assert.Equal(t, "alpha.txt", found[0].FileName())

	// Test sorting by size ascending
	sort = asset.AssetSort{By: asset.AssetSortTypeSize, Direction: asset.SortDirectionAsc}
	filter = asset.AssetFilter{}
	found, _, err = repo.FindByGroup(ctx, groupID, filter, sort, pagination)
	assert.NoError(t, err)
	assert.True(t, found[0].Size() <= found[1].Size())

	// Test sorting by name descending
	sort = asset.AssetSort{By: asset.AssetSortTypeName, Direction: asset.SortDirectionDesc}
	found, _, err = repo.FindByGroup(ctx, groupID, filter, sort, pagination)
	assert.NoError(t, err)
	assert.Equal(t, "gamma.txt", found[0].FileName()) // gamma comes last alphabetically, but first when descending
}

func TestAssetRepository_FindByProject(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create test asset
	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("project-test.txt")

	err := repo.Save(ctx, a)
	require.NoError(t, err)

	// Test find by project
	filter := asset.AssetFilter{}
	results, pageInfo, err := repo.FindByProject(ctx, groupID, filter)
	assert.NoError(t, err)
	assert.NotNil(t, pageInfo)
	assert.Len(t, results, 1)
	assert.Equal(t, assetID, results[0].ID())

	// Test with keyword filter
	keyword := "project"
	filter = asset.AssetFilter{Keyword: &keyword}
	results, pageInfo, err = repo.FindByProject(ctx, groupID, filter)
	assert.NoError(t, err)
	assert.NotNil(t, pageInfo)
	assert.Len(t, results, 1)
	assert.Equal(t, assetID, results[0].ID())
}

func TestAssetRepository_FindByWorkspaceProject(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create test asset
	workspaceID := accountdomain.NewWorkspaceID()
	groupID := asset.GroupID(workspaceID)
	assetID := asset.NewAssetID()
	now := time.Now()

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("workspace-test.txt")
	a.SetURL("http://localhost/test.txt") // matches localhost pattern

	// Insert directly with core support flag
	doc, _ := mongodoc.NewAsset(a)
	_, err := repo.client.Client().InsertOne(ctx, bson.M{
		"id":          doc.ID,
		"groupid":     doc.Project,
		"createdat":   doc.CreatedAt,
		"size":        doc.Size,
		"filename":    doc.FileName,
		"url":         a.URL(),
		"coresupport": true,
	})
	require.NoError(t, err)

	// Test find by workspace project
	filter := asset.AssetFilter{}
	results, pageInfo, err := repo.FindByWorkspaceProject(ctx, workspaceID, &groupID, filter)
	assert.NoError(t, err)
	assert.NotNil(t, pageInfo)
	assert.Len(t, results, 1)
}

func TestAssetRepository_Delete(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create and save test asset
	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()

	a := asset.NewAsset(assetID, &groupID, now, 1024, "text/plain")
	a.SetFileName("delete-test.txt")

	err := repo.Save(ctx, a)
	require.NoError(t, err)

	// Verify asset exists
	_, err = repo.FindByID(ctx, assetID)
	assert.NoError(t, err)

	// Delete asset
	err = repo.Delete(ctx, assetID)
	assert.NoError(t, err)

	// Verify asset is deleted
	_, err = repo.FindByID(ctx, assetID)
	assert.Error(t, err)
}

func TestAssetRepository_DeleteMany(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create test assets
	groupID := asset.NewGroupID()
	now := time.Now()

	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &groupID, now, 1024, "text/plain")
	asset1.SetFileName("delete1.txt")

	asset2ID := asset.NewAssetID()
	asset2 := asset.NewAsset(asset2ID, &groupID, now, 2048, "text/plain")
	asset2.SetFileName("delete2.txt")

	// Save assets
	err := repo.Save(ctx, asset1)
	require.NoError(t, err)
	err = repo.Save(ctx, asset2)
	require.NoError(t, err)

	// Delete multiple assets
	ids := []asset.AssetID{asset1ID, asset2ID}
	err = repo.DeleteMany(ctx, ids)
	assert.NoError(t, err)

	// Verify assets are deleted
	_, err = repo.FindByID(ctx, asset1ID)
	assert.Error(t, err)
	_, err = repo.FindByID(ctx, asset2ID)
	assert.Error(t, err)
}

func TestAssetRepository_BatchDelete(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Test empty IDs
	err := repo.BatchDelete(ctx, asset.AssetIDList{})
	assert.NoError(t, err)

	// Create test assets
	groupID := asset.NewGroupID()
	now := time.Now()

	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &groupID, now, 1024, "text/plain")
	asset1.SetFileName("batch1.txt")

	asset2ID := asset.NewAssetID()
	asset2 := asset.NewAsset(asset2ID, &groupID, now, 2048, "text/plain")
	asset2.SetFileName("batch2.txt")

	// Save assets
	err = repo.Save(ctx, asset1)
	require.NoError(t, err)
	err = repo.Save(ctx, asset2)
	require.NoError(t, err)

	ids := asset.AssetIDList{asset1ID, asset2ID}
	err = repo.BatchDelete(ctx, ids)
	assert.NoError(t, err)

	_, err = repo.FindByID(ctx, asset1ID)
	assert.Error(t, err)
	_, err = repo.FindByID(ctx, asset2ID)
	assert.Error(t, err)
}

func TestAssetRepository_UpdateExtractionStatus(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	groupID := asset.NewGroupID()
	assetID := asset.NewAssetID()
	now := time.Now()

	a := asset.NewAsset(assetID, &groupID, now, 1024, "application/zip")
	a.SetFileName("archive.zip")

	err := repo.Save(ctx, a)
	require.NoError(t, err)

	status := asset.ExtractionStatusInProgress
	err = repo.UpdateExtractionStatus(ctx, assetID, status)
	assert.NoError(t, err)

	found, err := repo.FindByID(ctx, assetID)
	assert.NoError(t, err)
	assert.NotNil(t, found.ArchiveExtractionStatus())
	assert.Equal(t, status, *found.ArchiveExtractionStatus())
}

func TestAssetRepository_UpdateProject(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	fromGroupID := asset.NewGroupID()
	toGroupID := asset.NewGroupID()
	now := time.Now()

	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &fromGroupID, now, 1024, "text/plain")
	asset1.SetFileName("move1.txt")

	asset2ID := asset.NewAssetID()
	asset2 := asset.NewAsset(asset2ID, &fromGroupID, now, 2048, "text/plain")
	asset2.SetFileName("move2.txt")

	err := repo.Save(ctx, asset1)
	require.NoError(t, err)
	err = repo.Save(ctx, asset2)
	require.NoError(t, err)

	err = repo.UpdateProject(ctx, fromGroupID, toGroupID)
	assert.NoError(t, err)

	found1, err := repo.FindByID(ctx, asset1ID)
	assert.NoError(t, err)
	assert.Equal(t, toGroupID, *found1.GroupID())

	found2, err := repo.FindByID(ctx, asset2ID)
	assert.NoError(t, err)
	assert.Equal(t, toGroupID, *found2.GroupID())
}

func TestAssetRepository_TotalSizeByWorkspace(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	workspaceID := accountdomain.NewWorkspaceID()
	groupID := asset.GroupID(workspaceID)
	now := time.Now()

	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &groupID, now, 1024, "text/plain")
	asset1.SetFileName("size1.txt")

	asset2ID := asset.NewAssetID()
	asset2 := asset.NewAsset(asset2ID, &groupID, now, 2048, "text/plain")
	asset2.SetFileName("size2.txt")

	err := repo.Save(ctx, asset1)
	require.NoError(t, err)
	err = repo.Save(ctx, asset2)
	require.NoError(t, err)

	totalSize, err := repo.TotalSizeByWorkspace(ctx, workspaceID)
	assert.NoError(t, err)
	assert.Equal(t, int64(3072), totalSize)

	emptyWorkspaceID := accountdomain.NewWorkspaceID()
	totalSize, err = repo.TotalSizeByWorkspace(ctx, emptyWorkspaceID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), totalSize)
}

type mockFileRemover struct {
	removed []string
}

func (m *mockFileRemover) RemoveAsset(ctx context.Context, u *url.URL) error {
	m.removed = append(m.removed, u.String())
	return nil
}

func TestAssetRepository_RemoveByProjectWithFile(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create test assets
	groupID := asset.NewGroupID()
	now := time.Now()

	asset1ID := asset.NewAssetID()
	asset1 := asset.NewAsset(asset1ID, &groupID, now, 1024, "text/plain")
	asset1.SetFileName("remove1.txt")
	asset1.SetURL("http://example.com/remove1.txt")

	asset2ID := asset.NewAssetID()
	asset2 := asset.NewAsset(asset2ID, &groupID, now, 2048, "text/plain")
	asset2.SetFileName("remove2.txt")
	asset2.SetURL("http://example.com/remove2.txt")

	err := repo.Save(ctx, asset1)
	require.NoError(t, err)
	err = repo.Save(ctx, asset2)
	require.NoError(t, err)

	fileRemover := &mockFileRemover{}

	err = repo.RemoveByProjectWithFile(ctx, groupID, fileRemover)
	assert.NoError(t, err)

	assert.Len(t, fileRemover.removed, 2)
	assert.Contains(t, fileRemover.removed, "http://example.com/remove1.txt")
	assert.Contains(t, fileRemover.removed, "http://example.com/remove2.txt")

	_, err = repo.FindByID(ctx, asset1ID)
	assert.Error(t, err)
	_, err = repo.FindByID(ctx, asset2ID)
	assert.Error(t, err)
}

func TestAssetRepository_Filtered(t *testing.T) {
	ctx := context.Background()
	db := mongotest.Connect(t)(t)
	repo := NewAssetRepository(db)

	// Create test groups
	readableGroupID := asset.NewGroupID()
	writableGroupID := asset.NewGroupID()
	restrictedGroupID := asset.NewGroupID()

	// Create filter
	filter := asset.GroupFilter{
		Readable: asset.GroupIDList{readableGroupID, writableGroupID},
		Writable: asset.GroupIDList{writableGroupID},
	}

	filteredRepo := repo.Filtered(filter)

	// Create assets in different groups
	now := time.Now()

	// Asset in readable group
	readableAssetID := asset.NewAssetID()
	readableAsset := asset.NewAsset(readableAssetID, &readableGroupID, now, 1024, "text/plain")
	readableAsset.SetFileName("readable.txt")

	// Asset in writable group
	writableAssetID := asset.NewAssetID()
	writableAsset := asset.NewAsset(writableAssetID, &writableGroupID, now, 1024, "text/plain")
	writableAsset.SetFileName("writable.txt")

	// Asset in restricted group
	restrictedAssetID := asset.NewAssetID()
	restrictedAsset := asset.NewAsset(restrictedAssetID, &restrictedGroupID, now, 1024, "text/plain")
	restrictedAsset.SetFileName("restricted.txt")

	// Save all assets using original repo (no filter)
	err := repo.Save(ctx, readableAsset)
	require.NoError(t, err)
	err = repo.Save(ctx, writableAsset)
	require.NoError(t, err)
	err = repo.Save(ctx, restrictedAsset)
	require.NoError(t, err)

	// Test that filtered repo can read from readable groups
	found, err := filteredRepo.FindByID(ctx, readableAssetID)
	assert.NoError(t, err)
	assert.Equal(t, readableAssetID, found.ID())

	found, err = filteredRepo.FindByID(ctx, writableAssetID)
	assert.NoError(t, err)
	assert.Equal(t, writableAssetID, found.ID())

	// Test that filtered repo cannot read from restricted groups
	found, err = filteredRepo.FindByID(ctx, restrictedAssetID)
	assert.Error(t, err)
	assert.Nil(t, found)

	// Test writing to writable group succeeds
	newWritableAssetID := asset.NewAssetID()
	newWritableAsset := asset.NewAsset(newWritableAssetID, &writableGroupID, now, 512, "text/plain")
	newWritableAsset.SetFileName("new-writable.txt")

	err = filteredRepo.Save(ctx, newWritableAsset)
	assert.NoError(t, err)

	// Test writing to readable-only group fails
	newReadableAssetID := asset.NewAssetID()
	newReadableAsset := asset.NewAsset(newReadableAssetID, &readableGroupID, now, 512, "text/plain")
	newReadableAsset.SetFileName("new-readable.txt")

	err = filteredRepo.Save(ctx, newReadableAsset)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation denied")
}

func TestAssetRepository_GroupFilter(t *testing.T) {
	filter := &asset.GroupFilter{
		Readable: asset.GroupIDList{asset.NewGroupID()},
		Writable: asset.GroupIDList{asset.NewGroupID()},
	}

	assert.True(t, filter.CanRead(filter.Readable[0]))

	assert.True(t, filter.CanRead(filter.Writable[0]))

	assert.False(t, filter.CanRead(asset.NewGroupID()))

	assert.True(t, filter.CanWrite(filter.Writable[0]))
	assert.False(t, filter.CanWrite(filter.Readable[0]))
	assert.False(t, filter.CanWrite(asset.NewGroupID()))

	nilFilter := &asset.GroupFilter{}
	anyGroupID := asset.NewGroupID()
	assert.True(t, nilFilter.CanRead(anyGroupID))
	assert.True(t, nilFilter.CanWrite(anyGroupID))
}

func TestAssetRepository_docToAsset(t *testing.T) {
	doc := &assetDocument{
		ID:                      asset.NewAssetID().String(),
		GroupID:                 asset.NewGroupID().String(),
		CreatedAt:               time.Now(),
		Size:                    1024,
		ContentType:             "text/plain",
		ContentEncoding:         "gzip",
		PreviewType:             string(asset.PreviewTypeGeo),
		UUID:                    "test-uuid",
		URL:                     "http://example.com/test.txt",
		FileName:                "test.txt",
		ArchiveExtractionStatus: string(asset.ExtractionStatusDone),
		IntegrationID:           idx.New[asset.IntegrationIDType]().String(),
	}

	a, err := docToAsset(doc)
	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Equal(t, doc.FileName, a.FileName())
	assert.Equal(t, doc.UUID, a.UUID())
	assert.Equal(t, doc.URL, a.URL())
	assert.Equal(t, doc.ContentType, a.ContentType())
	assert.Equal(t, doc.ContentEncoding, a.ContentEncoding())
	assert.Equal(t, asset.PreviewTypeGeo, *a.PreviewType())
	assert.Equal(t, asset.ExtractionStatusDone, *a.ArchiveExtractionStatus())

	invalidDoc := &assetDocument{
		ID:      "invalid-id",
		GroupID: asset.NewGroupID().String(),
	}
	_, err = docToAsset(invalidDoc)
	assert.Error(t, err)

	invalidDoc = &assetDocument{
		ID:      asset.NewAssetID().String(),
		GroupID: "invalid-group-id",
	}
	_, err = docToAsset(invalidDoc)
	assert.Error(t, err)
}
