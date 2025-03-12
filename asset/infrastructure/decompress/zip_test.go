package decompress

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZipDecompressor_DecompressWithContent(t *testing.T) {
	// Create test data
	files := map[string]string{
		"test1.txt": "Hello, World!",
		"test2.txt": "This is a test file",
		".hidden":   "This should be skipped",
	}

	// Create a zip file
	zipContent, err := createTestZip(files)
	assert.NoError(t, err)

	// Create decompressor
	d := NewZipDecompressor()

	// Test decompression
	ctx := context.Background()
	resultChan, err := d.DecompressWithContent(ctx, zipContent)
	assert.NoError(t, err)

	// Collect results
	results := make(map[string]string)
	for file := range resultChan {
		assert.NoError(t, file.Error)
		if file.Error != nil {
			continue
		}

		content, err := io.ReadAll(file.Content)
		assert.NoError(t, err)
		results[file.Name] = string(content)
	}

	// Verify results
	assert.Equal(t, 2, len(results)) // .hidden should be skipped
	assert.Equal(t, "Hello, World!", results["test1.txt"])
	assert.Equal(t, "This is a test file", results["test2.txt"])
}

func TestZipDecompressor_CompressWithContent(t *testing.T) {
	// Create test data with small files to avoid memory issues
	files := map[string]io.Reader{
		"test1.txt": strings.NewReader("Hello, World!"),
		"test2.txt": strings.NewReader("This is a test file"),
	}

	// Create decompressor
	d := NewZipDecompressor()

	// Test compression
	ctx := context.Background()
	compressChan, err := d.CompressWithContent(ctx, files)
	assert.NoError(t, err)

	// Get compression result
	result := <-compressChan
	assert.NoError(t, result.Error)
	compressed := result.Content

	// Test decompression of the compressed content
	resultChan, err := d.DecompressWithContent(ctx, compressed)
	assert.NoError(t, err)

	// Collect and verify results
	results := make(map[string]string)
	for file := range resultChan {
		assert.NoError(t, file.Error)
		if file.Error != nil {
			continue
		}

		content, err := io.ReadAll(file.Content)
		assert.NoError(t, err)
		results[file.Name] = string(content)
	}

	assert.Equal(t, 2, len(results))
	assert.Equal(t, "Hello, World!", results["test1.txt"])
	assert.Equal(t, "This is a test file", results["test2.txt"])
}

func TestZipDecompressor_ContextCancellation(t *testing.T) {
	// Create test data
	files := map[string]string{
		"test1.txt": "Hello, World!",
		"test2.txt": "This is a test file",
	}

	// Create a zip file
	zipContent, err := createTestZip(files)
	assert.NoError(t, err)

	// Create decompressor
	d := NewZipDecompressor()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test decompression with cancelled context
	resultChan, err := d.DecompressWithContent(ctx, zipContent)
	assert.NoError(t, err)

	// Verify that all files return context cancelled error
	for file := range resultChan {
		assert.Error(t, file.Error)
		assert.Equal(t, context.Canceled, file.Error)
	}
}

func TestZipDecompressor_InvalidZip(t *testing.T) {
	d := NewZipDecompressor()
	ctx := context.Background()

	// Test with invalid zip content
	_, err := d.DecompressWithContent(ctx, []byte("invalid zip content"))
	assert.Error(t, err)
}

// Helper function to create a test zip file
func createTestZip(files map[string]string) ([]byte, error) {
	d := NewZipDecompressor()
	ctx := context.Background()

	// Convert string content to io.Reader
	readers := make(map[string]io.Reader)
	for name, content := range files {
		readers[name] = strings.NewReader(content)
	}

	compressChan, err := d.CompressWithContent(ctx, readers)
	if err != nil {
		return nil, err
	}

	result := <-compressChan
	return result.Content, result.Error
}
