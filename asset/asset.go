package asset

//go:generate go run ./tools/gqlgen/main.go

import (
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
)

type Asset struct {
	id                      AssetID
	groupID                 GroupID // projectID in visualizer and cms, workspaceID in flow
	createdAt               time.Time
	user                    *accountdomain.UserID
	size                    int64
	thread                  *ThreadID
	contentType             string // visualizer && flow
	contentEncoding         string
	previewType             *PreviewType      // cms
	uuid                    string            //cms
	url                     string            //cms && visualizer && flow
	fileName                string            //cms
	archiveExtractionStatus *ExtractionStatus //cms
	flatFiles               bool              //cms
	integration             IntegrationID
	accessInfoResolver      *AccessInfoResolver // cms
}

type AccessInfoResolver = func(*Asset) *AccessInfo

type AccessInfo struct {
	Url    string
	Public bool
}

func NewAsset(id AssetID, groupID GroupID, createdAt time.Time, size int64, contentType string) *Asset {
	return &Asset{
		id:          id,
		groupID:     groupID,
		createdAt:   createdAt,
		size:        size,
		contentType: contentType,
	}
}

type ProjectFilter struct {
	Readable GroupIDList
	Writable GroupIDList
}

func (a *Asset) ID() AssetID {
	return a.id
}

func (a *Asset) Clone() *Asset {
	return &Asset{
		id:                      a.id,
		groupID:                 a.groupID,
		createdAt:               a.createdAt,
		size:                    a.size,
		contentType:             a.contentType,
		contentEncoding:         a.contentEncoding,
		previewType:             a.previewType,
		uuid:                    a.uuid,
		url:                     a.url,
		fileName:                a.fileName,
		archiveExtractionStatus: a.archiveExtractionStatus,
		flatFiles:               a.flatFiles,
		integration:             a.integration,
		accessInfoResolver:      a.accessInfoResolver,
	}
}

func (a *Asset) GroupID() GroupID {
	return a.groupID
}

func (a *Asset) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Asset) Size() int64 {
	return a.size
}

func (a *Asset) ContentType() string {
	return a.contentType
}

func (a *Asset) ContentEncoding() string {
	return a.contentEncoding
}

func (a *Asset) PreviewType() *PreviewType {
	return a.previewType
}

func (a *Asset) UUID() string {
	return a.uuid
}

func (a *Asset) URL() string {
	return a.url
}

func (a *Asset) FileName() string {
	return a.fileName
}

func (a *Asset) ArchiveExtractionStatus() *ExtractionStatus {
	return a.archiveExtractionStatus
}

func (a *Asset) FlatFiles() bool {
	return a.flatFiles
}

func (a *Asset) Integration() IntegrationID {
	return a.integration
}

// Setter methods for modifying private fields
func (a *Asset) SetID(id AssetID) {
	a.id = id
}

func (a *Asset) SetGroupID(groupID GroupID) {
	a.groupID = groupID
}

func (a *Asset) SetCreatedAt(createdAt time.Time) {
	a.createdAt = createdAt
}

func (a *Asset) SetAccessInfoResolver(resolver AccessInfoResolver) {
	if resolver == nil {
		a.accessInfoResolver = nil
		return
	}
	a.accessInfoResolver = &resolver
}

func (a *Asset) SetSize(size int64) {
	a.size = size
}

func (a *Asset) SetContentType(contentType string) {
	a.contentType = contentType
}

func (a *Asset) SetContentEncoding(contentEncoding string) {
	a.contentEncoding = contentEncoding
}

func (a *Asset) SetPreviewType(previewType PreviewType) {
	a.previewType = &previewType
}

func (a *Asset) SetUUID(uuid string) {
	a.uuid = uuid
}

func (a *Asset) SetURL(url string) {
	a.url = url
}

func (a *Asset) SetFileName(fileName string) {
	a.fileName = fileName
}

func (a *Asset) SetArchiveExtractionStatus(status *ExtractionStatus) {
	a.archiveExtractionStatus = status
}

func (a *Asset) SetFlatFiles(flatFiles bool) {
	a.flatFiles = flatFiles
}

func (a *Asset) AddIntegration(integrationID IntegrationID) {
	a.integration = integrationID
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
