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
	testID := domain.NewID()
	asset := domain.NewAsset(testID, "test-name", 100, "test/type")
	obj := mock.getObject("test-path/" + testID.String())
	obj.metadata = map[string]string{
		"name":         asset.Name(),
		"content_type": asset.ContentType(),
	}

	// Test Read
	attrs := &storage.ObjectAttrs{
		Metadata: obj.metadata,
	}
	readAsset := domain.NewAsset(
		testID,
		attrs.Metadata["name"],
		100,
		attrs.Metadata["content_type"],
	)
	assert.Equal(t, asset.ID(), readAsset.ID())
	assert.Equal(t, asset.Name(), readAsset.Name())
	assert.Equal(t, asset.ContentType(), readAsset.ContentType())

	// Test Update
	updatedAsset := domain.NewAsset(testID, "updated-name", 100, "updated/type")
	obj.metadata = map[string]string{
		"name":         updatedAsset.Name(),
		"content_type": updatedAsset.ContentType(),
	}
	attrs = &storage.ObjectAttrs{
		Metadata: obj.metadata,
	}
	assert.Equal(t, updatedAsset.Name(), attrs.Metadata["name"])
	assert.Equal(t, updatedAsset.ContentType(), attrs.Metadata["content_type"])

	// Test Delete
	delete(mock.objects, "test-path/"+testID.String())
	_, exists := mock.objects["test-path/"+testID.String()]
	assert.False(t, exists)

	// Test Upload
	content := "test content"
	obj = mock.getObject("test-path/" + testID.String())
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
	id1 := domain.NewID()
	id2 := domain.NewID()
	mock.getObject("test-path/" + id1.String())
	mock.getObject("test-path/" + id2.String())

	assert.Len(t, mock.objects, 2)
	assert.Contains(t, mock.objects, "test-path/"+id1.String())
	assert.Contains(t, mock.objects, "test-path/"+id2.String())
}

func TestClient_Move(t *testing.T) {
	mock := newMockClient()

	// Setup source object
	sourceID := domain.NewID()
	destID := domain.NewID()
	sourceObj := mock.getObject("test-path/" + sourceID.String())
	sourceObj.content = "test content"

	// Move object
	destObj := mock.getObject("test-path/" + destID.String())
	destObj.content = sourceObj.content
	delete(mock.objects, "test-path/"+sourceID.String())

	// Verify
	assert.NotContains(t, mock.objects, "test-path/"+sourceID.String())
	assert.Contains(t, mock.objects, "test-path/"+destID.String())
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

	id := domain.NewID()
	url := client.GetObjectURL(id)
	assert.Equal(t, "https://example.com/test-path/"+id.String(), url)

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

	validID := domain.NewID()
	// Get the empty ID that will be used for error cases
	emptyID := domain.NewID()

	tests := []struct {
		name    string
		url     string
		wantID  domain.ID
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "https://example.com/test-path/" + validID.String(),
			wantID:  validID,
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "://invalid-url",
			wantID:  emptyID,
			wantErr: true,
		},
		{
			name:    "different host",
			url:     "https://different.com/test-path/" + validID.String(),
			wantID:  emptyID,
			wantErr: true,
		},
		{
			name:    "empty path",
			url:     "https://example.com",
			wantID:  emptyID,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := client.GetIDFromURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				if !tt.wantErr {
					assert.Equal(t, tt.wantID, id)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, id)
			}
		})
	}

	// Test with nil baseURL
	client.baseURL = nil
	_, err := client.GetIDFromURL("https://example.com/test-path/" + validID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL not set")
}

func TestClient_objectPath(t *testing.T) {
	client := &Client{
		bucketName: "test-bucket",
		basePath:   "test-path",
	}

	id := domain.NewID()
	path := client.objectPath(id)
	assert.Equal(t, "test-path/"+id.String(), path)
}
