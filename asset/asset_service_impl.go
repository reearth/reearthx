package asset

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"path"
	"strings"
	"time"

	"fmt"

	"github.com/google/uuid"
)

var (
	ErrAssetNotFound     = errors.New("asset not found")
	ErrGroupNotFound     = errors.New("group not found")
	ErrInvalidParameters = errors.New("invalid parameters")
	ErrStorageFailure    = errors.New("storage operation failed")
)

const (
	DefaultPaginationLimit = 100
)

func logger(ctx context.Context, level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	switch level {
	case "error":
		log.Printf("[ERROR] %s", message)
	case "warn":
		log.Printf("[WARN] %s", message)
	case "info":
		log.Printf("[INFO] %s", message)
	case "debug":
		log.Printf("[DEBUG] %s", message)
	default:
		log.Printf("%s", message)
	}
}

var _ AssetService = &assetService{}

type assetService struct {
	assetRepo     AssetRepository
	groupRepo     GroupRepository
	storage       Storage
	fileProcessor FileProcessor
	zipExtractor  ZipExtractor
}

func NewAssetService(
	assetRepo AssetRepository,
	groupRepo GroupRepository,
	storage Storage,
	fileProcessor FileProcessor,
	zipExtractor ZipExtractor,
) AssetService {
	return &assetService{
		assetRepo:     assetRepo,
		groupRepo:     groupRepo,
		storage:       storage,
		fileProcessor: fileProcessor,
		zipExtractor:  zipExtractor,
	}
}

func (s *assetService) CreateAsset(ctx context.Context, param CreateAssetParam) (*Asset, error) {
	if param.GroupID.IsNil() {
		return nil, ErrInvalidParameters
	}

	group, err := s.groupRepo.FindByID(ctx, param.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	assetUUID := uuid.New().String()

	storageKey := path.Join(param.GroupID.String(), assetUUID, param.FileName)

	asset := &Asset{
		ID:              NewAssetID(),
		GroupID:         param.GroupID,
		CreatedAt:       time.Now(),
		Size:            param.Size,
		ContentType:     param.ContentType,
		ContentEncoding: param.ContentEncoding,
		UUID:            assetUUID,
		FileName:        param.FileName,
	}

	asset.PreviewType = s.fileProcessor.DetectPreviewType(param.FileName, param.ContentType)

	if param.File != nil {
		err = s.storage.Save(ctx, storageKey, param.File, param.Size, param.ContentType, param.ContentEncoding)
		if err != nil {
			return nil, ErrStorageFailure
		}

		url, err := s.storage.GenerateURL(ctx, storageKey, 24*time.Hour)
		if err != nil {
			return nil, ErrStorageFailure
		}
		asset.URL = url
	} else if param.URL != "" && param.Token != "" {
		asset.URL = param.URL
	} else {
		return nil, ErrInvalidParameters
	}

	if shouldExtractArchive(param.FileName, param.ContentType) && !param.SkipDecompression {
		status := ExtractionStatusPending
		asset.ArchiveExtractionStatus = &status
	} else {
		status := ExtractionStatusSkipped
		asset.ArchiveExtractionStatus = &status
	}

	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}

	if *asset.ArchiveExtractionStatus == ExtractionStatusPending {
		go s.handleArchiveExtraction(context.Background(), asset.ID, storageKey)
	}

	return asset, nil
}

func (s *assetService) GetAsset(ctx context.Context, id AssetID) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	return asset, nil
}

func (s *assetService) GetAssetFile(ctx context.Context, id AssetID) (*File, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)

	reader, err := s.storage.Get(ctx, storageKey)
	if err != nil {
		return nil, ErrStorageFailure
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	file := &File{}
	file.SetName(asset.FileName)
	if asset.Size > 0 {
		file.size = uint64(asset.Size)
	}
	file.contentType = asset.ContentType
	file.path = storageKey

	return file, nil
}

func (s *assetService) ListAssets(ctx context.Context, groupID GroupID, filter AssetFilter, sort AssetSort, pagination Pagination) ([]*Asset, int64, error) {
	return s.assetRepo.FindByGroup(ctx, groupID, filter, sort, pagination)
}

