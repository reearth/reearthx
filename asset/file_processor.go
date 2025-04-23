package asset

type FileProcessor interface {
	DetectContentType(filename string, data []byte) string
	DetectPreviewType(filename string, contentType string) PreviewType
}
