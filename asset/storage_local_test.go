package asset

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	baseURL := "http://localhost:8080/assets"
	storage, err := NewLocalStorage(tempDir, baseURL)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Save and Get", func(t *testing.T) {
		key := "test/file.txt"
		content := "Hello, World!"
		contentType := "text/plain"
		contentEncoding := ""
		size := int64(len(content))

		err := storage.Save(ctx, key, strings.NewReader(content), size, contentType, contentEncoding)
		require.NoError(t, err)

		fullPath := filepath.Join(tempDir, key)
		_, err = os.Stat(fullPath)
		require.NoError(t, err)

		reader, err := storage.Get(ctx, key)
		require.NoError(t, err)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("Delete", func(t *testing.T) {
		key := "test/delete.txt"
		content := "Delete me"
		contentType := "text/plain"
		contentEncoding := ""
		size := int64(len(content))

		err := storage.Save(ctx, key, strings.NewReader(content), size, contentType, contentEncoding)
		require.NoError(t, err)

		fullPath := filepath.Join(tempDir, key)
		_, err = os.Stat(fullPath)
		require.NoError(t, err)

		err = storage.Delete(ctx, key)
		require.NoError(t, err)

		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("GenerateURL", func(t *testing.T) {
		key := "test/url.txt"
		expires := 1 * time.Hour

		url, err := storage.GenerateURL(ctx, key, expires)
		require.NoError(t, err)

		expected := baseURL + "/" + "test%2Furl.txt"
		assert.Equal(t, expected, url)
	})

	t.Run("GenerateUploadURL", func(t *testing.T) {
		key := "test/upload.txt"
		size := int64(1024)
		contentType := "text/plain"
		contentEncoding := ""
		expires := 1 * time.Hour

		url, err := storage.GenerateUploadURL(ctx, key, size, contentType, contentEncoding, expires)
		require.NoError(t, err)

		expected := baseURL + "/upload?key=test%2Fupload.txt&size=1024&contentType=text%2Fplain"
		assert.Equal(t, expected, url)
	})

	t.Run("Get Nonexistent File", func(t *testing.T) {
		key := "nonexistent/file.txt"
		_, err := storage.Get(ctx, key)
		assert.Error(t, err)
	})

	t.Run("Delete Nonexistent File", func(t *testing.T) {
		key := "nonexistent/file.txt"
		err := storage.Delete(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("Save with Invalid Path", func(t *testing.T) {
		key := filepath.Join("..", "invalid", "path.txt")
		content := "Invalid path"
		contentType := "text/plain"
		contentEncoding := ""
		size := int64(len(content))

		err := storage.Save(ctx, key, strings.NewReader(content), size, contentType, contentEncoding)
		if err != nil {
			assert.Error(t, err)
		} else {
			invalidPath := filepath.Join(tempDir, "..", "invalid", "path.txt")
			_, err = os.Stat(invalidPath)
			if err != nil {
				assert.True(t, os.IsNotExist(err))
			} else {
				assert.Fail(t, "File should not have been created outside base directory")
				os.Remove(invalidPath)
			}
		}
	})
}
