package domain

import (
	"context"
	"io"
)

type Service interface {
	Create(ctx context.Context, asset *Asset) error
	Read(ctx context.Context, id ID) (*Asset, error)
	Update(ctx context.Context, asset *Asset) error
	Delete(ctx context.Context, id ID) error
	List(ctx context.Context) ([]*Asset, error)
	Upload(ctx context.Context, id ID, content io.Reader) error
	Download(ctx context.Context, id ID) (io.ReadCloser, error)
}

type Decompressor interface {
	DecompressAsync(ctx context.Context, assetID ID) error
}
