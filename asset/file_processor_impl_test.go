package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectContentType(t *testing.T) {
	processor := NewFileProcessor()

	testCases := []struct {
		name         string
		filename     string
		data         []byte
		expectedType string
	}{
		{
			name:         "Text file by extension",
			filename:     "test.txt",
			data:         nil,
			expectedType: "text/plain",
		},
		{
			name:         "JSON file by extension",
			filename:     "data.json",
			data:         nil,
			expectedType: "application/json",
		},
		{
			name:         "HTML file by extension",
			filename:     "page.html",
			data:         nil,
			expectedType: "text/html",
		},
		{
			name:         "PNG image by extension",
			filename:     "image.png",
			data:         nil,
			expectedType: "image/png",
		},
		{
			name:         "JPEG image by extension",
			filename:     "photo.jpg",
			data:         nil,
			expectedType: "image/jpeg",
		},
		{
			name:         "SVG image by extension",
			filename:     "vector.svg",
			data:         nil,
			expectedType: "image/svg+xml",
		},
		{
			name:         "PDF document by extension",
			filename:     "document.pdf",
			data:         nil,
			expectedType: "application/pdf",
		},
		{
			name:         "ZIP archive by extension",
			filename:     "archive.zip",
			data:         nil,
			expectedType: "application/zip",
		},
		{
			name:         "GeoJSON file by extension",
			filename:     "map.geojson",
			data:         nil,
			expectedType: "application/geo+json",
		},
		{
			name:         "KML file by extension",
			filename:     "map.kml",
			data:         nil,
			expectedType: "application/vnd.google-earth.kml+xml",
		},
		{
			name:         "Unknown extension",
			filename:     "file.xyz",
			data:         nil,
			expectedType: "application/octet-stream",
		},
		{
			name:         "No extension",
			filename:     "README",
			data:         nil,
			expectedType: "application/octet-stream",
		},
		{
			name:         "With content sniffing",
			filename:     "unknown",
			data:         []byte("<html><body>Test</body></html>"),
			expectedType: "text/html",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contentType := processor.DetectContentType(tc.filename, tc.data)
			assert.Equal(t, tc.expectedType, contentType)
		})
	}
}

func TestDetectPreviewType(t *testing.T) {
	processor := NewFileProcessor()

	testCases := []struct {
		name            string
		filename        string
		contentType     string
		expectedPreview PreviewType
	}{
		{
			name:            "JPEG image",
			filename:        "photo.jpg",
			contentType:     "image/jpeg",
			expectedPreview: PreviewTypeImage,
		},
		{
			name:            "PNG image",
			filename:        "graphic.png",
			contentType:     "image/png",
			expectedPreview: PreviewTypeImage,
		},
		{
			name:            "SVG image",
			filename:        "vector.svg",
			contentType:     "image/svg+xml",
			expectedPreview: PreviewTypeImageSVG,
		},
		{
			name:            "GeoJSON file",
			filename:        "map.geojson",
			contentType:     "application/geo+json",
			expectedPreview: PreviewTypeGeo,
		},
		{
			name:            "KML file",
			filename:        "map.kml",
			contentType:     "application/vnd.google-earth.kml+xml",
			expectedPreview: PreviewTypeGeo,
		},
		{
			name:            "3D Tiles",
			filename:        "model.3dtiles",
			contentType:     "application/json",
			expectedPreview: PreviewTypeGeo3DTiles,
		},
		{
			name:            "MVT file",
			filename:        "tiles.mvt",
			contentType:     "application/vnd.mapbox-vector-tile",
			expectedPreview: PreviewTypeGeoMVT,
		},
		{
			name:            "3D Model",
			filename:        "model.glb",
			contentType:     "model/gltf-binary",
			expectedPreview: PreviewType3DModel,
		},
		{
			name:            "CSV file",
			filename:        "data.csv",
			contentType:     "text/csv",
			expectedPreview: PreviewTypeCSV,
		},
		{
			name:            "Text file",
			filename:        "README.txt",
			contentType:     "text/plain",
			expectedPreview: PreviewTypeUnknown,
		},
		{
			name:            "Unknown type",
			filename:        "data.bin",
			contentType:     "application/octet-stream",
			expectedPreview: PreviewTypeUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			previewType := processor.DetectPreviewType(tc.filename, tc.contentType)
			assert.Equal(t, tc.expectedPreview, previewType)
		})
	}
}
