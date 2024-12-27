package decompress

import (
	"archive/zip"
	"context"
	"io"

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
		return err
	}
	defer zipFile.Close()

	// Create zip reader
	zipReader, err := zip.NewReader(zipFile.(io.ReaderAt), -1)
	if err != nil {
		return err
	}

	// Update asset status to EXTRACTING
	_, err = d.assetService.Update(ctx, assetID, asset.UpdateAssetInput{
		Status: asset.StatusExtracting,
	})
	if err != nil {
		return err
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
	// Process each file in the zip
	// Create new assets for each file
	return nil
}
