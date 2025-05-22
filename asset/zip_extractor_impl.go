package asset

import (
	"archive/zip"
	"context"
	"io"
	"path"
	"path/filepath"
	"strings"
)

var _ ZipExtractor = &zipExtractor{}

type zipExtractor struct {
	assetRepo AssetRepository
	storage   Storage
}

func NewZipExtractor(
	assetRepo AssetRepository,
	storage Storage,
) ZipExtractor {
	return &zipExtractor{
		assetRepo: assetRepo,
		storage:   storage,
	}
}

func (e *zipExtractor) Extract(ctx context.Context, assetID AssetID, reader io.ReaderAt, size int64) error {
	asset, err := e.assetRepo.FindByID(ctx, assetID)
	if err != nil {
		return err
	}
	if asset == nil {
		return ErrAssetNotFound
	}

	if err := e.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusInProgress); err != nil {
		return err
	}

	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		_ = e.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusFailed)
		return err
	}

	baseDir := path.Join(asset.GroupID().String(), asset.UUID(), "extracted")

	for _, zipFile := range zipReader.File {
		if !isValidZipPath(zipFile.Name) {
			continue
		}

		if zipFile.FileInfo().IsDir() {
			continue
		}

		storageKey := path.Join(baseDir, zipFile.Name)

		rc, err := zipFile.Open()
		if err != nil {
			continue
		}

		err = e.storage.Save(
			ctx,
			storageKey,
			rc,
			int64(zipFile.UncompressedSize64),
			detectContentType(zipFile.Name),
			"",
		)
		rc.Close()

		if err != nil {
			continue
		}
	}

	if err := e.assetRepo.UpdateExtractionStatus(ctx, assetID, ExtractionStatusDone); err != nil {
		return err
	}

	return nil
}

func isValidZipPath(filePath string) bool {
	cleanPath := filepath.Clean(filePath)

	if strings.HasPrefix(cleanPath, "../") || strings.HasPrefix(cleanPath, "..\\") {
		return false
	}

	if filepath.IsAbs(cleanPath) {
		return false
	}

	return true
}

func detectContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