func (s *assetService) UpdateAsset(ctx context.Context, param UpdateAssetParam) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, param.ID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	if param.PreviewType != nil {
		asset.PreviewType = *param.PreviewType
	}

	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *assetService) DeleteAsset(ctx context.Context, id AssetID) error {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotFound
	}

	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)

	if err := s.storage.Delete(ctx, storageKey); err != nil {
		return ErrStorageFailure
	}

	return s.assetRepo.Delete(ctx, id)
}

func (s *assetService) DeleteAssets(ctx context.Context, ids []AssetID) error {
	assets, err := s.assetRepo.FindByIDs(ctx, ids)
	if err != nil {
		return err
	}

	for _, asset := range assets {
		storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)
		if err := s.storage.Delete(ctx, storageKey); err != nil {
			continue
		}
	}

	return s.assetRepo.DeleteMany(ctx, ids)
}

func (s *assetService) DecompressAsset(ctx context.Context, id AssetID) error {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotFound
	}

	if !shouldExtractArchive(asset.FileName, asset.ContentType) {
		return errors.New("asset is not an extractable archive")
	}

	status := ExtractionStatusPending
	asset.ArchiveExtractionStatus = &status
	if err := s.assetRepo.UpdateExtractionStatus(ctx, id, status); err != nil {
		return err
	}

	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)
	go s.handleArchiveExtraction(context.Background(), id, storageKey)

	return nil
}

type UploadCursor struct {
	UUID   string
	Cursor string
}

func (c UploadCursor) String() string {
	return c.UUID + "_" + c.Cursor
}

func ParseUploadCursor(c string) (*UploadCursor, error) {
	uuid, cursor, found := strings.Cut(c, "_")
	if !found {
		return nil, fmt.Errorf("invalid cursor format: separator not found")
	}
	return &UploadCursor{
		UUID:   uuid,
		Cursor: cursor,
	}, nil
}

func WrapUploadCursor(uuid, cursor string) string {
	if cursor == "" {
		return ""
	}
	return UploadCursor{UUID: uuid, Cursor: cursor}.String()
}

func (s *assetService) CreateAssetUpload(ctx context.Context, param CreateAssetUploadParam) (*AssetUploadInfo, error) {
	if param.GroupID.IsNil() || param.FileName == "" || param.ContentLength <= 0 {
		return nil, ErrInvalidParameters
	}

	group, err := s.groupRepo.FindByID(ctx, param.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	var uploadParam struct {
		UUID          string
		FileName      string
		ContentLength int64
		ContentType   string
		ExpiresAt     time.Time
		Cursor        string
	}

	// Check if this is a continuation of a previous upload
	if param.Cursor != "" {
		cursor, err := ParseUploadCursor(param.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}

		// Here you would typically look up the previous upload info from a repository
		// For this implementation, we'll just use the cursor information
		uploadParam = struct {
			UUID          string
			FileName      string
			ContentLength int64
			ContentType   string
			ExpiresAt     time.Time
			Cursor        string
		}{
			UUID:          cursor.UUID,
			FileName:      param.FileName,
			ContentLength: param.ContentLength,
			ContentType:   s.fileProcessor.DetectContentType(param.FileName, nil),
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			Cursor:        cursor.Cursor,
		}
	} else {
		// New upload
		uploadParam = struct {
			UUID          string
			FileName      string
			ContentLength int64
			ContentType   string
			ExpiresAt     time.Time
			Cursor        string
		}{
			UUID:          uuid.New().String(),
			FileName:      param.FileName,
			ContentLength: param.ContentLength,
			ContentType:   s.fileProcessor.DetectContentType(param.FileName, nil),
			ExpiresAt:     time.Now().Add(24 * time.Hour),
			Cursor:        "",
		}
	}

	// Generate storage key
	storageKey := path.Join(param.GroupID.String(), uploadParam.UUID, uploadParam.FileName)

	// Generate upload URL
	uploadURL, err := s.storage.GenerateUploadURL(
		ctx,
		storageKey,
		uploadParam.ContentLength,
		uploadParam.ContentType,
		param.ContentEncoding,
		1*time.Hour,
	)
	if err != nil {
		return nil, ErrStorageFailure
	}

	// Create upload info
	return &AssetUploadInfo{
		Token:           uploadParam.UUID,
		URL:             uploadURL,
		ContentType:     uploadParam.ContentType,
		ContentLength:   uploadParam.ContentLength,
		ContentEncoding: param.ContentEncoding,
		Next:            WrapUploadCursor(uploadParam.UUID, "next-chunk"), // In a real implementation, this would be a real cursor
	}, nil
}

func (s *assetService) handleArchiveExtraction(ctx context.Context, assetID AssetID, storageKey string) {
	// Update asset status to in progress
	err := s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusInProgress)
	if err != nil {
		logger(ctx, "error", "Failed to update extraction status: %v", err)
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}

	// Get asset file
	reader, err := s.storage.Get(ctx, storageKey)
	if err != nil {
		logger(ctx, "error", "Failed to get asset file: %v", err)
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	// Read all data from reader
	data, err := io.ReadAll(reader)
	if err != nil {
		logger(ctx, "error", "Failed to read asset data: %v", err)
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}

	// Create a reader from the data
	readerAt := bytes.NewReader(data)

	// Extract the archive
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err = s.zipExtractor.Extract(ctx, assetID, readerAt, int64(len(data)))
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			logger(ctx, "warn", "Archive extraction attempt %d failed: %v, retrying...", i+1, err)
			time.Sleep(2 * time.Second)
			// Reset reader position for next attempt
			_, _ = readerAt.Seek(0, io.SeekStart)
		} else {
			logger(ctx, "error", "Failed to extract archive after %d attempts: %v", maxRetries, err)
			_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
			return
		}
	}

	// Update extraction status to done
	err = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusDone)
	if err != nil {
		logger(ctx, "error", "Failed to update extraction status to done: %v", err)
		return
	}

	// Get the asset to check if we need to update its preview type
	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		logger(ctx, "error", "Failed to find asset after extraction: %v", err)
		return
	}

	logger(ctx, "info", "Successfully extracted archive for asset %s", assetID)

	// Update preview type based on extracted files if needed
	detectAndUpdatePreviewType(ctx, s, asset)
}

