package gcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/reearth/reearthx/asset/domain/entity"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/stretchr/testify/assert"
)

type mockBucketHandle struct {
	objects map[string]*mockObject
}

type mockObject struct {
	data   []byte
	attrs  *storage.ObjectAttrs
	bucket *mockBucketHandle
	name   string
}

func (o *mockObject) Delete(context.Context) error {
	delete(o.bucket.objects, o.name)
	return nil
}

func (o *mockObject) Attrs(context.Context) (*storage.ObjectAttrs, error) {
	if o.attrs == nil {
		return nil, storage.ErrObjectNotExist
	}
	return o.attrs, nil
}

func (o *mockObject) NewReader(context.Context) (io.ReadCloser, error) {
	if o.data == nil {
		return nil, storage.ErrObjectNotExist
	}
	return &mockReader{bytes.NewReader(o.data)}, nil
}

func (o *mockObject) NewWriter(context.Context) io.WriteCloser {
	return &mockWriter{
		buf:        bytes.NewBuffer(nil),
		bucket:     o.bucket,
		objectName: o.name,
		attrs:      o.attrs,
	}
}

func (o *mockObject) Update(ctx context.Context, uattrs storage.ObjectAttrsToUpdate) (*storage.ObjectAttrs, error) {
	if o.attrs == nil {
		return nil, storage.ErrObjectNotExist
	}
	if uattrs.Metadata != nil {
		o.attrs.Metadata = uattrs.Metadata
	}
	return o.attrs, nil
}

func (o *mockObject) CopierFrom(src *storage.ObjectHandle) *storage.Copier {
	return &storage.Copier{}
}

type mockReader struct {
	*bytes.Reader
}

func (r *mockReader) Close() error {
	return nil
}

type mockWriter struct {
	buf        *bytes.Buffer
	attrs      *storage.ObjectAttrs
	bucket     *mockBucketHandle
	objectName string
}

func (w *mockWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *mockWriter) Close() error {
	obj := w.bucket.objects[w.objectName]
	obj.data = w.buf.Bytes()
	obj.attrs = w.attrs
	return nil
}

func newMockBucketHandle() *mockBucketHandle {
	return &mockBucketHandle{
		objects: make(map[string]*mockObject),
	}
}

type testClient struct {
	*Client
	mockBucket *mockBucketHandle
}

func newTestClient(_ *testing.T) *testClient {
	mockBucket := newMockBucketHandle()
	client := &Client{
		bucketName: "test-bucket",
		basePath:   "test-path",
		baseURL: &url.URL{
			Scheme: "https",
			Host:   "storage.googleapis.com",
		},
	}
	return &testClient{
		Client:     client,
		mockBucket: mockBucket,
	}
}

func (c *testClient) Create(ctx context.Context, asset *entity.Asset) error {
	objPath := c.objectPath(asset.ID())
	if _, exists := c.mockBucket.objects[objPath]; exists {
		return fmt.Errorf(errAssetAlreadyExists, asset.ID())
	}

	c.mockBucket.objects[objPath] = &mockObject{
		bucket: c.mockBucket,
		name:   objPath,
		attrs: &storage.ObjectAttrs{
			Name: objPath,
			Metadata: map[string]string{
				"name":         asset.Name(),
				"content_type": asset.ContentType(),
			},
		},
	}
	return nil
}

func (c *testClient) Read(ctx context.Context, id id.ID) (*entity.Asset, error) {
	objPath := c.objectPath(id)
	obj, exists := c.mockBucket.objects[objPath]
	if !exists {
		return nil, fmt.Errorf(errAssetNotFound, id)
	}

	return entity.NewAsset(
		id,
		obj.attrs.Metadata["name"],
		int64(len(obj.data)),
		obj.attrs.Metadata["content_type"],
	), nil
}

func (c *testClient) Update(ctx context.Context, asset *entity.Asset) error {
	objPath := c.objectPath(asset.ID())
	obj, exists := c.mockBucket.objects[objPath]
	if !exists {
		return fmt.Errorf(errAssetNotFound, asset.ID())
	}

	obj.attrs.Metadata["name"] = asset.Name()
	obj.attrs.Metadata["content_type"] = asset.ContentType()
	return nil
}

func (c *testClient) Delete(ctx context.Context, id id.ID) error {
	objPath := c.objectPath(id)
	delete(c.mockBucket.objects, objPath)
	return nil
}

