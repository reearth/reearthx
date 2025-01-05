package repository

import (
	"context"
	"io"
)

// Decompressor defines the interface for compression and decompression operations
type Decompressor interface {
	// DecompressWithContent decompresses zip content directly and returns a channel of decompressed files
	// The channel will be closed when all files have been processed or an error occurs
	DecompressWithContent(ctx context.Context, content []byte) (<-chan DecompressedFile, error)

	// CompressWithContent compresses the provided content into a zip archive
	CompressWithContent(ctx context.Context, files map[string]io.Reader) ([]byte, error)
}

// DecompressedFile represents a single file from the zip archive
type DecompressedFile struct {
	Filename string
	Content  io.Reader
	Error    error
}
