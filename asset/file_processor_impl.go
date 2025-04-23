package asset

import (
	"net/http"
	"path/filepath"
	"strings"
)

var _ FileProcessor = &fileProcessor{}

type fileProcessor struct{}

func NewFileProcessor() FileProcessor {
	return &fileProcessor{}
}

func (p *fileProcessor) DetectContentType(filename string, data []byte) string {
	if len(data) > 0 {
		detectSize := 512
		if len(data) < detectSize {
			detectSize = len(data)
		}
		contentType := http.DetectContentType(data[:detectSize])
		if idx := strings.Index(contentType, ";"); idx > 0 {
			return contentType[:idx]
		}
		return contentType
	}

	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	case ".gz":
		return "application/gzip"
	case ".tar":
		return "application/x-tar"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".obj":
		return "model/obj"
	case ".gltf":
		return "model/gltf+json"
	case ".glb":
		return "model/gltf-binary"
	case ".geojson":
		return "application/geo+json"
	case ".kml":
		return "application/vnd.google-earth.kml+xml"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}

func (p *fileProcessor) DetectPreviewType(filename string, contentType string) PreviewType {
	if contentType == "" {
		contentType = p.DetectContentType(filename, nil)
	}

	ext := strings.ToLower(filepath.Ext(filename))

	if strings.HasPrefix(contentType, "image/") {
		if ext == ".svg" || contentType == "image/svg+xml" {
			return PreviewTypeImageSVG
		}
		return PreviewTypeImage
	}

	if ext == ".obj" || ext == ".gltf" || ext == ".glb" ||
		strings.HasPrefix(contentType, "model/") {
		return PreviewType3DModel
	}

	if ext == ".geojson" || contentType == "application/geo+json" {
		return PreviewTypeGeo
	}

	if ext == ".kml" || contentType == "application/vnd.google-earth.kml+xml" {
		return PreviewTypeGeo
	}

	if strings.Contains(filename, "tileset.json") ||
		strings.Contains(filename, "3dtiles") {
		return PreviewTypeGeo3DTiles
	}

	if ext == ".mvt" || strings.Contains(filename, "mvt") {
		return PreviewTypeGeoMVT
	}

	if ext == ".csv" || contentType == "text/csv" {
		return PreviewTypeCSV
	}

	return PreviewTypeUnknown
}
