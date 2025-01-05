package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"cloud.google.com/go/storage"
	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
	"google.golang.org/api/iterator"
)

const (
	errFailedToCreateClient = "failed to create client: %w"
	errAssetAlreadyExists   = "asset already exists: %s"
	errAssetNotFound        = "asset not found: %s"
	errFailedToUpdateAsset  = "failed to update asset: %w"
	errFailedToDeleteAsset  = "failed to delete asset: %w"
	errFailedToListAssets   = "failed to list assets: %w"
	errFailedToUploadFile   = "failed to upload file: %w"
	errFailedToCloseWriter  = "failed to close writer: %w"
	errFailedToReadFile     = "failed to read file: %w"
	errFailedToGetAsset     = "failed to get asset: %w"
	errFailedToGenerateURL  = "failed to generate upload URL: %w"
)

type Client struct {
	bucket     *storage.BucketHandle
	bucketName string
	basePath   string
}

var _ repository.PersistenceRepository = (*Client)(nil)

func NewClient(ctx context.Context, bucketName string, basePath string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(errFailedToCreateClient, err)
	}

	return &Client{
		bucket:     client.Bucket(bucketName),
		bucketName: bucketName,
		basePath:   basePath,
	}, nil
}

func (c *Client) Create(ctx context.Context, asset *domain.Asset) error {
	obj := c.getObject(asset.ID())
	attrs := storage.ObjectAttrs{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Attrs(ctx); err == nil {
		return fmt.Errorf(errAssetAlreadyExists, asset.ID())
	}

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs = attrs
	return writer.Close()
}

func (c *Client) Read(ctx context.Context, id domain.ID) (*domain.Asset, error) {
	attrs, err := c.getObject(id).Attrs(ctx)
	if err != nil {
		return nil, c.handleNotFound(err, id)
	}

	asset := domain.NewAsset(
		id,
		attrs.Metadata["name"],
		attrs.Size,
		attrs.ContentType,
	)

	return asset, nil
}

func (c *Client) Update(ctx context.Context, asset *domain.Asset) error {
	obj := c.getObject(asset.ID())
	update := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Update(ctx, update); err != nil {
		return fmt.Errorf(errFailedToUpdateAsset, err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, id domain.ID) error {
	obj := c.getObject(id)
	if err := obj.Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf(errFailedToDeleteAsset, err)
	}
	return nil
}

func (c *Client) List(ctx context.Context) ([]*domain.Asset, error) {
	var assets []*domain.Asset
	it := c.bucket.Objects(ctx, &storage.Query{Prefix: c.basePath})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf(errFailedToListAssets, err)
		}

		asset := domain.NewAsset(
			domain.ID(path.Base(attrs.Name)),
			attrs.Metadata["name"],
			attrs.Size,
			attrs.ContentType,
		)
		assets = append(assets, asset)
	}

	return assets, nil
}

func (c *Client) Upload(ctx context.Context, id domain.ID, content io.Reader) error {
	obj := c.getObject(id)
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, content); err != nil {
		err := writer.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf(errFailedToUploadFile, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf(errFailedToCloseWriter, err)
	}
	return nil
}

func (c *Client) Download(ctx context.Context, id domain.ID) (io.ReadCloser, error) {
	obj := c.getObject(id)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, fmt.Errorf(errAssetNotFound, id)
		}
		return nil, fmt.Errorf(errFailedToReadFile, err)
	}
	return reader, nil
}

func (c *Client) GetUploadURL(ctx context.Context, id domain.ID) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(15 * time.Minute),
	}

	url, err := c.bucket.SignedURL(c.objectPath(id), opts)
	if err != nil {
		return "", fmt.Errorf(errFailedToGenerateURL, err)
	}
	return url, nil
}

func (c *Client) getObject(id domain.ID) *storage.ObjectHandle {
	return c.bucket.Object(c.objectPath(id))
}

func (c *Client) objectPath(id domain.ID) string {
	return path.Join(c.basePath, id.String())
}

func (c *Client) handleNotFound(err error, id domain.ID) error {
	if errors.Is(err, storage.ErrObjectNotExist) {
		return fmt.Errorf(errAssetNotFound, id)
	}
	return fmt.Errorf(errFailedToGetAsset, err)
}
