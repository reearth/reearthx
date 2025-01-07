package repository

import (
	"context"
	"io"
)

// CompressResult represents the result of a compression operation
type CompressResult struct {
	Content []byte
	Error   error
}

// Decompressor defines the interface for compression and decompression operations
type Decompressor interface {
	// DecompressWithContent decompresses zip content directly and returns a channel of decompressed files
	// The channel will be closed when all files have been processed or an error occurs
	DecompressWithContent(ctx context.Context, content []byte) (<-chan DecompressedFile, error)

	// CompressWithContent compresses the provided content into a zip archive
	// Returns a channel that will receive the compressed bytes or an error
	// The channel will be closed when compression is complete or if an error occurs
	CompressWithContent(ctx context.Context, files map[string]io.Reader) (<-chan CompressResult, error)
}

// DecompressedFile represents a single file from the zip archive
type DecompressedFile struct {
	Filename string
	Content  io.Reader
	Error    error
}
