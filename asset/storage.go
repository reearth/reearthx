package asset

import (
	"context"
	"io"
	"time"
)

type Storage interface {
	Save(ctx context.Context, key string, data io.Reader, size int64, contentType string, contentEncoding string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GenerateURL(ctx context.Context, key string, expires time.Duration) (string, error)
	GenerateUploadURL(ctx context.Context, key string, size int64, contentType string, contentEncoding string, expires time.Duration) (string, error)
}
