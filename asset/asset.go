package asset

//go:generate go run ./tools/gqlgen/main.go

import (
	"time"
)

type Asset struct {
	ID                      AssetID
	GroupID                 GroupID   // projectID in visualizer and cms, workspaceID in flow
	CreatedAt               time.Time //
	Size                    int64
	ContentType             string // visualizer && flow
	ContentEncoding         string
	PreviewType             PreviewType       // cms
	UUID                    string            //cms
	URL                     string            //cms && visualizer && flow
	FileName                string            //cms
	ArchiveExtractionStatus *ExtractionStatus //cms
	FlatFiles               bool              //cms
	Integration             IntegrationID
}

type PreviewType string

// CMS
const (
	PreviewTypeImage      PreviewType = "IMAGE"
	PreviewTypeImageSVG   PreviewType = "IMAGE_SVG"
	PreviewTypeGeo        PreviewType = "GEO"
	PreviewTypeGeo3DTiles PreviewType = "GEO_3D_TILES"
	PreviewTypeGeoMVT     PreviewType = "GEO_MVT"
	PreviewType3DModel    PreviewType = "MODEL_3D"
	PreviewTypeCSV        PreviewType = "CSV"
	PreviewTypeUnknown    PreviewType = "UNKNOWN"
)

type ExtractionStatus string

const (
	ExtractionStatusSkipped    ExtractionStatus = "SKIPPED"
	ExtractionStatusPending    ExtractionStatus = "PENDING"
	ExtractionStatusInProgress ExtractionStatus = "IN_PROGRESS"
	ExtractionStatusDone       ExtractionStatus = "DONE"
	ExtractionStatusFailed     ExtractionStatus = "FAILED"
)
