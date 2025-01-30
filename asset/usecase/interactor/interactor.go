package assetinteractor

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/infrastructure/decompress"
	"github.com/reearth/reearthx/asset/repository"
	assetusecase "github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/log"
)

type AssetInteractor struct {
	repo         repository.PersistenceRepository
	decompressor repository.Decompressor
	pubsub       repository.PubSubRepository
	txManager    assetusecase.TransactionManager
	jobRepo      repository.DecompressJobRepository
}

func NewAssetInteractor(
	repo repository.PersistenceRepository,
	pubsub repository.PubSubRepository,
	txManager assetusecase.TransactionManager,
	jobRepo repository.DecompressJobRepository,
) *AssetInteractor {
	return &AssetInteractor{
		repo:         repo,
		decompressor: decompress.NewZipDecompressor(),
		pubsub:       pubsub,
		txManager:    txManager,
		jobRepo:      jobRepo,
	}
}

// validateAsset validates an asset using domain rules
func (i *AssetInteractor) validateAsset(ctx context.Context, asset *entity.Asset) *assetusecase.Result {
	if result := asset.Validate(ctx); !result.IsValid {
		return assetusecase.NewValidationErrorResult(result.Errors)
	}
	return nil
}

// CreateAsset creates a new asset
func (i *AssetInteractor) CreateAsset(ctx context.Context, asset *entity.Asset) *assetusecase.Result {
	if validationResult := i.validateAsset(ctx, asset); validationResult != nil {
		return validationResult
	}

	var createdAsset *entity.Asset
	err := i.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := i.repo.Create(ctx, asset); err != nil {
			return err
		}

		if err := i.pubsub.PublishAssetCreated(ctx, asset); err != nil {
			log.Errorfc(ctx, "failed to publish asset created event: %v", err)
			return err
		}

		createdAsset = asset
		return nil
	})

	if err != nil {
		return assetusecase.NewErrorResult("CREATE_ASSET_FAILED", err.Error(), nil)
	}

	return assetusecase.NewResult(createdAsset)
}

// GetAsset retrieves an asset by ID
func (i *AssetInteractor) GetAsset(ctx context.Context, id id.ID) *assetusecase.Result {
	asset, err := i.repo.Read(ctx, id)
	if err != nil {
		return assetusecase.NewErrorResult("GET_ASSET_FAILED", err.Error(), nil)
	}
	return assetusecase.NewResult(asset)
}

// UpdateAsset updates an existing asset
func (i *AssetInteractor) UpdateAsset(ctx context.Context, asset *entity.Asset) *assetusecase.Result {
	if validationResult := i.validateAsset(ctx, asset); validationResult != nil {
		return validationResult
	}

	var updatedAsset *entity.Asset
	err := i.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := i.repo.Update(ctx, asset); err != nil {
			return err
		}

		if err := i.pubsub.PublishAssetUpdated(ctx, asset); err != nil {
			log.Errorfc(ctx, "failed to publish asset updated event: %v", err)
			return err
		}

		updatedAsset = asset
		return nil
	})

	if err != nil {
		return assetusecase.NewErrorResult("UPDATE_ASSET_FAILED", err.Error(), nil)
	}

	return assetusecase.NewResult(updatedAsset)
}

// DeleteAsset removes an asset by ID
func (i *AssetInteractor) DeleteAsset(ctx context.Context, id id.ID) *assetusecase.Result {
	err := i.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := i.repo.Delete(ctx, id); err != nil {
			return err
		}

		if err := i.pubsub.PublishAssetDeleted(ctx, id); err != nil {
			log.Errorfc(ctx, "failed to publish asset deleted event: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		return assetusecase.NewErrorResult("DELETE_ASSET_FAILED", err.Error(), nil)
	}

	return assetusecase.NewResult(nil)
}

// UploadAssetContent uploads content for an asset with the given ID
func (i *AssetInteractor) UploadAssetContent(ctx context.Context, id id.ID, content io.Reader) *assetusecase.Result {
	var uploadedAsset *entity.Asset
	err := i.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := i.repo.Upload(ctx, id, content); err != nil {
			return err
		}

		asset, err := i.repo.Read(ctx, id)
		if err != nil {
			return err
		}

		if err := i.pubsub.PublishAssetUploaded(ctx, asset); err != nil {
			log.Errorfc(ctx, "failed to publish asset uploaded event: %v", err)
			return err
		}

		uploadedAsset = asset
		return nil
	})

	if err != nil {
		return assetusecase.NewErrorResult("UPLOAD_CONTENT_FAILED", err.Error(), nil)
	}

	return assetusecase.NewResult(uploadedAsset)
}

