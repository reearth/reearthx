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

	"github.com/reearth/reearthx/log"
	"go.uber.org/zap"

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

	// Process files in a separate goroutine to avoid race conditions
	go func() {
		for _, f := range zipReader.File {
			if f.FileInfo().IsDir() || isHiddenFile(f.Name) {
				continue
			}

			wg.Add(1)
			go d.processZipFile(ctx, f, resultChan, &wg)
		}

		// Wait for all files to be processed before closing the channel
		wg.Wait()
		close(resultChan)
	}()

	return resultChan, nil
}

// processZipFile handles processing of a single zip file entry
func (d *ZipDecompressor) processZipFile(ctx context.Context, f *zip.File, resultChan chan<- repository.DecompressedFile, wg *sync.WaitGroup) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		resultChan <- repository.DecompressedFile{
			Name:  f.Name,
			Error: ctx.Err(),
		}
		return
	default:
		content, err := d.processFile(f)
		if err != nil {
			resultChan <- repository.DecompressedFile{
				Name:  f.Name,
				Error: err,
			}
			return
		}

		resultChan <- repository.DecompressedFile{
			Name:    f.Name,
			Content: content,
		}
	}
}

// processFile handles a single file from the zip archive
func (d *ZipDecompressor) processFile(f *zip.File) (io.Reader, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer func(rc io.ReadCloser) {
		err := rc.Close()
		if err != nil {
			log.Warn("failed to close file in zip", zap.Error(err))
		}
	}(rc)

	// Read the entire file content into memory
	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return bytes.NewReader(content), nil
}

// CompressWithContent compresses the provided content into a zip archive.
// It takes a map of filenames to their content readers and returns a channel that will receive the compressed bytes.
// The channel will be closed when compression is complete or if an error occurs.
func (d *ZipDecompressor) CompressWithContent(ctx context.Context, files map[string]io.Reader) (<-chan repository.CompressResult, error) {
	resultChan := make(chan repository.CompressResult, 1)

	go func() {
		defer close(resultChan)

		select {
		case <-ctx.Done():
			resultChan <- repository.CompressResult{Error: ctx.Err()}
			return
		default:
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)
		defer func(zipWriter *zip.Writer) {
			err := zipWriter.Close()
			if err != nil {
				log.Warn("failed to close zip writer", zap.Error(err))
			}
		}(zipWriter)

		// Use a mutex to protect concurrent writes to the zip writer
		var mu sync.Mutex
		var wg sync.WaitGroup
		errChan := make(chan error, 1)

		// Process each file sequentially to avoid corruption
		for filename, content := range files {
			select {
			case <-ctx.Done():
				resultChan <- repository.CompressResult{Error: ctx.Err()}
				return
			default:
			}

			wg.Add(1)
			go func(filename string, content io.Reader) {
				defer wg.Done()

				// Read the entire content first to avoid holding the lock during I/O
				data, err := io.ReadAll(content)
				if err != nil {
					select {
					case errChan <- fmt.Errorf("failed to read content: %w", err):
					default:
					}
					return
				}

				mu.Lock()
				err = d.addFileToZip(zipWriter, filename, bytes.NewReader(data))
				mu.Unlock()

				if err != nil {
					select {
					case errChan <- err:
					default:
					}
				}
			}(filename, content)
		}

		// Wait for all goroutines to finish
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		// Wait for either completion or error
		select {
		case <-ctx.Done():
			resultChan <- repository.CompressResult{Error: ctx.Err()}
		case err := <-errChan:
			resultChan <- repository.CompressResult{Error: err}
		case <-done:
			if err := zipWriter.Close(); err != nil {
				resultChan <- repository.CompressResult{Error: fmt.Errorf("failed to close zip writer: %w", err)}
				return
			}
			resultChan <- repository.CompressResult{Content: buf.Bytes()}
		}
	}()

	return resultChan, nil
}

// addFileToZip adds a single file to the zip archive
// Note: This function is not thread-safe and should be protected by a mutex
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
