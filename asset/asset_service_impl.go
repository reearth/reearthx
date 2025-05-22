package asset

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"time"

	"log/slog"

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

	if param.Cursor != "" {
		cursor, err := ParseUploadCursor(param.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}

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

	return &AssetUploadInfo{
		Token:           uploadParam.UUID,
		URL:             uploadURL,
		ContentType:     uploadParam.ContentType,
		ContentLength:   uploadParam.ContentLength,
		ContentEncoding: param.ContentEncoding,
		Next:            WrapUploadCursor(uploadParam.UUID, "next-chunk"),
	}, nil
}

func (s *assetService) handleArchiveExtraction(ctx context.Context, assetID AssetID, storageKey string) {

	err := s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusInProgress)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update extraction status", "error", err)
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}

	reader, err := s.storage.Get(ctx, storageKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get asset file", "error", err)
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read asset data", "error", err)
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}

	readerAt := bytes.NewReader(data)

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err = s.zipExtractor.Extract(ctx, assetID, readerAt, int64(len(data)))
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			slog.WarnContext(ctx, "Archive extraction attempt failed", "error", err, "attempt", i+1)
			time.Sleep(2 * time.Second)
			_, _ = readerAt.Seek(0, io.SeekStart)
		} else {
			slog.ErrorContext(ctx, "Failed to extract archive", "error", err, "attempts", maxRetries)
			_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
			return
		}
	}

	err = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusDone)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update extraction status to done", "error", err)
		return
	}

	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find asset after extraction", "error", err)
		return
	}

	slog.InfoContext(ctx, "Successfully extracted archive for asset", "assetID", assetID)

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

func (s *assetService) FindByID(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}
	return asset, nil
}

func (s *assetService) FindByUUID(ctx context.Context, uuid string, operator *AssetOperator) (*Asset, error) {
	if uuid == "" {
		return nil, ErrInvalidParameters
	}

	emptyGroupID := GroupID{}

	filter := AssetFilter{
		Keyword: uuid,
	}

	pagination := Pagination{
		Offset: 0,
		Limit:  1,
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

func (s *assetService) FindByIDs(ctx context.Context, ids []AssetID, operator *AssetOperator) ([]*Asset, error) {
	assets, err := s.assetRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func (s *assetService) FindByProject(ctx context.Context, groupID GroupID, filter AssetFilter, operator *AssetOperator) ([]*Asset, *PageInfo, error) {

	pagination := Pagination{
		Offset: 0,
		Limit:  DefaultPaginationLimit,
	}

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

func (s *assetService) FindFileByID(ctx context.Context, id AssetID, operator *AssetOperator) (*File, error) {
	return s.GetAssetFile(ctx, id)
}

func (s *assetService) FindFilesByIDs(ctx context.Context, ids AssetIDList, operator *AssetOperator) (map[AssetID]*File, error) {
	result := make(map[AssetID]*File)

	for _, id := range ids {
		file, err := s.GetAssetFile(ctx, id)
		if err != nil {
			continue
		}
		result[id] = file
	}

	return result, nil
}

func (s *assetService) DownloadByID(ctx context.Context, id AssetID, headers map[string]string, operator *AssetOperator) (io.ReadCloser, map[string]string, error) {
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

func (s *assetService) Create(ctx context.Context, param CreateAssetParam, operator *AssetOperator) (*Asset, *File, error) {

	if operator.AcOperator.User == nil && operator.Integration == (IntegrationID{}) {
		return nil, nil, errors.New("invalid operator")
	}

	asset, err := s.CreateAsset(ctx, param)
	if err != nil {
		return nil, nil, err
	}

	file, err := s.GetAssetFile(ctx, asset.ID)
	if err != nil {
		return asset, nil, err
	}

	return asset, file, nil
}

func (s *assetService) Update(ctx context.Context, param UpdateAssetParam, operator *AssetOperator) (*Asset, error) {
	return s.UpdateAsset(ctx, param)
}

func (s *assetService) UpdateFiles(ctx context.Context, id AssetID, status *ExtractionStatus, operator *AssetOperator) (*Asset, error) {
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

func (s *assetService) Delete(ctx context.Context, id AssetID, operator *AssetOperator) (AssetID, error) {
	err := s.DeleteAsset(ctx, id)
	if err != nil {
		return AssetID{}, err
	}
	return id, nil
}

func (s *assetService) BatchDelete(ctx context.Context, ids AssetIDList, operator *AssetOperator) ([]AssetID, error) {
	idArray := []AssetID(ids)
	err := s.DeleteAssets(ctx, idArray)
	if err != nil {
		return nil, err
	}
	return idArray, nil
}

func (s *assetService) Decompress(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error) {
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

func (s *assetService) Publish(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)

	publicURL, err := s.storage.GenerateURL(ctx, storageKey, 10*365*24*time.Hour)
	if err != nil {
		return nil, ErrStorageFailure
	}

	asset.URL = publicURL

	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *assetService) Unpublish(ctx context.Context, id AssetID, operator *AssetOperator) (*Asset, error) {
	asset, err := s.assetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, ErrAssetNotFound
	}

	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)

	temporaryURL, err := s.storage.GenerateURL(ctx, storageKey, 24*time.Hour)
	if err != nil {
		return nil, ErrStorageFailure
	}

	asset.URL = temporaryURL

	if err := s.assetRepo.Save(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *assetService) CreateUpload(ctx context.Context, param CreateAssetUploadParam, operator *AssetOperator) (*AssetUpload, error) {
	info, err := s.CreateAssetUpload(ctx, param)
	if err != nil {
		return nil, err
	}

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

func (s *assetService) RetryDecompression(ctx context.Context, id string) error {
	assetID, err := AssetIDFrom(id)
	if err != nil {
		return err
	}

	asset, err := s.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotFound
	}

	if asset.ArchiveExtractionStatus == nil || *asset.ArchiveExtractionStatus != ExtractionStatusFailed {
		return fmt.Errorf("cannot retry decompression, current status: %v", asset.ArchiveExtractionStatus)
	}

	status := ExtractionStatusPending
	if err := s.assetRepo.UpdateExtractionStatus(ctx, assetID, status); err != nil {
		return err
	}

	storageKey := path.Join(asset.GroupID.String(), asset.UUID, asset.FileName)
	go s.handleArchiveExtraction(context.Background(), assetID, storageKey)

	return nil
}

func detectAndUpdatePreviewType(ctx context.Context, s *assetService, asset *Asset) {
	files, err := s.storage.ListFiles(ctx, path.Join(asset.GroupID.String(), asset.UUID))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to list extracted files", "error", err)
		return
	}

	var previewType PreviewType
	for _, file := range files {
		filename := path.Base(file)
		ext := strings.ToLower(path.Ext(filename))

		if filename == "tileset.json" {
			previewType = PreviewTypeGeo3DTiles
			break
		}

		if ext == ".mvt" {
			previewType = PreviewTypeGeoMVT
			break
		}

		if ext == ".geojson" {
			previewType = PreviewTypeGeo
			break
		}
	}

	if previewType != "" && previewType != asset.PreviewType {
		slog.InfoContext(ctx, "Updating asset preview type", "assetID", asset.ID, "from", asset.PreviewType, "to", previewType)
		asset.PreviewType = previewType
		if err := s.assetRepo.Save(ctx, asset); err != nil {
			slog.ErrorContext(ctx, "Failed to update asset preview type", "error", err)
		}
	}
}
