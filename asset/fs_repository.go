package asset

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FSRepository struct {
	baseDir string
}

func NewFSRepository(baseDir string) (*FSRepository, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FSRepository{
		baseDir: baseDir,
	}, nil
}

func (r *FSRepository) Fetch(ctx context.Context, id ID) (*Asset, error) {
	path := r.getPath(id)

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("asset not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get asset info: %w", err)
	}

	return &Asset{
		ID:        id,
		Name:      filepath.Base(path),
		Size:      info.Size(),
		CreatedAt: info.ModTime(),
		UpdatedAt: info.ModTime(),
	}, nil
}

func (r *FSRepository) FetchFile(ctx context.Context, id ID) (io.ReadCloser, error) {
	path := r.getPath(id)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("asset file not found: %s", id)
		}
		return nil, fmt.Errorf("failed to open asset file: %w", err)
	}

	return file, nil
}

func (r *FSRepository) Save(ctx context.Context, asset *Asset) error {
	// Only update metadata in this case
	// Actual file content is handled by Upload method
	return nil
}

func (r *FSRepository) Remove(ctx context.Context, id ID) error {
	path := r.getPath(id)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("asset not found: %s", id)
		}
		return fmt.Errorf("failed to remove asset: %w", err)
	}

	return nil
}

func (r *FSRepository) Upload(ctx context.Context, id ID, file io.Reader) error {
	path := r.getPath(id)

	// Create destination file
	dst, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

func (r *FSRepository) GetUploadURL(ctx context.Context, id ID) (string, error) {
	// For file system implementation, we don't support direct upload URLs
	// In a real implementation (e.g., S3), this would return a pre-signed URL
	return "", fmt.Errorf("direct upload URLs not supported for file system repository")
}

func (r *FSRepository) getPath(id ID) string {
	return filepath.Join(r.baseDir, id.String())
}

// Create creates a new asset
func (r *FSRepository) Create(ctx context.Context, asset *Asset) error {
	if asset == nil {
		return fmt.Errorf("asset is nil")
	}

	// Save metadata
	return r.Save(ctx, asset)
}

// Read returns an asset by ID
func (r *FSRepository) Read(ctx context.Context, id ID) (*Asset, error) {
	return r.Fetch(ctx, id)
}

// Update updates an existing asset
func (r *FSRepository) Update(ctx context.Context, asset *Asset) error {
	if asset == nil {
		return fmt.Errorf("asset is nil")
	}

	// Check if asset exists
	_, err := r.Fetch(ctx, asset.ID)
	if err != nil {
		return err
	}

	return r.Save(ctx, asset)
}

// Delete removes an asset by ID
func (r *FSRepository) Delete(ctx context.Context, id ID) error {
	return r.Remove(ctx, id)
}

// List returns all assets
func (r *FSRepository) List(ctx context.Context) ([]*Asset, error) {
	files, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var assets []*Asset
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		asset := &Asset{
			ID:        ID(file.Name()),
			Name:      info.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			UpdatedAt: info.ModTime(),
		}
		assets = append(assets, asset)
	}

	return assets, nil
}
