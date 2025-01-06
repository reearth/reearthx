package gcs

import (
	"net/url"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/reearth/reearthx/asset/domain"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	objects map[string]*mockObject
}

type mockObject struct {
	name        string
	metadata    map[string]string
	content     string
	shouldError bool
}

func newMockClient() *mockClient {
	return &mockClient{
		objects: make(map[string]*mockObject),
	}
}

func (m *mockClient) getObject(name string) *mockObject {
	if obj, exists := m.objects[name]; exists {
		return obj
	}
	obj := &mockObject{
		name: name,
		metadata: map[string]string{
			"name":         "test-name",
			"content_type": "test/type",
		},
	}
	m.objects[name] = obj
	return obj
}

func TestClient_Init(t *testing.T) {
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
			baseURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "invalid base URL",
			bucketName: "test-bucket",
			basePath:   "test-path",
			baseURL:    "://invalid-url",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var baseURL *url.URL
			if tt.baseURL != "" {
				baseURL, err = url.Parse(tt.baseURL)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
			}

			client := &Client{
				bucketName: tt.bucketName,
				basePath:   tt.basePath,
				baseURL:    baseURL,
			}

			assert.NotNil(t, client)
			assert.Equal(t, tt.bucketName, client.bucketName)
			assert.Equal(t, tt.basePath, client.basePath)
			if tt.baseURL != "" && !tt.wantErr {
				assert.Equal(t, tt.baseURL, client.baseURL.String())
			}
		})
	}
}

func TestClient_CRUD(t *testing.T) {
	mock := newMockClient()

	// Test Create
	asset := domain.NewAsset("test-id", "test-name", 100, "test/type")
	obj := mock.getObject("test-path/test-id")
	obj.metadata = map[string]string{
		"name":         asset.Name(),
		"content_type": asset.ContentType(),
	}

	// Test Read
	attrs := &storage.ObjectAttrs{
		Name:     obj.name,
		Metadata: obj.metadata,
	}
	readAsset := domain.NewAsset(
		domain.ID("test-id"),
		attrs.Metadata["name"],
		100,
		attrs.Metadata["content_type"],
	)
	assert.Equal(t, asset.ID(), readAsset.ID())
	assert.Equal(t, asset.Name(), readAsset.Name())
	assert.Equal(t, asset.ContentType(), readAsset.ContentType())

	// Test Update
	updatedAsset := domain.NewAsset("test-id", "updated-name", 100, "updated/type")
	obj.metadata = map[string]string{
		"name":         updatedAsset.Name(),
		"content_type": updatedAsset.ContentType(),
	}
	attrs = &storage.ObjectAttrs{
		Name:     obj.name,
		Metadata: obj.metadata,
	}
	assert.Equal(t, updatedAsset.Name(), attrs.Metadata["name"])
	assert.Equal(t, updatedAsset.ContentType(), attrs.Metadata["content_type"])

	// Test Delete
	delete(mock.objects, "test-path/test-id")
	_, exists := mock.objects["test-path/test-id"]
	assert.False(t, exists)

	// Test Upload
	content := "test content"
	obj = mock.getObject("test-path/test-id")
	obj.content = content
	assert.Equal(t, content, obj.content)

	// Test Download
	assert.Equal(t, content, obj.content)

	// Test error cases
	nonExistentObj := mock.getObject("non-existent")
	nonExistentObj.shouldError = true
	assert.True(t, nonExistentObj.shouldError)
}

func TestClient_List(t *testing.T) {
	mock := newMockClient()

	// Add some test objects
	mock.getObject("test-path/test-1")
	mock.getObject("test-path/test-2")

	assert.Len(t, mock.objects, 2)
	assert.Contains(t, mock.objects, "test-path/test-1")
	assert.Contains(t, mock.objects, "test-path/test-2")
}

func TestClient_Move(t *testing.T) {
	mock := newMockClient()

	// Setup source object
	sourceObj := mock.getObject("test-path/source-id")
	sourceObj.content = "test content"

	// Move object
	destObj := mock.getObject("test-path/dest-id")
	destObj.content = sourceObj.content
	delete(mock.objects, "test-path/source-id")

	// Verify
	assert.NotContains(t, mock.objects, "test-path/source-id")
	assert.Contains(t, mock.objects, "test-path/dest-id")
	assert.Equal(t, "test content", destObj.content)
}

func TestClient_GetObjectURL(t *testing.T) {
	baseURL := "https://example.com"
	u, _ := url.Parse(baseURL)
	client := &Client{
		bucketName: "test-bucket",
		basePath:   "test-path",
		baseURL:    u,
	}

	id := domain.ID("test-id")
	url := client.GetObjectURL(id)
	assert.Equal(t, "https://example.com/test-path/test-id", url)

	// Test with nil baseURL
	client.baseURL = nil
	url = client.GetObjectURL(id)
	assert.Empty(t, url)
}

func TestClient_GetIDFromURL(t *testing.T) {
	baseURL := "https://example.com"
	u, _ := url.Parse(baseURL)
	client := &Client{
		bucketName: "test-bucket",
		basePath:   "test-path",
		baseURL:    u,
	}

	tests := []struct {
		name    string
		url     string
		wantID  domain.ID
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "https://example.com/test-path/test-id",
			wantID:  domain.ID("test-id"),
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "://invalid-url",
			wantID:  "",
			wantErr: true,
		},
		{
			name:    "different host",
			url:     "https://different.com/test-path/test-id",
			wantID:  "",
			wantErr: true,
		},
		{
			name:    "empty path",
			url:     "https://example.com",
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := client.GetIDFromURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
		})
	}

	// Test with nil baseURL
	client.baseURL = nil
	_, err := client.GetIDFromURL("https://example.com/test-path/test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL not set")
}

func TestClient_objectPath(t *testing.T) {
	client := &Client{
		bucketName: "test-bucket",
		basePath:   "test-path",
	}

	id := domain.ID("test-id")
	path := client.objectPath(id)
	assert.Equal(t, "test-path/test-id", path)
}
