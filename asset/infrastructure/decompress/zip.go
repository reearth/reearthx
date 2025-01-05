// Package decompress provides functionality for decompressing various file formats.
package decompress

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/reearth/reearthx/asset/repository"
)

// ZipDecompressor handles decompression of zip files.
type ZipDecompressor struct{}

var _ repository.Decompressor = (*ZipDecompressor)(nil)

// NewZipDecompressor creates a new zip decompressor
func NewZipDecompressor() repository.Decompressor {
	return &ZipDecompressor{}
}

// DecompressWithContent decompresses zip content directly.
// It processes each file asynchronously and returns a channel of decompressed files.
func (d *ZipDecompressor) DecompressWithContent(ctx context.Context, content []byte) (<-chan repository.DecompressedFile, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Create a buffered channel to hold the decompressed files
	resultChan := make(chan repository.DecompressedFile, len(zipReader.File))
	var wg sync.WaitGroup

	// Start a goroutine to close the channel when all files are processed
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process each file in the zip archive
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() || isHiddenFile(f.Name) {
			continue
		}

		wg.Add(1)
		go func(f *zip.File) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				resultChan <- repository.DecompressedFile{
					Filename: f.Name,
					Error:    ctx.Err(),
				}
				return
			default:
				content, err := d.processFile(f)
				if err != nil {
					resultChan <- repository.DecompressedFile{
						Filename: f.Name,
						Error:    err,
					}
					return
				}

				resultChan <- repository.DecompressedFile{
					Filename: f.Name,
					Content:  content,
				}
			}
		}(f)
	}

	return resultChan, nil
}

// processFile handles a single file from the zip archive
func (d *ZipDecompressor) processFile(f *zip.File) (io.Reader, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer rc.Close()

	// Read the entire file content into memory
	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return bytes.NewReader(content), nil
}

// CompressWithContent compresses the provided content into a zip archive.
// It takes a map of filenames to their content readers and returns the compressed bytes.
func (d *ZipDecompressor) CompressWithContent(ctx context.Context, files map[string]io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	// Process each file in parallel
	for filename, content := range files {
		wg.Add(1)
		go func(filename string, content io.Reader) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				if err := d.addFileToZip(zipWriter, filename, content); err != nil {
					errChan <- err
				}
			}
		}(filename, content)
	}

	// Wait for all files to be processed
	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		if err != nil {
			return nil, fmt.Errorf("compression error: %w", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// addFileToZip adds a single file to the zip archive
func (d *ZipDecompressor) addFileToZip(zipWriter *zip.Writer, filename string, content io.Reader) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file in zip: %w", err)
	}

	if _, err := io.Copy(writer, content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}

// isHiddenFile checks if a file is hidden (starts with a dot).
func isHiddenFile(name string) bool {
	base := filepath.Base(name)
	return len(base) > 0 && base[0] == '.'
}
