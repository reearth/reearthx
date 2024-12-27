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
	assetService asset.Service
}

// NewZipDecompressor creates a new zip decompressor
func NewZipDecompressor(assetService asset.Service) Decompressor {
	return &ZipDecompressor{
		assetService: assetService,
	}
}

// DecompressAsync implements Decompressor interface
func (d *ZipDecompressor) DecompressAsync(ctx context.Context, assetID asset.ID) error {
	zipContent, err := d.fetchZipContent(ctx, assetID)
	if err != nil {
		return err
	}

	zipReader, err := d.createZipReader(zipContent)
	if err != nil {
		return err
	}

	if err := d.updateAssetStatus(ctx, assetID, asset.StatusExtracting); err != nil {
		return err
	}

	go d.processZipAsync(ctx, assetID, zipReader)

	return nil
}

func (d *ZipDecompressor) fetchZipContent(ctx context.Context, assetID asset.ID) ([]byte, error) {
	zipFile, err := d.assetService.GetFile(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zip file: %w", err)
	}
	defer zipFile.Close()

	content, err := io.ReadAll(zipFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip content: %w", err)
	}

	return content, nil
}

func (d *ZipDecompressor) createZipReader(content []byte) (*zip.Reader, error) {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}
	return reader, nil
}

func (d *ZipDecompressor) updateAssetStatus(ctx context.Context, assetID asset.ID, status asset.Status) error {
	_, err := d.assetService.Update(ctx, assetID, asset.UpdateAssetInput{
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to update asset status: %w", err)
	}
	return nil
}

func (d *ZipDecompressor) processZipAsync(ctx context.Context, assetID asset.ID, zipReader *zip.Reader) {
	if err := d.processZipContents(ctx, zipReader); err != nil {
		d.updateAssetStatus(ctx, assetID, asset.StatusError)
		return
	}
	d.updateAssetStatus(ctx, assetID, asset.StatusActive)
}

func (d *ZipDecompressor) processZipContents(ctx context.Context, zipReader *zip.Reader) error {
	for _, f := range zipReader.File {
		if err := d.processZipEntry(ctx, f); err != nil {
			return err
		}
	}
	return nil
}

func (d *ZipDecompressor) processZipEntry(ctx context.Context, f *zip.File) error {
	if f.FileInfo().IsDir() || isHiddenFile(f.Name) {
		return nil
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer rc.Close()

	newAsset, err := d.createAsset(ctx, f)
	if err != nil {
		return err
	}

	if err := d.assetService.Upload(ctx, newAsset.ID, rc); err != nil {
		return fmt.Errorf("failed to upload file content: %w", err)
	}

	return nil
}

func (d *ZipDecompressor) createAsset(ctx context.Context, f *zip.File) (*asset.Asset, error) {
	input := asset.CreateAssetInput{
		Name:        filepath.Base(f.Name),
		Size:        int64(f.UncompressedSize64),
		ContentType: detectContentType(f.Name),
	}

	newAsset, err := d.assetService.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return newAsset, nil
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
