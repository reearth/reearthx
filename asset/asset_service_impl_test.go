package asset

// import (
// 	"bytes"
// 	"context"
// 	"io"
// 	"path"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func toPtr[T any](v T) *T {
// 	return &v
// }

// type MockAssetRepository struct {
// 	mock.Mock
// }

// func (m *MockAssetRepository) Save(ctx context.Context, asset *Asset) error {
// 	args := m.Called(ctx, asset)
// 	return args.Error(0)
// }

// func (m *MockAssetRepository) FindByID(ctx context.Context, id AssetID) (*Asset, error) {
// 	args := m.Called(ctx, id)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*Asset), args.Error(1)
// }

// func (m *MockAssetRepository) FindByUUID(ctx context.Context, uuid string) (*Asset, error) {
// 	args := m.Called(ctx, uuid)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*Asset), args.Error(1)
// }

// func (m *MockAssetRepository) FindByIDs(ctx context.Context, ids []AssetID) ([]*Asset, error) {
// 	args := m.Called(ctx, ids)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).([]*Asset), args.Error(1)
// }

// func (m *MockAssetRepository) FindByGroup(ctx context.Context, groupID GroupID, filter AssetFilter, sort AssetSort, pagination Pagination) ([]*Asset, int64, error) {
// 	args := m.Called(ctx, groupID, filter, sort, pagination)
// 	return args.Get(0).([]*Asset), args.Get(1).(int64), args.Error(2)
// }

// func (m *MockAssetRepository) Delete(ctx context.Context, id AssetID) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }

// func (m *MockAssetRepository) DeleteMany(ctx context.Context, ids []AssetID) error {
// 	args := m.Called(ctx, ids)
// 	return args.Error(0)
// }

// func (m *MockAssetRepository) UpdateExtractionStatus(ctx context.Context, id AssetID, status ExtractionStatus) error {
// 	args := m.Called(ctx, id, status)
// 	return args.Error(0)
// }

// type MockGroupRepository struct {
// 	mock.Mock
// }

// func (m *MockGroupRepository) FindByID(ctx context.Context, id GroupID) (*Group, error) {
// 	args := m.Called(ctx, id)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*Group), args.Error(1)
// }

// func (m *MockGroupRepository) Save(ctx context.Context, group *Group) error {
// 	args := m.Called(ctx, group)
// 	return args.Error(0)
// }

// func (m *MockGroupRepository) Delete(ctx context.Context, id GroupID) error {
// 	args := m.Called(ctx, id)
// 	return args.Error(0)
// }

// func (m *MockGroupRepository) UpdatePolicy(ctx context.Context, id GroupID, policyID *PolicyID) error {
// 	args := m.Called(ctx, id, policyID)
// 	return args.Error(0)
// }

// type MockStorage struct {
// 	mock.Mock
// }

// func (m *MockStorage) Save(ctx context.Context, key string, data io.Reader, size int64, contentType string, contentEncoding string) error {
// 	args := m.Called(ctx, key, data, size, contentType, contentEncoding)
// 	return args.Error(0)
// }

// func (m *MockStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
// 	args := m.Called(ctx, key)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(io.ReadCloser), args.Error(1)
// }

// func (m *MockStorage) Delete(ctx context.Context, key string) error {
// 	args := m.Called(ctx, key)
// 	return args.Error(0)
// }

// func (m *MockStorage) GenerateURL(ctx context.Context, key string, expires time.Duration) (string, error) {
// 	args := m.Called(ctx, key, expires)
// 	return args.String(0), args.Error(1)
// }

// func (m *MockStorage) GenerateUploadURL(ctx context.Context, key string, size int64, contentType string, contentEncoding string, expires time.Duration) (string, error) {
// 	args := m.Called(ctx, key, size, contentType, contentEncoding, expires)
// 	return args.String(0), args.Error(1)
// }

// func (m *MockStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
// 	args := m.Called(ctx, prefix)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).([]string), args.Error(1)
// }

// type MockFileProcessor struct {
// 	mock.Mock
// }

// func (m *MockFileProcessor) DetectContentType(filename string, data []byte) string {
// 	args := m.Called(filename, data)
// 	return args.String(0)
// }

// func (m *MockFileProcessor) DetectPreviewType(filename string, contentType string) PreviewType {
// 	args := m.Called(filename, contentType)
// 	return args.Get(0).(PreviewType)
// }

// type MockZipExtractor struct {
// 	mock.Mock
// }

