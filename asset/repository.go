package asset

import (
	"context"
	"io"
)

type Reader interface {
	Read(ctx context.Context, id ID) (*Asset, error)
	List(ctx context.Context) ([]*Asset, error)
}

type Writer interface {
	Create(ctx context.Context, asset *Asset) error
	Update(ctx context.Context, asset *Asset) error
	Delete(ctx context.Context, id ID) error
}

type FileOperator interface {
	Upload(ctx context.Context, id ID, file io.Reader) error
	FetchFile(ctx context.Context, id ID) (io.ReadCloser, error)
	GetUploadURL(ctx context.Context, id ID) (string, error)
}

type Repository interface {
	Reader
	Writer
	FileOperator
}