func shouldExtractArchive(fileName string, contentType string) bool {
	lowerFileName := strings.ToLower(fileName)

	if strings.HasSuffix(lowerFileName, ".zip") ||
		strings.HasSuffix(lowerFileName, ".tar.gz") ||
		strings.HasSuffix(lowerFileName, ".tar") ||
		strings.HasSuffix(lowerFileName, ".7z") {
		return true
	}

	switch contentType {
	case "application/zip", "application/x-zip-compressed",
		"application/gzip", "application/x-gzip",
		"application/x-tar", "application/tar",
		"application/x-7z-compressed":
		return true
	default:
		return false
	}
}

// FindByID implements the Asset interface method
func (s *assetService) FindByID(ctx context.Context, id AssetID, operator *Operator) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	return asset, nil
}

// FindByUUID implements the Asset interface method
func (s *assetService) FindByUUID(ctx context.Context, uuid string, operator *Operator) (*Asset, error) {
	if uuid == "" {
		return nil, ErrInvalidParameters
	}

	// Add a FindByUUID method to AssetRepository interface and call it here
	// For now, we'll implement a search across groups with a filter
	emptyGroupID := GroupID{} // This will be ignored when using the UUID filter

	// Create a custom filter for UUID
	filter := AssetFilter{
		Keyword: uuid, // Assuming the keyword can be used to search UUIDs
	}

	pagination := Pagination{
		Offset: 0,
		Limit:  1, // We only need one result
	}

	sort := AssetSort{
		By:        AssetSortTypeDate,
		Direction: SortDirectionDesc,
	}

	assets, _, err := s.assetRepo.FindByGroup(ctx, emptyGroupID, filter, sort, pagination)
	if err != nil {
		return nil, err
	}

	for _, asset := range assets {
		if asset.UUID == uuid {
			return asset, nil
		}
	}

	return nil, ErrAssetNotFound
}