// func (m *MockZipExtractor) Extract(ctx context.Context, assetID AssetID, reader io.ReaderAt, size int64) error {
// 	args := m.Called(ctx, assetID, reader, size)
// 	return args.Error(0)
// }

// type MockReadCloser struct {
// 	*bytes.Reader
// 	mock.Mock
// }

// func NewMockReadCloser(s string) *MockReadCloser {
// 	return &MockReadCloser{
// 		Reader: bytes.NewReader([]byte(s)),
// 	}
// }

// func (m *MockReadCloser) Close() error {
// 	args := m.Called()
// 	return args.Error(0)
// }

// func TestCreateAsset(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	groupID := NewGroupID()
// 	group := &Group{ID: groupID}

// 	fileContent := "test file content"
// 	fileName := "test.txt"
// 	fileSize := int64(len(fileContent))
// 	contentType := "text/plain"

// 	groupRepo.On("FindByID", ctx, groupID).Return(group, nil)
// 	fileProcessor.On("DetectPreviewType", fileName, contentType).Return(PreviewTypeUnknown)
// 	storage.On("Save", ctx, mock.Anything, mock.Anything, fileSize, contentType, "").Return(nil)
// 	storage.On("GenerateURL", ctx, mock.Anything, mock.Anything).Return("https://example.com/test.txt", nil)
// 	assetRepo.On("Save", ctx, mock.AnythingOfType("*asset.Asset")).Return(nil)

// 	param := CreateAssetParam{
// 		GroupID:     groupID,
// 		File:        strings.NewReader(fileContent),
// 		FileName:    fileName,
// 		Size:        fileSize,
// 		ContentType: contentType,
// 	}

// 	asset, err := service.CreateAsset(ctx, param)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, asset)
// 	assert.Equal(t, groupID, asset.GroupID())
// 	assert.Equal(t, fileName, asset.FileName())
// 	assert.Equal(t, fileSize, asset.Size())
// 	assert.Equal(t, contentType, asset.ContentType())
// 	assert.Equal(t, PreviewTypeUnknown, asset.PreviewType())
// 	assert.Equal(t, "https://example.com/test.txt", asset.URL())

// 	groupRepo.AssertExpectations(t)
// 	fileProcessor.AssertExpectations(t)
// 	storage.AssertExpectations(t)
// 	assetRepo.AssertExpectations(t)
// }

// func TestGetAsset(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	assetID := NewAssetID()
// 	groupID := NewGroupID()

// 	expectedAsset := NewAsset(assetID, groupID, time.Now(), 1024, "text/plain")
// 	expectedAsset.SetUUID(uuid.New().String())
// 	expectedAsset.SetFileName("test.txt")
// 	expectedAsset.SetPreviewType(PreviewTypeUnknown)

// 	assetRepo.On("FindByID", ctx, assetID).Return(expectedAsset, nil)

// 	asset, err := service.GetAsset(ctx, assetID)

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedAsset, asset)

// 	assetRepo.AssertExpectations(t)
// }

// func TestGetAssetNotFound(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	assetID := NewAssetID()

// 	assetRepo.On("FindByID", ctx, assetID).Return(nil, nil)

// 	asset, err := service.GetAsset(ctx, assetID)

// 	assert.Error(t, err)
// 	assert.Equal(t, ErrAssetNotFound, err)
// 	assert.Nil(t, asset)

// 	assetRepo.AssertExpectations(t)
// }

// func TestListAssets(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	groupID := NewGroupID()
// 	filter := AssetFilter{Keyword: "test"}
// 	sort := AssetSort{By: AssetSortTypeDate, Direction: SortDirectionDesc}
// 	pagination := Pagination{Offset: 0, Limit: 10}

// 	asset1 := NewAsset(NewAssetID(), groupID, time.Now(), 1024, "text/plain")
// 	asset1.SetUUID(uuid.New().String())
// 	asset1.SetFileName("test1.txt")
// 	asset1.SetPreviewType(PreviewTypeUnknown)

// 	asset2 := NewAsset(NewAssetID(), groupID, time.Now().Add(-time.Hour), 2048, "text/plain")
// 	asset2.SetUUID(uuid.New().String())
// 	asset2.SetFileName("test2.txt")
// 	asset2.SetPreviewType(PreviewTypeUnknown)

// 	expectedAssets := []*Asset{asset1, asset2}

// 	expectedTotal := int64(2)

// 	assetRepo.On("FindByGroup", ctx, groupID, filter, sort, pagination).Return(expectedAssets, expectedTotal, nil)

