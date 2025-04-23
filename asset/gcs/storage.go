package gcs

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/reearth/reearthx/asset"
	"google.golang.org/api/option"
)

type Storage struct {
	client     *storage.Client
	bucket     *storage.BucketHandle
	bucketName string
	baseURL    string
}

func NewStorage(ctx context.Context, bucketName, baseURL string, opts ...option.ClientOption) (asset.Storage, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	bucket := client.Bucket(bucketName)

	_, err = bucket.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to access GCS bucket: %w", err)
	}

	return &Storage{
		client:     client,
		bucket:     bucket,
		bucketName: bucketName,
		baseURL:    baseURL,
	}, nil
}

func (s *Storage) Save(ctx context.Context, key string, data io.Reader, size int64, contentType string, contentEncoding string) error {
	obj := s.bucket.Object(key)
	w := obj.NewWriter(ctx)

	w.ContentType = contentType
	if contentEncoding != "" {
		w.ContentEncoding = contentEncoding
	}

	if _, err := io.Copy(w, data); err != nil {
		w.Close()
		return fmt.Errorf("failed to copy data to GCS: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	obj := s.bucket.Object(key)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open GCS object: %w", err)
	}

	return r, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	obj := s.bucket.Object(key)

	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("failed to delete GCS object: %w", err)
	}

	return nil
}

func (s *Storage) GenerateURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	expiration := time.Now().Add(expires)

	url, err := storage.SignedURL(s.bucketName, key, &storage.SignedURLOptions{
		Method:      "GET",
		Expires:     expiration,
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

func (s *Storage) GenerateUploadURL(ctx context.Context, key string, size int64, contentType string, contentEncoding string, expires time.Duration) (string, error) {
	expiration := time.Now().Add(expires)

	opts := &storage.SignedURLOptions{
		Method:      "PUT",
		Expires:     expiration,
		ContentType: contentType,
	}

	if contentEncoding != "" {
		opts.Headers = []string{
			fmt.Sprintf("Content-Encoding: %s", contentEncoding),
		}
	}

	url, err := storage.SignedURL(s.bucketName, key, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return url, nil
}

func (s *Storage) Close() error {
	return s.client.Close()
}