// FindByIDs implements the Asset interface method
func (s *assetService) FindByIDs(ctx context.Context, ids []AssetID, operator *Operator) ([]*Asset, error) {
	assets, err := s.assetRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// FindByProject implements the Asset interface method
func (s *assetService) FindByProject(ctx context.Context, projectID ProjectID, filter AssetFilter, operator *Operator) ([]*Asset, *PageInfo, error) {
	// Convert ProjectID to GroupID
	groupID, err := GroupIDFrom(projectID.String())
	if err != nil {
		return nil, nil, err
	}

	// Create pagination
	pagination := Pagination{
		Offset: 0,
		Limit:  DefaultPaginationLimit,
	}

	// Use default sort
	sort := AssetSort{
		By:        AssetSortTypeDate,
		Direction: SortDirectionDesc,
	}

	assets, totalCount, err := s.assetRepo.FindByGroup(ctx, groupID, filter, sort, pagination)
	if err != nil {
		return nil, nil, err
	}

	pageInfo := &PageInfo{
		TotalCount: totalCount,
		HasNext:    totalCount > int64(len(assets)),
	}

	return assets, pageInfo, nil
}

// FindFileByID implements the Asset interface method
func (s *assetService) FindFileByID(ctx context.Context, id AssetID, operator *Operator) (*File, error) {
	return s.GetAssetFile(ctx, id)
}

// FindFilesByIDs implements the Asset interface method
func (s *assetService) FindFilesByIDs(ctx context.Context, ids AssetIDList, operator *Operator) (map[AssetID]*File, error) {
	result := make(map[AssetID]*File)

	for _, id := range ids {
		file, err := s.GetAssetFile(ctx, id)
		if err != nil {
			continue // Skip files that can't be found
		}
		result[id] = file
	}

	return result, nil
}

// DownloadByID implements the Asset interface method
func (s *assetService) DownloadByID(ctx context.Context, id AssetID, headers map[string]string, operator *Operator) (io.ReadCloser, map[string]string, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if asset == nil {
		return nil, nil, ErrAssetNotFound
	}

	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)

	reader, err := s.storage.Get(ctx, storageKey)
	if err != nil {
		return nil, nil, ErrStorageFailure
	}

	responseHeaders := map[string]string{
		"Content-Type":        asset.ContentType,
		"Content-Disposition": "attachment; filename=" + asset.FileName,
		"Content-Length":      fmt.Sprint(asset.Size),
	}

	if asset.ContentEncoding != "" {
		responseHeaders["Content-Encoding"] = asset.ContentEncoding
	}

	return reader, responseHeaders, nil
}

// Create implements the Asset interface method
func (s *assetService) Create(ctx context.Context, param CreateAssetParam, operator *Operator) (*Asset, *File, error) {
	// Set operator info if provided
	var opInfo OperatorInfo
	if operator != nil {
		opInfo = OperatorInfo{
			Type: operator.Type,
			ID:   operator.ID,
		}
	}

	asset, err := s.CreateAsset(ctx, param)
	if err != nil {
		return nil, nil, err
	}

	// Set the operator info
	asset.CreatedBy = opInfo

	// Save the updated asset
	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, nil, err
	}

	// Get the file
	file, err := s.GetAssetFile(ctx, asset.ID)
	if err != nil {
		return asset, nil, err
	}

	return asset, file, nil
}

// Update implements the Asset interface method
func (s *assetService) Update(ctx context.Context, param UpdateAssetParam, operator *Operator) (*Asset, error) {
	return s.UpdateAsset(ctx, param)
}

// UpdateFiles implements the Asset interface method
func (s *assetService) UpdateFiles(ctx context.Context, id AssetID, status *ExtractionStatus, operator *Operator) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	if status != nil {
		asset.ArchiveExtractionStatus = status
		if err := s.assetRepo.UpdateExtractionStatus(ctx, id, *status); err != nil {
			return nil, err
		}
	}

	return asset, nil
}

// Delete implements the Asset interface method
func (s *assetService) Delete(ctx context.Context, id AssetID, operator *Operator) (AssetID, error) {
	err := s.DeleteAsset(ctx, id)
	if err != nil {
		// Return a zero value AssetID
		return AssetID{}, err
	}
	return id, nil
}

// BatchDelete implements the Asset interface method
func (s *assetService) BatchDelete(ctx context.Context, ids AssetIDList, operator *Operator) ([]AssetID, error) {
	idArray := []AssetID(ids)
	err := s.DeleteAssets(ctx, idArray)
	if err != nil {
		return nil, err
	}
	return idArray, nil
}

