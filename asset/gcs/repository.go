package gcs

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"cloud.google.com/go/storage"
	"github.com/reearth/reearthx/asset"
	"google.golang.org/api/iterator"
)

type Repository struct {
	bucket     *storage.BucketHandle
	bucketName string
	basePath   string
}

func NewRepository(ctx context.Context, bucketName string) (*Repository, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Repository{
		bucket:     client.Bucket(bucketName),
		bucketName: bucketName,
		basePath:   "assets",
	}, nil
}

func (r *Repository) Create(ctx context.Context, asset *asset.Asset) error {
	obj := r.bucket.Object(r.objectPath(asset.ID))
	attrs := storage.ObjectAttrs{
		Metadata: map[string]string{
			"name":         asset.Name,
			"content_type": asset.ContentType,
		},
	}

	if _, err := obj.Attrs(ctx); err == nil {
		return fmt.Errorf("asset already exists: %s", asset.ID)
	}

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs = attrs
	return writer.Close()
}

func (r *Repository) Read(ctx context.Context, id asset.ID) (*asset.Asset, error) {
	obj := r.bucket.Object(r.objectPath(id))
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, fmt.Errorf("asset not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return &asset.Asset{
		ID:          id,
		Name:        attrs.Metadata["name"],
		Size:        attrs.Size,
		ContentType: attrs.ContentType,
		CreatedAt:   attrs.Created,
		UpdatedAt:   attrs.Updated,
	}, nil
}

func (r *Repository) Update(ctx context.Context, asset *asset.Asset) error {
	obj := r.bucket.Object(r.objectPath(asset.ID))
	update := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"name":         asset.Name,
			"content_type": asset.ContentType,
		},
	}

	if _, err := obj.Update(ctx, update); err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id asset.ID) error {
	obj := r.bucket.Object(r.objectPath(id))
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("failed to delete asset: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]*asset.Asset, error) {
	var assets []*asset.Asset
	it := r.bucket.Objects(ctx, &storage.Query{Prefix: r.basePath})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list assets: %w", err)
		}

		assets = append(assets, &asset.Asset{
			ID:          asset.ID(path.Base(attrs.Name)),
			Name:        attrs.Metadata["name"],
			Size:        attrs.Size,
			ContentType: attrs.ContentType,
			CreatedAt:   attrs.Created,
			UpdatedAt:   attrs.Updated,
		})
	}

	return assets, nil
}

func (r *Repository) Upload(ctx context.Context, id asset.ID, file io.Reader) error {
	obj := r.bucket.Object(r.objectPath(id))
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, file); err != nil {
		writer.Close()
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}

func (r *Repository) FetchFile(ctx context.Context, id asset.ID) (io.ReadCloser, error) {
	obj := r.bucket.Object(r.objectPath(id))
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, fmt.Errorf("asset not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return reader, nil
}

func (r *Repository) GetUploadURL(ctx context.Context, id asset.ID) (string, error) {
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

func (r *Repository) objectPath(id asset.ID) string {
	return path.Join(r.basePath, id.String())
}
