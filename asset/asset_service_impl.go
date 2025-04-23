package asset

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAssetNotFound     = errors.New("asset not found")
	ErrGroupNotFound     = errors.New("group not found")
	ErrInvalidParameters = errors.New("invalid parameters")
	ErrStorageFailure    = errors.New("storage operation failed")
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

func (s *assetService) GetAssetFile(ctx context.Context, id AssetID) (*AssetFile, error) {
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

	assetFile := &AssetFile{
		Name:            asset.FileName,
		Size:            asset.Size,
		ContentType:     asset.ContentType,
		ContentEncoding: asset.ContentEncoding,
		Path:            storageKey,
	}

	return assetFile, nil
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

	assetUUID := uuid.New().String()

	contentType := s.fileProcessor.DetectContentType(param.FileName, nil)

	storageKey := path.Join(param.GroupID.String(), assetUUID, param.FileName)

	uploadURL, err := s.storage.GenerateUploadURL(
		ctx,
		storageKey,
		param.ContentLength,
		contentType,
		param.ContentEncoding,
		1*time.Hour,
	)
	if err != nil {
		return nil, ErrStorageFailure
	}

	token := uuid.New().String()

	return &AssetUploadInfo{
		Token:           token,
		URL:             uploadURL,
		ContentType:     contentType,
		ContentLength:   param.ContentLength,
		ContentEncoding: param.ContentEncoding,
		Next:            param.Cursor,
	}, nil
}

func (s *assetService) handleArchiveExtraction(ctx context.Context, assetID AssetID, storageKey string) {
	err := s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusInProgress)
	if err != nil {
		return
	}

	reader, err := s.storage.Get(ctx, storageKey)
	if err != nil {
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}
	defer func(reader io.ReadCloser) {
		_ = reader.Close()
	}(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}

	readerAt := bytes.NewReader(data)

	err = s.zipExtractor.Extract(ctx, assetID, readerAt, int64(len(data)))
	if err != nil {
		_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return
	}

	_ = s.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusDone)
}

func shouldExtractArchive(fileName string, contentType string) bool {
	lowerFileName := strings.ToLower(fileName)

	// Check file extensions
	if strings.HasSuffix(lowerFileName, ".zip") ||
		strings.HasSuffix(lowerFileName, ".tar.gz") ||
		strings.HasSuffix(lowerFileName, ".tar") ||
		strings.HasSuffix(lowerFileName, ".7z") {
		return true
	}

	// Check content types
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
