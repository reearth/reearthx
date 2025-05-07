package asset

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestZipContent() []byte {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	testFile, err := zipWriter.Create("test.txt")
	if err == nil {
		_, _ = testFile.Write([]byte("This is test content"))
	}

	_ = zipWriter.Close()

	return buf.Bytes()
}

func TestZipExtractorExtract(t *testing.T) {
	ctx := context.Background()
	assetID := NewAssetID()
	zipContent := createTestZipContent()
	zipReader := bytes.NewReader(zipContent)
	zipSize := int64(len(zipContent))

	assetRepo := new(MockAssetRepository)
	storage := new(MockStorage)

	extractor := NewZipExtractor(assetRepo, storage)

	t.Run("Successful extraction", func(t *testing.T) {
		asset := &Asset{
			ID:        assetID,
			GroupID:   NewGroupID(),
			CreatedAt: time.Now(),
			UUID:      "test-uuid",
			FileName:  "test.zip",
		}

		assetRepo.On("FindByID", ctx, assetID).Return(asset, nil)
		assetRepo.On("UpdateExtractionStatus", ctx, assetID, ExtractionStatusInProgress).Return(nil)
		assetRepo.On("UpdateExtractionStatus", ctx, assetID, ExtractionStatusDone).Return(nil)

		storage.On("Save", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := extractor.Extract(ctx, assetID, zipReader, zipSize)

		assert.NoError(t, err)
		assetRepo.AssertExpectations(t)
		storage.AssertExpectations(t)
	})

	t.Run("Asset not found", func(t *testing.T) {
		assetRepo.ExpectedCalls = nil
		assetRepo.On("FindByID", ctx, assetID).Return(nil, nil)

		err := extractor.Extract(ctx, assetID, zipReader, zipSize)

		assert.Error(t, err)
		assert.Equal(t, ErrAssetNotFound, err)
		assetRepo.AssertExpectations(t)
	})

	t.Run("Extraction error handling", func(t *testing.T) {
		assetRepo.ExpectedCalls = nil
		storage.ExpectedCalls = nil

		asset := &Asset{
			ID:        assetID,
			GroupID:   NewGroupID(),
			CreatedAt: time.Now(),
			UUID:      "test-uuid",
			FileName:  "test.zip",
		}

		assetRepo.On("FindByID", ctx, assetID).Return(asset, nil)
		assetRepo.On("UpdateExtractionStatus", ctx, assetID, ExtractionStatusInProgress).Return(nil)
		assetRepo.On("UpdateExtractionStatus", ctx, assetID, ExtractionStatusFailed).Return(nil)

		corruptedContent := []byte("corrupted zip content")
		corruptedReader := bytes.NewReader(corruptedContent)

		err := extractor.Extract(ctx, assetID, corruptedReader, int64(len(corruptedContent)))

		assert.Error(t, err)
		assetRepo.AssertExpectations(t)
	})
}

func TestShouldExtractArchive(t *testing.T) {
	testCases := []struct {
		name          string
		fileName      string
		contentType   string
		shouldExtract bool
	}{
		{
			name:          "ZIP file by extension",
			fileName:      "archive.zip",
			contentType:   "application/zip",
			shouldExtract: true,
		},
		{
			name:          "ZIP file by content type only",
			fileName:      "archive",
			contentType:   "application/zip",
			shouldExtract: true,
		},
		{
			name:          "ZIP file with uppercase extension",
			fileName:      "archive.ZIP",
			contentType:   "application/octet-stream",
			shouldExtract: true,
		},
		{
			name:          "TAR.GZ file",
			fileName:      "archive.tar.gz",
			contentType:   "application/gzip",
			shouldExtract: true,
		},
		{
			name:          "TAR file",
			fileName:      "archive.tar",
			contentType:   "application/x-tar",
			shouldExtract: true,
		},
		{
			name:          "7z file",
			fileName:      "archive.7z",
			contentType:   "application/x-7z-compressed",
			shouldExtract: true,
		},
		{
			name:          "Non-archive file",
			fileName:      "document.pdf",
			contentType:   "application/pdf",
			shouldExtract: false,
		},
		{
			name:          "Text file",
			fileName:      "readme.txt",
			contentType:   "text/plain",
			shouldExtract: false,
		},
		{
			name:          "Image file",
			fileName:      "photo.jpg",
			contentType:   "image/jpeg",
			shouldExtract: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := shouldExtractArchive(tc.fileName, tc.contentType)
			assert.Equal(t, tc.shouldExtract, result)
		})
	}
}
