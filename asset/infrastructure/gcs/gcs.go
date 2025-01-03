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

type Client struct {
	bucket     *storage.BucketHandle
	bucketName string
	basePath   string
}

var _ repository.Repository = (*Client)(nil)

func NewGCSClient(ctx context.Context, bucketName string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{
		bucket:     client.Bucket(bucketName),
		bucketName: bucketName,
		basePath:   "assets",
	}, nil
}

func (r *Client) Create(ctx context.Context, asset *domain.Asset) error {
	obj := r.getObject(asset.ID())
	attrs := storage.ObjectAttrs{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Attrs(ctx); err == nil {
		return fmt.Errorf("asset already exists: %s", asset.ID())
	}

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs = attrs
	return writer.Close()
}

func (r *Client) Read(ctx context.Context, id domain.ID) (*domain.Asset, error) {
	attrs, err := r.getObject(id).Attrs(ctx)
	if err != nil {
		return nil, r.handleNotFound(err, id)
	}

	asset := domain.NewAsset(
		id,
		attrs.Metadata["name"],
		attrs.Size,
		attrs.ContentType,
	)

	return asset, nil
}

func (r *Client) Update(ctx context.Context, asset *domain.Asset) error {
	obj := r.getObject(asset.ID())
	update := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Update(ctx, update); err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}
	return nil
}

func (r *Client) Delete(ctx context.Context, id domain.ID) error {
	obj := r.getObject(id)
	if err := obj.Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf("failed to delete asset: %w", err)
	}
	return nil
}

func (r *Client) List(ctx context.Context) ([]*domain.Asset, error) {
	var assets []*domain.Asset
	it := r.bucket.Objects(ctx, &storage.Query{Prefix: r.basePath})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list assets: %w", err)
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

func (r *Client) Upload(ctx context.Context, id domain.ID, content io.Reader) error {
	obj := r.getObject(id)
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, content); err != nil {
		err := writer.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

func (r *Client) Download(ctx context.Context, id domain.ID) (io.ReadCloser, error) {
	obj := r.getObject(id)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, fmt.Errorf("asset not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return reader, nil
}

func (r *Client) GetUploadURL(ctx context.Context, id domain.ID) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(15 * time.Minute),
	}

	url, err := r.bucket.SignedURL(r.objectPath(id), opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}
	return url, nil
}

func (r *Client) getObject(id domain.ID) *storage.ObjectHandle {
	return r.bucket.Object(r.objectPath(id))
}

func (r *Client) objectPath(id domain.ID) string {
	return path.Join(r.basePath, id.String())
}

func (r *Client) handleNotFound(err error, id domain.ID) error {
	if errors.Is(err, storage.ErrObjectNotExist) {
		return fmt.Errorf("asset not found: %s", id)
	}
	return fmt.Errorf("failed to get asset: %w", err)
}