func (c *testClient) Upload(ctx context.Context, id id.ID, content io.Reader) error {
	objPath := c.objectPath(id)
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	obj, exists := c.mockBucket.objects[objPath]
	if !exists {
		obj = &mockObject{
			bucket: c.mockBucket,
			name:   objPath,
			attrs: &storage.ObjectAttrs{
				Name:     objPath,
				Metadata: make(map[string]string),
			},
		}
		c.mockBucket.objects[objPath] = obj
	}

	obj.data = data
	return nil
}

func (c *testClient) Download(ctx context.Context, id id.ID) (io.ReadCloser, error) {
	objPath := c.objectPath(id)
	obj, exists := c.mockBucket.objects[objPath]
	if !exists {
		return nil, fmt.Errorf(errAssetNotFound, id)
	}

	return &mockReader{bytes.NewReader(obj.data)}, nil
}

func (c *testClient) GetUploadURL(ctx context.Context, id id.ID) (string, error) {
	return fmt.Sprintf("https://storage.googleapis.com/%s", c.objectPath(id)), nil
}

func (c *testClient) Move(ctx context.Context, fromID, toID id.ID) error {
	fromPath := c.objectPath(fromID)
	toPath := c.objectPath(toID)

	fromObj, exists := c.mockBucket.objects[fromPath]
	if !exists {
		return fmt.Errorf(errAssetNotFound, fromID)
	}

	if _, exists := c.mockBucket.objects[toPath]; exists {
		return fmt.Errorf("destination already exists")
	}

	c.mockBucket.objects[toPath] = &mockObject{
		bucket: c.mockBucket,
		name:   toPath,
		data:   fromObj.data,
		attrs: &storage.ObjectAttrs{
			Name:     toPath,
			Metadata: fromObj.attrs.Metadata,
		},
	}

	delete(c.mockBucket.objects, fromPath)
	return nil
}

func (c *testClient) List(ctx context.Context) ([]*entity.Asset, error) {
	var assets []*entity.Asset
	for _, obj := range c.mockBucket.objects {
		id, err := id.IDFrom(path.Base(obj.name))
		if err != nil {
			continue
		}

		asset := entity.NewAsset(
			id,
			obj.attrs.Metadata["name"],
			int64(len(obj.data)),
			obj.attrs.Metadata["content_type"],
		)
		assets = append(assets, asset)
	}
	return assets, nil
}

func (c *testClient) DeleteAll(ctx context.Context, prefix string) error {
	fullPrefix := path.Join(c.basePath, prefix)
	for name := range c.mockBucket.objects {
		if strings.HasPrefix(name, fullPrefix) {
			delete(c.mockBucket.objects, name)
		}
	}
	return nil
}

func TestClient_Create(t *testing.T) {
	client := newTestClient(t)

	asset := entity.NewAsset(
		id.NewID(),
		"test-asset",
		100,
		"application/json",
	)

	err := client.Create(context.Background(), asset)
	assert.NoError(t, err)

	obj := client.mockBucket.objects[client.objectPath(asset.ID())]
	assert.NotNil(t, obj)
	assert.Equal(t, asset.Name(), obj.attrs.Metadata["name"])
	assert.Equal(t, asset.ContentType(), obj.attrs.Metadata["content_type"])
}

func TestClient_Read(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	name := "test-asset"
	contentType := "application/json"
	objPath := client.objectPath(id)

	client.mockBucket.objects[objPath] = &mockObject{
		bucket: client.mockBucket,
		name:   objPath,
		attrs: &storage.ObjectAttrs{
			Name: objPath,
			Metadata: map[string]string{
				"name":         name,
				"content_type": contentType,
			},
		},
	}

	asset, err := client.Read(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, id, asset.ID())
	assert.Equal(t, name, asset.Name())
	assert.Equal(t, contentType, asset.ContentType())
}

func TestClient_Update(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	objPath := client.objectPath(id)

	client.mockBucket.objects[objPath] = &mockObject{
		bucket: client.mockBucket,
		name:   objPath,
		attrs: &storage.ObjectAttrs{
			Name: objPath,
			Metadata: map[string]string{
				"name":         "test-asset",
				"content_type": "application/json",
			},
		},
	}

	updatedAsset := entity.NewAsset(
		id,
		"updated-asset",
		100,
		"application/json",
	)

	err := client.Update(context.Background(), updatedAsset)
	assert.NoError(t, err)

	obj := client.mockBucket.objects[objPath]
	assert.Equal(t, updatedAsset.Name(), obj.attrs.Metadata["name"])
}

