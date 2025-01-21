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
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/repository"
	"google.golang.org/api/iterator"
)

var (
	ErrFailedToCreateClient = errors.New("failed to create client")
	ErrAssetAlreadyExists   = errors.New("asset already exists")
	ErrAssetNotFound        = errors.New("asset not found")
	ErrFailedToUpdateAsset  = errors.New("failed to update asset")
	ErrFailedToDeleteAsset  = errors.New("failed to delete asset")
	ErrFailedToListAssets   = errors.New("failed to list assets")
	ErrFailedToUploadFile   = errors.New("failed to upload file")
	ErrFailedToCloseWriter  = errors.New("failed to close writer")
	ErrFailedToReadFile     = errors.New("failed to read file")
	ErrFailedToGetAsset     = errors.New("failed to get asset")
	ErrFailedToGenerateURL  = errors.New("failed to generate upload URL")
	ErrFailedToMoveAsset    = errors.New("failed to move asset")
	ErrInvalidURL           = errors.New("invalid URL format")
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
		return nil, fmt.Errorf("%w: %v", ErrFailedToCreateClient, err)
	}

	var u *url.URL
	if baseURL != "" {
		u, err = url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
		}
	}

	return &Client{
		bucket:     client.Bucket(bucketName),
		bucketName: bucketName,
		basePath:   basePath,
		baseURL:    u,
	}, nil
}

func (c *Client) Create(ctx context.Context, asset *entity.Asset) error {
	obj := c.getObject(asset.ID())
	attrs := storage.ObjectAttrs{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Attrs(ctx); err == nil {
		return fmt.Errorf("%w: %s", ErrAssetAlreadyExists, asset.ID())
	}

	writer := obj.NewWriter(ctx)
	writer.ObjectAttrs = attrs
	return writer.Close()
}

func (c *Client) Read(ctx context.Context, id id.ID) (*entity.Asset, error) {
	attrs, err := c.getObject(id).Attrs(ctx)
	if err != nil {
		return nil, c.handleNotFound(err, id)
	}

	asset := entity.NewAsset(
		id,
		attrs.Metadata["name"],
		attrs.Size,
		attrs.ContentType,
	)

	return asset, nil
}

func (c *Client) Update(ctx context.Context, asset *entity.Asset) error {
	obj := c.getObject(asset.ID())
	update := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"name":         asset.Name(),
			"content_type": asset.ContentType(),
		},
	}

	if _, err := obj.Update(ctx, update); err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToUpdateAsset, err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, id id.ID) error {
	obj := c.getObject(id)
	if err := obj.Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf("%w: %v", ErrFailedToDeleteAsset, err)
	}
	return nil
}

func (c *Client) List(ctx context.Context) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	it := c.bucket.Objects(ctx, &storage.Query{Prefix: c.basePath})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedToListAssets, err)
		}

		id, err := id.IDFrom(path.Base(attrs.Name))
		if err != nil {
			continue // skip invalid IDs
		}

		asset := entity.NewAsset(
			id,
			attrs.Metadata["name"],
			attrs.Size,
			attrs.ContentType,
		)
		assets = append(assets, asset)
	}

	return assets, nil
}

func (c *Client) Upload(ctx context.Context, id id.ID, content io.Reader) error {
	obj := c.getObject(id)
	writer := obj.NewWriter(ctx)

	if _, err := io.Copy(writer, content); err != nil {
		_ = writer.Close()
		return fmt.Errorf("%w: %v", ErrFailedToUploadFile, err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToCloseWriter, err)
	}
	return nil
}

func (c *Client) Download(ctx context.Context, id id.ID) (io.ReadCloser, error) {
	obj := c.getObject(id)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", ErrFailedToReadFile, err)
	}
	return reader, nil
}

func (c *Client) GetUploadURL(ctx context.Context, id id.ID) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:  "PUT",
		Expires: time.Now().Add(15 * time.Minute),
	}

	signedURL, err := c.bucket.SignedURL(c.objectPath(id), opts)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFailedToGenerateURL, err)
	}
	return signedURL, nil
}

func (c *Client) Move(ctx context.Context, fromID, toID id.ID) error {
	src := c.getObject(fromID)
	dst := c.getObject(toID)

	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToMoveAsset, err)
	}

	if err := src.Delete(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToMoveAsset, err)
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
			return fmt.Errorf("%w: %v", ErrFailedToDeleteAsset, err)
		}

		if err := c.bucket.Object(attrs.Name).Delete(ctx); err != nil {
			if !errors.Is(err, storage.ErrObjectNotExist) {
				return fmt.Errorf("%w: %v", ErrFailedToDeleteAsset, err)
			}
		}
	}
	return nil
}

func (c *Client) GetObjectURL(id id.ID) string {
	if c.baseURL == nil {
		return ""
	}
	u := *c.baseURL
	u.Path = path.Join(u.Path, c.objectPath(id))
	return u.String()
}

func (c *Client) GetIDFromURL(urlStr string) (id.ID, error) {
	emptyID := id.NewID()

	if c.baseURL == nil {
		return emptyID, fmt.Errorf("%w: base URL not set", ErrInvalidURL)
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return emptyID, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	if u.Host != c.baseURL.Host {
		return emptyID, fmt.Errorf("%w: host mismatch", ErrInvalidURL)
	}

	urlPath := strings.TrimPrefix(u.Path, c.baseURL.Path)
	urlPath = strings.TrimPrefix(urlPath, "/")
	urlPath = strings.TrimPrefix(urlPath, c.basePath)
	urlPath = strings.TrimPrefix(urlPath, "/")

	return id.IDFrom(urlPath)
}

func (c *Client) getObject(id id.ID) *storage.ObjectHandle {
	return c.bucket.Object(c.objectPath(id))
}

func (c *Client) objectPath(id id.ID) string {
	return path.Join(c.basePath, id.String())
}

func (c *Client) handleNotFound(err error, id id.ID) error {
	if errors.Is(err, storage.ErrObjectNotExist) {
		return fmt.Errorf("%w: %s", ErrAssetNotFound, id)
	}
	return fmt.Errorf("%w: %v", ErrFailedToGetAsset, err)
}

func (c *Client) FindByGroup(ctx context.Context, groupID id.GroupID) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	it := c.bucket.Objects(ctx, &storage.Query{Prefix: c.basePath})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedToListAssets, err)
		}

		assetID, err := id.IDFrom(path.Base(attrs.Name))
		if err != nil {
			continue // skip invalid IDs
		}

		asset := entity.NewAsset(
			assetID,
			attrs.Metadata["name"],
			attrs.Size,
			attrs.ContentType,
		)

		if asset.GroupID() == groupID {
			assets = append(assets, asset)
		}
	}

	return assets, nil
}