// Decompress implements the Asset interface method
func (s *assetService) Decompress(ctx context.Context, id AssetID, operator *Operator) (*Asset, error) {
	err := s.DecompressAsset(ctx, id)
	if err != nil {
		return nil, err
	}

	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	return asset, nil
}

// Publish implements the Asset interface method
func (s *assetService) Publish(ctx context.Context, id AssetID, operator *Operator) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	// Implementation of publishing logic
	// 1. Generate a public URL for the asset that doesn't expire
	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)

	// Generate a long-term public URL (e.g., 10 years)
	publicURL, err := s.storage.GenerateURL(ctx, storageKey, 10*365*24*time.Hour)
	if err != nil {
		return nil, ErrStorageFailure
	}

	// Update the asset's URL to the public one
	asset.URL = publicURL

	// Save the updated asset
	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// Unpublish implements the Asset interface method
func (s *assetService) Unpublish(ctx context.Context, id AssetID, operator *Operator) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	// Implementation of unpublishing logic
	// 1. Generate a short-term URL for the asset
	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)

	// Generate a temporary URL (e.g., 24 hours)
	temporaryURL, err := s.storage.GenerateURL(ctx, storageKey, 24*time.Hour)
	if err != nil {
		return nil, ErrStorageFailure
	}

	// Update the asset's URL to the temporary one
	asset.URL = temporaryURL

	// Save the updated asset
	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// CreateUpload implements the Asset interface method
func (s *assetService) CreateUpload(ctx context.Context, param CreateAssetUploadParam, operator *Operator) (*AssetUpload, error) {
	info, err := s.CreateAssetUpload(ctx, param)
	if err != nil {
		return nil, err
	}

	// Convert AssetUploadInfo to AssetUpload
	upload := &AssetUpload{
		Token:           info.Token,
		URL:             info.URL,
		ContentType:     info.ContentType,
		ContentLength:   info.ContentLength,
		ContentEncoding: info.ContentEncoding,
		Next:            info.Next,
	}

	return upload, nil
}

// RetryDecompression retries a failed archive extraction
func (s *assetService) RetryDecompression(ctx context.Context, id string) error {
	// Convert string id to AssetID
	assetID, err := AssetIDFrom(id)
	if err != nil {
		return err
	}

	// Check if asset exists and get its status
	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotFound
	}

	// Only retry if current status is failed
	if asset.ArchiveExtractionStatus == nil || *asset.ArchiveExtractionStatus != ExtractionStatusFailed {
		return fmt.Errorf("cannot retry decompression, current status: %v", asset.ArchiveExtractionStatus)
	}

	// Update status to pending
	status := ExtractionStatusPending
	if err := s.assetRepo.UpdateExtractionStatus(ctx, assetID, status); err != nil {
		return err
	}

	// Start extraction in background
	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)
	go s.handleArchiveExtraction(context.Background(), assetID, storageKey)

	return nil
}

// detectAndUpdatePreviewType analyzes extracted files and updates the asset's preview type if needed
func detectAndUpdatePreviewType(ctx context.Context, s *assetService, asset *Asset) {
	// Get list of extracted files
	files, err := s.storage.ListFiles(ctx, path.Join(asset.GroupID.String(), asset.UUID))
	if err != nil {
		logger(ctx, "error", "Failed to list extracted files: %v", err)
		return
	}

	// Detect preview type based on extracted files
	var previewType PreviewType
	for _, file := range files {
		filename := path.Base(file)
		ext := strings.ToLower(path.Ext(filename))

		// Check for 3D tiles
		if filename == "tileset.json" {
			previewType = PreviewTypeGeo3DTiles
			break
		}

		// Check for MVT (Mapbox Vector Tiles)
		if ext == ".mvt" {
			previewType = PreviewTypeGeoMVT
			break
		}

		// Check for GeoJSON
		if ext == ".geojson" {
			previewType = PreviewTypeGeo
			break
		}
	}

	// Only update if a relevant preview type was detected
	if previewType != "" && previewType != asset.PreviewType {
		logger(ctx, "info", "Updating asset %s preview type from %s to %s", asset.ID, asset.PreviewType, previewType)
		asset.PreviewType = previewType
		if err := s.assetRepo.Save(ctx, asset); err != nil {
			logger(ctx, "error", "Failed to update asset preview type: %v", err)
		}
	}
}