func TestClient_Delete(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	objPath := client.objectPath(id)

	client.mockBucket.objects[objPath] = &mockObject{
		bucket: client.mockBucket,
		name:   objPath,
		attrs: &storage.ObjectAttrs{
			Name: objPath,
			Metadata: map[string]string{
				"name":         "test-asset",
				"content_type": "application/json",
			},
		},
	}

	err := client.Delete(context.Background(), id)
	assert.NoError(t, err)

	_, exists := client.mockBucket.objects[objPath]
	assert.False(t, exists)
}

func TestClient_Upload(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	content := []byte("test content")
	objPath := client.objectPath(id)

	err := client.Upload(context.Background(), id, bytes.NewReader(content))
	assert.NoError(t, err)

	obj := client.mockBucket.objects[objPath]
	assert.Equal(t, content, obj.data)
}

func TestClient_Download(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	content := []byte("test content")
	objPath := client.objectPath(id)

	client.mockBucket.objects[objPath] = &mockObject{
		bucket: client.mockBucket,
		name:   objPath,
		data:   content,
		attrs: &storage.ObjectAttrs{
			Name:     objPath,
			Metadata: make(map[string]string),
		},
	}

	reader, err := client.Download(context.Background(), id)
	assert.NoError(t, err)

	downloaded, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, content, downloaded)
}

func TestClient_Create_AlreadyExists(t *testing.T) {
	client := newTestClient(t)

	asset := entity.NewAsset(
		id.NewID(),
		"test-asset",
		100,
		"application/json",
	)

	objPath := client.objectPath(asset.ID())
	client.mockBucket.objects[objPath] = &mockObject{
		bucket: client.mockBucket,
		name:   objPath,
		attrs: &storage.ObjectAttrs{
			Name:     objPath,
			Metadata: make(map[string]string),
		},
	}

	err := client.Create(context.Background(), asset)
	assert.Error(t, err)
}

func TestClient_Read_NotFound(t *testing.T) {
	client := newTestClient(t)

	_, err := client.Read(context.Background(), id.NewID())
	assert.Error(t, err)
}

func TestClient_Update_NotFound(t *testing.T) {
	client := newTestClient(t)

	asset := entity.NewAsset(
		id.NewID(),
		"test-asset",
		100,
		"application/json",
	)

	err := client.Update(context.Background(), asset)
	assert.Error(t, err)
}

func TestClient_Download_NotFound(t *testing.T) {
	client := newTestClient(t)

	_, err := client.Download(context.Background(), id.NewID())
	assert.Error(t, err)
}

func TestClient_GetObjectURL(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	url := client.GetObjectURL(id)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, client.objectPath(id))
}

func TestClient_GetIDFromURL(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	url := client.GetObjectURL(id)

	parsedID, err := client.GetIDFromURL(url)
	assert.NoError(t, err)
	assert.Equal(t, id, parsedID)
}

func TestClient_GetIDFromURL_InvalidURL(t *testing.T) {
	client := newTestClient(t)

	_, err := client.GetIDFromURL("invalid-url")
	assert.Error(t, err)
}

func TestClient_GetIDFromURL_MismatchedHost(t *testing.T) {
	client := newTestClient(t)

	_, err := client.GetIDFromURL("https://different-host.com/test-path/123")
	assert.Error(t, err)
}

func TestClient_GetIDFromURL_EmptyPath(t *testing.T) {
	client := newTestClient(t)

	_, err := client.GetIDFromURL("https://storage.googleapis.com")
	assert.Error(t, err)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		basePath   string
		baseURL    string
		wantErr    bool
	}{
		{
			name:       "valid configuration",
			bucketName: "test-bucket",
			basePath:   "test-path",
			baseURL:    "https://storage.googleapis.com",
			wantErr:    false,
		},
		{
			name:       "empty bucket name",
			bucketName: "",
			basePath:   "test-path",
			baseURL:    "https://storage.googleapis.com",
			wantErr:    true,
		},
		{
			name:       "invalid base URL",
			bucketName: "test-bucket",
			basePath:   "test-path",
			baseURL:    "://invalid-url",
			wantErr:    true,
		},
		{
			name:       "empty base URL",
			bucketName: "test-bucket",
			basePath:   "test-path",
			baseURL:    "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.bucketName == "" {
				assert.Error(t, fmt.Errorf("bucket name is required"))
				return
			}

			client := &Client{
				bucketName: tt.bucketName,
				basePath:   tt.basePath,
			}

			var err error
			if tt.baseURL != "" {
				client.baseURL, err = url.Parse(tt.baseURL)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.NotNil(t, client.baseURL)
				assert.Equal(t, tt.baseURL, client.baseURL.String())
			}

			assert.Equal(t, tt.bucketName, client.bucketName)
			assert.Equal(t, tt.basePath, client.basePath)
		})
	}
}

