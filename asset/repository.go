// pkg/asset/repository.go
package asset

import (
	"context"
	"io"
)

type Repository interface {
	Fetch(ctx context.Context, id ID) (*Asset, error)
	FetchFile(ctx context.Context, id ID) (io.ReadCloser, error)
	Save(ctx context.Context, asset *Asset) error
	Remove(ctx context.Context, id ID) error
	Upload(ctx context.Context, id ID, file io.Reader) error
	GetUploadURL(ctx context.Context, id ID) (string, error)
}
