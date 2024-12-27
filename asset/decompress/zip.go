package decompress

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/reearth/reearthx/asset"
)

type ZipDecompressor struct {
	assetService *asset.Service
}

func NewZipDecompressor(assetService *asset.Service) *ZipDecompressor {
	return &ZipDecompressor{
		assetService: assetService,
	}
}

func (d *ZipDecompressor) DecompressAsync(ctx context.Context, assetID asset.ID) error {
	// Get the zip file from asset service
	zipFile, err := d.assetService.GetFile(ctx, assetID)
	if err != nil {
		return fmt.Errorf("failed to get zip file: %w", err)
	}
	defer zipFile.Close()

	// Read all content to buffer for zip reader
	content, err := io.ReadAll(zipFile)
	if err != nil {
		return fmt.Errorf("failed to read zip content: %w", err)
	}

	// Create zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Update asset status to EXTRACTING
	_, err = d.assetService.Update(ctx, assetID, asset.UpdateAssetInput{
		Status: asset.StatusExtracting,
	})
	if err != nil {
		return fmt.Errorf("failed to update asset status: %w", err)
	}

	// Start async processing
	go func() {
		if err := d.processZipFile(ctx, zipReader); err != nil {
			// Update status to ERROR if processing fails
			d.assetService.Update(ctx, assetID, asset.UpdateAssetInput{
				Status: asset.StatusError,
				Error:  err.Error(),
			})
		} else {
			// Update status to ACTIVE if processing succeeds
			d.assetService.Update(ctx, assetID, asset.UpdateAssetInput{
				Status: asset.StatusActive,
			})
		}
	}()

	return nil
}

func (d *ZipDecompressor) processZipFile(ctx context.Context, zipReader *zip.Reader) error {
	for _, f := range zipReader.File {
		// Skip directories and hidden files
		if f.FileInfo().IsDir() || isHiddenFile(f.Name) {
			continue
		}

		// Open file in zip
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		// Create new asset for the file
		input := asset.CreateAssetInput{
			Name:        filepath.Base(f.Name),
			Size:        int64(f.UncompressedSize64),
			ContentType: detectContentType(f.Name),
		}

		newAsset, err := d.assetService.Create(ctx, input)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create asset: %w", err)
		}

		// Upload file content
		if err := d.assetService.Upload(ctx, newAsset.ID, rc); err != nil {
			rc.Close()
			return fmt.Errorf("failed to upload file content: %w", err)
		}

		rc.Close()
	}

	return nil
}

func isHiddenFile(name string) bool {
	base := filepath.Base(name)
	return len(base) > 0 && base[0] == '.'
}

func detectContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