func TestClient_List(t *testing.T) {
	client := newTestClient(t)

	// Create multiple test objects
	objects := []struct {
		id          id.ID
		name        string
		contentType string
	}{
		{id.NewID(), "asset1", "application/json"},
		{id.NewID(), "asset2", "application/json"},
		{id.NewID(), "asset3", "application/json"},
	}

	for _, obj := range objects {
		objPath := client.objectPath(obj.id)
		client.mockBucket.objects[objPath] = &mockObject{
			bucket: client.mockBucket,
			name:   objPath,
			attrs: &storage.ObjectAttrs{
				Name: objPath,
				Metadata: map[string]string{
					"name":         obj.name,
					"content_type": obj.contentType,
				},
			},
		}
	}

	assets, err := client.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, assets, len(objects))
}

func TestClient_DeleteAll(t *testing.T) {
	client := newTestClient(t)

	// Create test objects with different prefixes
	objects := []struct {
		id          id.ID
		name        string
		contentType string
		prefix      string
	}{
		{id.NewID(), "asset1", "application/json", "test-prefix"},
		{id.NewID(), "asset2", "application/json", "test-prefix"},
		{id.NewID(), "asset3", "application/json", "other-prefix"},
	}

	for _, obj := range objects {
		objPath := path.Join(client.basePath, obj.prefix, obj.id.String())
		client.mockBucket.objects[objPath] = &mockObject{
			bucket: client.mockBucket,
			name:   objPath,
			attrs: &storage.ObjectAttrs{
				Name: objPath,
				Metadata: map[string]string{
					"name":         obj.name,
					"content_type": obj.contentType,
				},
			},
		}
	}

	// Delete objects with test-prefix
	err := client.DeleteAll(context.Background(), "test-prefix")
	assert.NoError(t, err)

	// Verify only objects with test-prefix are deleted
	var remainingCount int
	for name := range client.mockBucket.objects {
		if strings.Contains(name, "test-prefix") {
			t.Errorf("Object with test-prefix should be deleted: %s", name)
		}
		remainingCount++
	}
	assert.Equal(t, 1, remainingCount, "Should have one object remaining with other-prefix")
}

func TestClient_Move(t *testing.T) {
	client := newTestClient(t)

	fromID := id.NewID()
	toID := id.NewID()
	content := []byte("test content")
	fromPath := client.objectPath(fromID)
	toPath := client.objectPath(toID)

	client.mockBucket.objects[fromPath] = &mockObject{
		bucket: client.mockBucket,
		name:   fromPath,
		data:   content,
		attrs: &storage.ObjectAttrs{
			Name: fromPath,
			Metadata: map[string]string{
				"name":         "test-asset",
				"content_type": "application/json",
			},
		},
	}

	err := client.Move(context.Background(), fromID, toID)
	assert.NoError(t, err)

	_, exists := client.mockBucket.objects[fromPath]
	assert.False(t, exists)

	obj := client.mockBucket.objects[toPath]
	assert.NotNil(t, obj)
	assert.Equal(t, content, obj.data)
}

func TestClient_Move_SourceNotFound(t *testing.T) {
	client := newTestClient(t)

	err := client.Move(context.Background(), id.NewID(), id.NewID())
	assert.Error(t, err)
}

func TestClient_Move_DestinationExists(t *testing.T) {
	client := newTestClient(t)

	fromID := id.NewID()
	toID := id.NewID()
	fromPath := client.objectPath(fromID)
	toPath := client.objectPath(toID)

	// Create source object
	client.mockBucket.objects[fromPath] = &mockObject{
		bucket: client.mockBucket,
		name:   fromPath,
		attrs: &storage.ObjectAttrs{
			Name:     fromPath,
			Metadata: make(map[string]string),
		},
	}

	// Create destination object
	client.mockBucket.objects[toPath] = &mockObject{
		bucket: client.mockBucket,
		name:   toPath,
		attrs: &storage.ObjectAttrs{
			Name:     toPath,
			Metadata: make(map[string]string),
		},
	}

	err := client.Move(context.Background(), fromID, toID)
	assert.Error(t, err)
}

func TestClient_GetUploadURL(t *testing.T) {
	client := newTestClient(t)

	id := id.NewID()
	objPath := client.objectPath(id)

	url, err := client.GetUploadURL(context.Background(), id)
	assert.NoError(t, err)
	assert.Contains(t, url, objPath)
}
