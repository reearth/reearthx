package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/reearth/reearthx/asset/domain"
	"github.com/reearth/reearthx/asset/repository"
	"google.golang.org/api/iterator"
)

const (
	errFailedToCreateClient = "failed to create client: %w"
	errAssetAlreadyExists   = "asset already exists: %s"
	errAssetNotFound        = "asset not found: %s"
	errFailedToUpdateAsset  = "failed to update asset: %w"
	errFailedToDeleteAsset  = "failed to delete asset: %w"
	errFailedToListAssets   = "failed to list assets: %w"
	errFailedToUploadFile   = "failed to upload file: %w"
	errFailedToCloseWriter  = "failed to close writer: %w"
	errFailedToReadFile     = "failed to read file: %w"
	errFailedToGetAsset     = "failed to get asset: %w"
	errFailedToGenerateURL  = "failed to generate upload URL: %w"
	errFailedToMoveAsset    = "failed to move asset: %w"
	errInvalidURL           = "invalid URL format: %s"
)

type Client struct {
	bucket     *storage.BucketHandle
	bucketName string
	basePath   string
	baseURL    *url.URL
}

var _ repository.PersistenceRepository = (*Client)(nil)

func NewClient(ctx context.Context, bucketName string, basePath string, baseURL string) (*Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf(errFailedToCreateClient, err)
	}

	var u *url.URL
	if baseURL != "" {
		u, err = url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf(errInvalidURL, err)
		}
	}

	return &Client{
		bucket:     client.Bucket(bucketName),
		bucketName: bucketName,
		basePath:   basePath,
		baseURL:    u,
	}, nil
}

func (c *Client) Create(ctx context.Context, asset *domain.Asset) error {
	obj := c.getObject(asset.ID())
	attrs := storage.ObjectAttrs{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Attrs(ctx); err == nil {
		return fmt.Errorf(errAssetAlreadyExists, asset.ID())
	}

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs = attrs
	return writer.Close()
}

func (c *Client) Read(ctx context.Context, id domain.ID) (*domain.Asset, error) {
	attrs, err := c.getObject(id).Attrs(ctx)
	if err != nil {
		return nil, c.handleNotFound(err, id)
	}

	asset := domain.NewAsset(
		id,
		attrs.Metadata["name"],
		attrs.Size,
		attrs.ContentType,
	)

	return asset, nil
}

func (c *Client) Update(ctx context.Context, asset *domain.Asset) error {
	obj := c.getObject(asset.ID())
	update := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Update(ctx, update); err != nil {
		return fmt.Errorf(errFailedToUpdateAsset, err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, id domain.ID) error {
	obj := c.getObject(id)
	if err := obj.Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf(errFailedToDeleteAsset, err)
	}
	return nil
}

func (c *Client) List(ctx context.Context) ([]*domain.Asset, error) {
	var assets []*domain.Asset
	it := c.bucket.Objects(ctx, &storage.Query{Prefix: c.basePath})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf(errFailedToListAssets, err)
		}

		id, err := domain.IDFrom(path.Base(attrs.Name))
		if err != nil {
			continue // skip invalid IDs
		}

		asset := domain.NewAsset(
			id,
			attrs.Metadata["name"],
			attrs.Size,
			attrs.ContentType,
		)
		assets = append(assets, asset)
	}

	return assets, nil
}

func (c *Client) Upload(ctx context.Context, id domain.ID, content io.Reader) error {
	obj := c.getObject(id)
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, content); err != nil {
		_ = writer.Close()
		return fmt.Errorf(errFailedToUploadFile, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf(errFailedToCloseWriter, err)
	}
	return nil
}

func (c *Client) Download(ctx context.Context, id domain.ID) (io.ReadCloser, error) {
	obj := c.getObject(id)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, fmt.Errorf(errAssetNotFound, id)
		}
		return nil, fmt.Errorf(errFailedToReadFile, err)
	}
	return reader, nil
}

func (c *Client) GetUploadURL(ctx context.Context, id domain.ID) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(15 * time.Minute),
	}

	signedURL, err := c.bucket.SignedURL(c.objectPath(id), opts)
	if err != nil {
		return "", fmt.Errorf(errFailedToGenerateURL, err)
	}
	return signedURL, nil
}

func (c *Client) Move(ctx context.Context, fromID, toID domain.ID) error {
	src := c.getObject(fromID)
	dst := c.getObject(toID)

	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf(errFailedToMoveAsset, err)
	}

	if err := src.Delete(ctx); err != nil {
		return fmt.Errorf(errFailedToMoveAsset, err)
	}

	return nil
}

func (c *Client) DeleteAll(ctx context.Context, prefix string) error {
	it := c.bucket.Objects(ctx, &storage.Query{
		Prefix: path.Join(c.basePath, prefix),
	})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf(errFailedToDeleteAsset, err)
		}

		if err := c.bucket.Object(attrs.Name).Delete(ctx); err != nil {
			if !errors.Is(err, storage.ErrObjectNotExist) {
				return fmt.Errorf(errFailedToDeleteAsset, err)
			}
		}
	}
	return nil
}

func (c *Client) GetObjectURL(id domain.ID) string {
	if c.baseURL == nil {
		return ""
	}
	u := *c.baseURL
	u.Path = path.Join(u.Path, c.objectPath(id))
	return u.String()
}

func (c *Client) GetIDFromURL(urlStr string) (domain.ID, error) {
	emptyID := domain.NewID()

	if c.baseURL == nil {
		return emptyID, fmt.Errorf(errInvalidURL, "base URL not set")
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return emptyID, fmt.Errorf(errInvalidURL, err)
	}

	if u.Host != c.baseURL.Host || u.Scheme != c.baseURL.Scheme {
		return emptyID, fmt.Errorf(errInvalidURL, "host or scheme mismatch")
	}

	p := strings.TrimPrefix(u.Path, "/")
	p = strings.TrimPrefix(p, c.basePath)
	p = strings.TrimPrefix(p, "/")

	if p == "" {
		return emptyID, fmt.Errorf(errInvalidURL, "empty path")
	}

	id, err := domain.IDFrom(p)
	if err != nil {
		return emptyID, fmt.Errorf(errInvalidURL, err)
	}

	return id, nil
}

func (c *Client) getObject(id domain.ID) *storage.ObjectHandle {
	return c.bucket.Object(c.objectPath(id))
}

func (c *Client) objectPath(id domain.ID) string {
	return path.Join(c.basePath, id.String())
}

func (c *Client) handleNotFound(err error, id domain.ID) error {
	if errors.Is(err, storage.ErrObjectNotExist) {
		return fmt.Errorf(errAssetNotFound, id)
	}
	return fmt.Errorf(errFailedToGetAsset, err)
}

func (c *Client) FindByGroup(ctx context.Context, groupID domain.GroupID) ([]*domain.Asset, error) {
	var assets []*domain.Asset
	it := c.bucket.Objects(ctx, &storage.Query{
		Prefix: path.Join(c.basePath, groupID.String()),
	})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf(errFailedToListAssets, err)
		}

		id, err := domain.IDFrom(path.Base(attrs.Name))
		if err != nil {
			continue // skip invalid IDs
		}

		asset := domain.NewAsset(
			id,
			attrs.Metadata["name"],
			attrs.Size,
			attrs.ContentType,
		)
		assets = append(assets, asset)
	}

	return assets, nil
}
