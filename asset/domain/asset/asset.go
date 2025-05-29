package asset

//go:generate go run ./tools/gqlgen/main.go

import (
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/usecasex"

	"github.com/reearth/reearthx/account/accountdomain"
)

type Asset struct {
	id                      id.ID
	projectID               *id.ProjectID
	workspaceID             *id.WorkspaceID
	createdAt               time.Time
	user                    *accountdomain.UserID
	size                    int64
	thread                  *id.ThreadID
	contentType             string // visualizer && flow
	contentEncoding         string
	previewType             *PreviewType // cms
	uuid                    string       //cms
	url                     string
	fileName                string            //cms
	archiveExtractionStatus *ExtractionStatus //cms
	flatFiles               bool              //cms
	integration             *id.IntegrationID
	public                  bool                //cms
	accessInfoResolver      *AccessInfoResolver // cms
}

// cms
type AccessInfoResolver = func(*Asset) *AccessInfo

// cms
type AccessInfo struct {
	Url    string
	Public bool
}

func NewAsset(id id.ID, projectID *id.ProjectID, workspaceID *id.WorkspaceID, createdAt time.Time, size int64, contentType string) *Asset {
	return &Asset{
		id:          id,
		projectID:   projectID,
		workspaceID: workspaceID,
		createdAt:   createdAt,
		size:        size,
		contentType: contentType,
	}
}

type ProjectFilter struct {
	Readable id.ProjectIDList
	Writable id.ProjectIDList
}

func (f *ProjectFilter) CanRead(id *id.ProjectID) bool {
	if id == nil {
		return false
	}
	return f.Readable == nil || f.Readable.Has(*id) || f.CanWrite(id)
}

func (f *ProjectFilter) CanWrite(id *id.ProjectID) bool {
	if id == nil {
		return false
	}
	return f.Writable == nil || f.Writable.Has(*id)
}

func (f *ProjectFilter) Merge(g *ProjectFilter) *ProjectFilter {
	var r, w id.ProjectIDList
	if f.Readable != nil || g.Readable != nil {
		if f.Readable == nil {
			r = g.Readable.Clone()
		} else {
			r = append(f.Readable, g.Readable...)
		}
	}
	if f.Writable != nil || g.Writable != nil {
		if f.Writable == nil {
			w = g.Writable.Clone()
		} else {
			w = append(f.Writable, g.Writable...)
		}
	}
	return &ProjectFilter{
		Readable: r,
		Writable: w,
	}
}

type WorkspaceFilter struct {
	Readable id.WorkspaceIDList
	Writable id.WorkspaceIDList
}

func (f *WorkspaceFilter) CanRead(id *id.WorkspaceID) bool {
	if id == nil {
		return false
	}
	return f.Readable == nil || f.Readable.Has(*id) || f.CanWrite(id)
}

func (f *WorkspaceFilter) CanWrite(id *id.WorkspaceID) bool {
	if id == nil {
		return false
	}
	return f.Writable == nil || f.Writable.Has(*id)
}

func (f *WorkspaceFilter) Merge(g *WorkspaceFilter) *WorkspaceFilter {
	var r, w id.WorkspaceIDList
	if f.Readable != nil || g.Readable != nil {
		if f.Readable == nil {
			r = g.Readable.Clone()
		} else {
			r = append(f.Readable, g.Readable...)
		}
	}
	if f.Writable != nil || g.Writable != nil {
		if f.Writable == nil {
			w = g.Writable.Clone()
		} else {
			w = append(f.Writable, g.Writable...)
		}
	}
	return &WorkspaceFilter{
		Readable: r,
		Writable: w,
	}
}

func (a *Asset) IsPublic() bool {
	if a == nil {
		return false
	}
	return a.public

}

func (a *Asset) ID() id.ID {
	return a.id
}

func (a *Asset) Clone() *Asset {
	return &Asset{
		id:                      a.id,
		projectID:               a.projectID,
		workspaceID:             a.workspaceID,
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
		public:                  a.public,
		accessInfoResolver:      a.accessInfoResolver,
	}
}

func (a *Asset) ProjectID() *id.ProjectID {
	return a.projectID
}

func (a *Asset) WorkspaceID() *id.WorkspaceID {
	return a.workspaceID
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

func (a *Asset) Integration() id.IntegrationID {
	return a.integration
}

func (a *Asset) User() *accountdomain.UserID {
	return a.user
}

func (a *Asset) Thread() *id.ThreadID {
	return a.thread
}

func (a *Asset) Public() bool {
	return a.public
}

// Setter methods for modifying private fields
func (a *Asset) SetID(id id.ID) {
	a.id = id
}

func (a *Asset) SetProjectID(projectID *id.ProjectID) {
	a.projectID = projectID
}

func (a *Asset) SetWorkspaceID(workspaceID *id.WorkspaceID) {
	a.workspaceID = workspaceID
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

func (a *Asset) AddIntegration(integrationID id.IntegrationID) {
	a.integration = integrationID
}

func (a *Asset) SetUser(user *accountdomain.UserID) {
	a.user = user
}

func (a *Asset) SetThread(thread *id.ThreadID) {
	a.thread = thread
}

func (a *Asset) SetPublic(public bool) {
	a.public = public
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

type SortType string

const (
	SortTypeDate SortType = "DATE"
	SortTypeSize SortType = "SIZE"
	SortTypeName SortType = "NAME"
)

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

type Sort struct {
	By        SortType
	Direction SortDirection
}

type Pagination struct {
	Offset int64
	Limit  int64
}

type Filter struct {
	Sort         *usecasex.Sort
	Keyword      *string
	Pagination   *usecasex.Pagination
	ContentTypes []string
}