// DownloadAssetContent retrieves the content of an asset by ID
func (i *AssetInteractor) DownloadAssetContent(ctx context.Context, id id.ID) *assetusecase.Result {
	content, err := i.repo.Download(ctx, id)
	if err != nil {
		return assetusecase.NewErrorResult("DOWNLOAD_CONTENT_FAILED", err.Error(), nil)
	}
	return assetusecase.NewResult(content)
}

// GetAssetUploadURL generates a URL for uploading content to an asset
func (i *AssetInteractor) GetAssetUploadURL(ctx context.Context, id id.ID) *assetusecase.Result {
	url, err := i.repo.GetUploadURL(ctx, id)
	if err != nil {
		return assetusecase.NewErrorResult("GET_UPLOAD_URL_FAILED", err.Error(), nil)
	}
	return assetusecase.NewResult(url)
}

// ListAssets returns all assets
func (i *AssetInteractor) ListAssets(ctx context.Context) *assetusecase.Result {
	assets, err := i.repo.List(ctx)
	if err != nil {
		return assetusecase.NewErrorResult("LIST_ASSETS_FAILED", err.Error(), nil)
	}
	return assetusecase.NewResult(assets)
}

// DecompressZipContent decompresses zip content and returns a channel of decompressed files
func (i *AssetInteractor) DecompressZipContent(ctx context.Context, content []byte) *assetusecase.Result {
	ch, err := i.decompressor.DecompressWithContent(ctx, content)
	if err != nil {
		return assetusecase.NewErrorResult("DECOMPRESS_FAILED", err.Error(), nil)
	}

	if assetID, ok := ctx.Value("assetID").(id.ID); ok {
		jobID := uuid.New().String()
		status := &assetusecase.DecompressStatus{
			JobID:     jobID,
			AssetID:   assetID,
			Status:    "pending",
			Progress:  0,
			StartedAt: time.Now(),
		}

		err := i.jobRepo.Save(ctx, status)
		if err != nil {
			return assetusecase.NewErrorResult("JOB_CREATION_FAILED", err.Error(), nil)
		}

		go func() {
			ctx := context.Background()
			status.Status = "processing"
			if err := i.jobRepo.Save(ctx, status); err != nil {
				log.Errorfc(ctx, "failed to save job status: %v", err)
			}

			err := i.txManager.WithTransaction(ctx, func(ctx context.Context) error {
				asset, err := i.repo.Read(ctx, assetID)
				if err != nil {
					return err
				}

				asset.UpdateStatus(entity.StatusExtracting, "")
				if err := i.repo.Update(ctx, asset); err != nil {
					return err
				}

				if err := i.pubsub.PublishAssetExtracted(ctx, asset); err != nil {
					log.Errorfc(ctx, "failed to publish asset extracted event: %v", err)
					return err
				}

				return nil
			})

			if err != nil {
				status.Status = "failed"
				status.Error = err.Error()
				if err := i.jobRepo.Save(ctx, status); err != nil {
					log.Errorfc(ctx, "failed to save job status: %v", err)
				}
				return
			}

			// Process decompressed files
			totalFiles := 0
			processedFiles := 0
			for range ch {
				totalFiles++
			}

			for range ch {
				processedFiles++
				progress := float64(processedFiles) / float64(totalFiles) * 100
				if err := i.jobRepo.UpdateProgress(ctx, jobID, progress); err != nil {
					log.Errorfc(ctx, "failed to update job progress: %v", err)
				}
			}

			status.Status = "completed"
			status.Progress = 100
			status.CompletedAt = time.Now()
			if err := i.jobRepo.Save(ctx, status); err != nil {
				log.Errorfc(ctx, "failed to save job status: %v", err)
			}
		}()

		return assetusecase.NewResult(map[string]interface{}{
			"jobID": jobID,
			"ch":    ch,
		})
	}

	return assetusecase.NewResult(ch)
}

