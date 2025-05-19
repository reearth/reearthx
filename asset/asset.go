package asset

//go:generate go run ./tools/gqlgen/main.go

import (
	"time"
)

type Asset struct {
	ID                      AssetID
	GroupID                 GroupID   // projectID in visualizer and cms, workspaceID in flow
	CreatedAt               time.Time //
	CreatedBy               OperatorInfo
	Size                    int64
	ContentType             string // visualizer && flow
	ContentEncoding         string
	PreviewType             PreviewType       // cms
	UUID                    string            //cms
	URL                     string            //cms && visualizer && flow
	FileName                string            //cms
	ArchiveExtractionStatus *ExtractionStatus //cms
	FlatFiles               bool              //cms
	//Thread ThreadID //cms
}

type PreviewType string

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

type OperatorType string

const (
	OperatorTypeUser        OperatorType = "USER"        //cms
	OperatorTypeIntegration OperatorType = "INTEGRATION" //cms
)

type OperatorInfo struct {
	Type OperatorType
	ID   string
}

type AssetFile struct {
	Name            string
	Size            int64
	ContentType     string
	ContentEncoding string
	Path            string
	FilePaths       []string
}