// 	assets, total, err := service.ListAssets(ctx, groupID, filter, sort, pagination)

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedAssets, assets)
// 	assert.Equal(t, expectedTotal, total)

// 	assetRepo.AssertExpectations(t)
// }

// func TestDeleteAsset(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	assetID := NewAssetID()
// 	groupID := NewGroupID()
// 	uuid := uuid.New().String()
// 	fileName := "test.txt"

// 	asset := NewAsset(assetID, groupID, time.Now(), 1024, "text/plain")
// 	asset.SetUUID(uuid)
// 	asset.SetFileName(fileName)

// 	storageKey := groupID.String() + "/" + uuid + "/" + fileName

// 	assetRepo.On("FindByID", ctx, assetID).Return(asset, nil)
// 	storage.On("Delete", ctx, storageKey).Return(nil)
// 	assetRepo.On("Delete", ctx, assetID).Return(nil)

// 	err := service.DeleteAsset(ctx, assetID)

// 	assert.NoError(t, err)

// 	assetRepo.AssertExpectations(t)
// 	storage.AssertExpectations(t)
// }

// func TestArchiveExtraction(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	assetID := NewAssetID()
// 	groupID := NewGroupID()
// 	uuid := "test-uuid"
// 	fileName := "test.zip"

// 	// Set up the asset
// 	asset := NewAsset(assetID, groupID, time.Now(), 1024, "application/zip")
// 	asset.SetUUID(uuid)
// 	asset.SetFileName(fileName)
// 	asset.SetArchiveExtractionStatus(toPtr(ExtractionStatusPending))

// 	// Setup file content
// 	fileContent := "mock zip file content"
// 	storageKey := path.Join(groupID.String(), uuid, fileName)

// 	// Mock repository calls
// 	assetRepo.On("UpdateExtractionStatus", ctx, assetID, ExtractionStatusInProgress).Return(nil)
// 	assetRepo.On("UpdateExtractionStatus", ctx, assetID, ExtractionStatusDone).Return(nil)

// 	// Mock storage calls
// 	mockReader := NewMockReadCloser(fileContent)
// 	mockReader.On("Close").Return(nil)
// 	storage.On("Get", ctx, storageKey).Return(mockReader, nil)

// 	// After extraction, it should list files to detect preview type
// 	extractedFiles := []string{
// 		path.Join(groupID.String(), uuid, "tileset.json"),
// 		path.Join(groupID.String(), uuid, "data.b3dm"),
// 	}
// 	storage.On("ListFiles", ctx, path.Join(groupID.String(), uuid)).Return(extractedFiles, nil)

// 	// After listing files, it should get the asset to update preview type
// 	assetRepo.On("FindByID", ctx, assetID).Return(asset, nil)

// 	// Should update the asset with the new preview type
// 	assetRepo.On("Save", ctx, mock.MatchedBy(func(a *Asset) bool {
// 		return a.ID() == assetID && a.PreviewType() == PreviewTypeGeo3DTiles
// 	})).Return(nil)

// 	// Mock zip extractor
// 	zipExtractor.On("Extract", ctx, assetID, mock.AnythingOfType("*bytes.Reader"), int64(len(fileContent))).Return(nil)

// 	// Call the method directly
// 	service.(*assetService).handleArchiveExtraction(ctx, assetID, storageKey)

// 	// Assert all expectations were met
// 	assetRepo.AssertExpectations(t)
// 	storage.AssertExpectations(t)
// 	zipExtractor.AssertExpectations(t)
// 	mockReader.AssertExpectations(t)

// 	// Verify the asset was updated with the correct preview type
// 	assert.Equal(t, PreviewTypeGeo3DTiles, asset.PreviewType())
// }

// func TestRetryDecompression(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	assetID := NewAssetID()
// 	groupID := NewGroupID()
// 	uuid := "test-uuid"
// 	fileName := "test.zip"

// 	asset := NewAsset(assetID, groupID, time.Now(), 1024, "application/zip")
// 	asset.SetUUID(uuid)
// 	asset.SetFileName(fileName)
// 	asset.SetArchiveExtractionStatus(toPtr(ExtractionStatusFailed))

// 	assetRepo.On("FindByID", ctx, assetID).Return(asset, nil)
// 	assetRepo.On("UpdateExtractionStatus", ctx, assetID, ExtractionStatusPending).Return(nil)

// 	err := service.RetryDecompression(ctx, assetID.String())

// 	assert.NoError(t, err)
// 	assetRepo.AssertExpectations(t)