// CompressToZip compresses the provided files into a zip archive
func (i *AssetInteractor) CompressToZip(ctx context.Context, files map[string]io.Reader) *assetusecase.Result {
	ch, err := i.decompressor.CompressWithContent(ctx, files)
	if err != nil {
		return assetusecase.NewErrorResult("COMPRESS_FAILED", err.Error(), nil)
	}
	return assetusecase.NewResult(ch)
}

// DeleteAllAssetsInGroup deletes all assets in a group
func (i *AssetInteractor) DeleteAllAssetsInGroup(ctx context.Context, groupID id.GroupID) *assetusecase.Result {
	err := i.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		assets, err := i.repo.FindByGroup(ctx, groupID)
		if err != nil {
			return err
		}

		for _, asset := range assets {
			if err := i.DeleteAsset(ctx, asset.ID()).GetError(); err != nil {
				log.Errorfc(ctx, "failed to delete asset %s in group %s: %v", asset.ID(), groupID, err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		return assetusecase.NewErrorResult("DELETE_GROUP_ASSETS_FAILED", err.Error(), nil)
	}

	return assetusecase.NewResult(nil)
}

// DeliverAsset implements the asset delivery functionality
func (i *AssetInteractor) DeliverAsset(ctx context.Context, id id.ID, options *assetusecase.DeliverOptions) *assetusecase.Result {
	// Get asset metadata
	asset, err := i.repo.Read(ctx, id)
	if err != nil {
		return assetusecase.NewErrorResult("ASSET_NOT_FOUND", err.Error(), nil)
	}

	// Get content
	content, err := i.repo.Download(ctx, id)
	if err != nil {
		return assetusecase.NewErrorResult("CONTENT_DOWNLOAD_FAILED", err.Error(), nil)
	}

	// Apply transformations if needed
	if options != nil && options.Transform {
		// TODO: Implement transformation logic when needed
		log.Infofc(ctx, "Asset transformation requested but not implemented yet")
	}

	// Prepare response metadata
	contentType := asset.ContentType()
	if options != nil && options.ContentType != "" {
		contentType = options.ContentType
	}

	headers := map[string]string{
		"Content-Type": contentType,
	}

	if options != nil {
		// Add cache control
		if options.MaxAge > 0 {
			headers["Cache-Control"] = fmt.Sprintf("max-age=%d", options.MaxAge)
		}

		// Add content disposition
		if options.Disposition != "" {
			headers["Content-Disposition"] = options.Disposition
		}

		// Add custom headers
		for k, v := range options.Headers {
			headers[k] = v
		}
	}

	return assetusecase.NewResult(map[string]interface{}{
		"content": content,
		"headers": headers,
	})
}

// GetDecompressStatus implements the decompress status retrieval
func (i *AssetInteractor) GetDecompressStatus(ctx context.Context, jobID string) *assetusecase.Result {
	status, err := i.jobRepo.Get(ctx, jobID)
	if err != nil {
		return assetusecase.NewErrorResult("JOB_NOT_FOUND", err.Error(), nil)
	}
	return assetusecase.NewResult(status)
}

// ListDecompressJobs implements the decompress jobs listing
func (i *AssetInteractor) ListDecompressJobs(ctx context.Context) *assetusecase.Result {
	jobs, err := i.jobRepo.List(ctx)
	if err != nil {
		return assetusecase.NewErrorResult("LIST_JOBS_FAILED", err.Error(), nil)
	}
	return assetusecase.NewResult(jobs)
}
