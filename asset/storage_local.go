package asset

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type localStorage struct {
	baseDir string
	baseURL string
}

func NewLocalStorage(baseDir, baseURL string) (Storage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &localStorage{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

func (s *localStorage) Save(ctx context.Context, key string, data io.Reader, size int64, contentType string, contentEncoding string) error {
	fullPath := filepath.Join(s.baseDir, key)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *localStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, key)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (s *localStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.baseDir, key)

	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *localStorage) GenerateURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	encodedKey := url.PathEscape(key)

	return fmt.Sprintf("%s/%s", s.baseURL, encodedKey), nil
}

func (s *localStorage) GenerateUploadURL(ctx context.Context, key string, size int64, contentType string, contentEncoding string, expires time.Duration) (string, error) {
	encodedKey := url.PathEscape(key)

	uploadURL := fmt.Sprintf("%s/upload?key=%s&size=%d&contentType=%s",
		s.baseURL,
		encodedKey,
		size,
		url.QueryEscape(contentType),
	)

	if contentEncoding != "" {
		uploadURL += "&contentEncoding=" + url.QueryEscape(contentEncoding)
	}

	return uploadURL, nil
}