// 	assetWithWrongStatus := NewAsset(assetID, groupID, time.Now(), 1024, "application/zip")
// 	assetWithWrongStatus.SetUUID(uuid)
// 	assetWithWrongStatus.SetFileName(fileName)
// 	assetWithWrongStatus.SetArchiveExtractionStatus(toPtr(ExtractionStatusDone))

// 	assetRepo = new(MockAssetRepository)
// 	service = NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)
// 	assetRepo.On("FindByID", ctx, assetID).Return(assetWithWrongStatus, nil)

// 	err = service.RetryDecompression(ctx, assetID.String())
// 	assert.Error(t, err)
// 	assetRepo.AssertExpectations(t)
// }

// func TestCreateAssetUpload(t *testing.T) {
// 	ctx := context.Background()
// 	assetRepo := new(MockAssetRepository)
// 	groupRepo := new(MockGroupRepository)
// 	storage := new(MockStorage)
// 	fileProcessor := new(MockFileProcessor)
// 	zipExtractor := new(MockZipExtractor)

// 	service := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)

// 	groupID := NewGroupID()
// 	group := &Group{ID: groupID}
// 	fileName := "test.txt"
// 	contentType := "text/plain"
// 	contentLength := int64(1024)

// 	// Mock the group repository
// 	groupRepo.On("FindByID", ctx, groupID).Return(group, nil)

// 	// Mock content type detection - use mock.Anything for the data parameter to avoid type issues
// 	fileProcessor.On("DetectContentType", fileName, mock.Anything).Return(contentType)

// 	// For the first test, we'll use a specific token and storage key
// 	initialURL := "https://example.com/upload"

// 	// We need to use a matcher that can distinguish between first and second calls
// 	// First call - create initial upload
// 	storage.On(
// 		"GenerateUploadURL",
// 		ctx,
// 		mock.AnythingOfType("string"), // Accept any string for the key
// 		contentLength,
// 		contentType,
// 		"",
// 		mock.Anything,
// 	).Return(initialURL, nil).Once() // Only match once

// 	// Test new upload (without cursor)
// 	param := CreateAssetUploadParam{
// 		GroupID:       groupID,
// 		FileName:      fileName,
// 		ContentLength: contentLength,
// 	}

// 	result, err := service.CreateAssetUpload(ctx, param)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, result)
// 	assert.Equal(t, initialURL, result.URL)
// 	assert.Equal(t, contentType, result.ContentType)
// 	assert.Equal(t, contentLength, result.ContentLength)
// 	assert.NotEmpty(t, result.Token)
// 	assert.NotEmpty(t, result.Next)

// 	// For the second test with the resumable upload
// 	resumableURL := "https://example.com/upload-resumable"

// 	// Set up second mock for resumable upload - will only match after the first one is used
// 	storage.On(
// 		"GenerateUploadURL",
// 		ctx,
// 		mock.AnythingOfType("string"), // Accept any string for the key
// 		contentLength,
// 		contentType,
// 		"",
// 		mock.Anything,
// 	).Return(resumableURL, nil).Once() // Only match once

// 	// Create cursor with the actual token we got from the first call
// 	cursor := WrapUploadCursor(result.Token, "chunk1")

// 	// Create param with cursor
// 	resumableParam := CreateAssetUploadParam{
// 		GroupID:       groupID,
// 		FileName:      fileName,
// 		ContentLength: contentLength,
// 		Cursor:        cursor,
// 	}

// 	resumableResult, err := service.CreateAssetUpload(ctx, resumableParam)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, resumableResult)
// 	assert.Equal(t, resumableURL, resumableResult.URL)
// 	assert.Equal(t, contentType, resumableResult.ContentType)
// 	assert.Equal(t, contentLength, resumableResult.ContentLength)
// 	assert.Equal(t, result.Token, resumableResult.Token) // Should keep same token from first result
// 	assert.NotEmpty(t, resumableResult.Next)

// 	// Verify all mocks were called correctly
// 	groupRepo.AssertExpectations(t)
// 	fileProcessor.AssertExpectations(t)
// 	storage.AssertExpectations(t)

// 	// Test cursor parsing functions
// 	parsedCursor, err := ParseUploadCursor(cursor)
// 	assert.NoError(t, err)
// 	assert.Equal(t, result.Token, parsedCursor.UUID)
// 	assert.Equal(t, "chunk1", parsedCursor.Cursor)

// 	// Test invalid cursor
// 	_, err = ParseUploadCursor("invalid-cursor-format")
// 	assert.Error(t, err)
// }
