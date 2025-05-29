package asset

//go:generate go run ./tools/gqlgen/main.go

import (
	"github.com/reearth/reearthx/usecasex"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
)

type Asset struct {
	id                      ID
	groupID                 *GroupID // projectID in cms, workspaceID in flow and viz
	createdAt               time.Time
	user                    *accountdomain.UserID
	size                    int64
	thread                  *ThreadID
	contentType             string // visualizer && flow
	contentEncoding         string
	previewType             *PreviewType // cms
	uuid                    string       //cms
	url                     string
	fileName                string            //cms
	archiveExtractionStatus *ExtractionStatus //cms
	flatFiles               bool              //cms
	integration             IntegrationID
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

func NewAsset(id ID, groupID *GroupID, createdAt time.Time, size int64, contentType string) *Asset {
	return &Asset{
		id:          id,
		groupID:     groupID,
		createdAt:   createdAt,
		size:        size,
		contentType: contentType,
	}
}

type GroupFilter struct {
	Readable GroupIDList
	Writable GroupIDList
}

func (f *GroupFilter) CanRead(id GroupID) bool {
	return f.Readable == nil || f.Readable.Has(id) || f.CanWrite(id)
}

func (f *GroupFilter) CanWrite(id GroupID) bool {
	return f.Writable == nil || f.Writable.Has(id)
}

func (f *GroupFilter) Merge(g *GroupFilter) *GroupFilter {
	var r, w GroupIDList
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
	return &GroupFilter{
		Readable: r,
		Writable: w,
	}
}

func (a *Asset) ID() ID {
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
		public:                  a.public,
		accessInfoResolver:      a.accessInfoResolver,
	}
}

func (a *Asset) GroupID() *GroupID {
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

func (a *Asset) User() *accountdomain.UserID {
	return a.user
}

func (a *Asset) Thread() *ThreadID {
	return a.thread
}

func (a *Asset) Public() bool {
	return a.public
}

// Setter methods for modifying private fields
func (a *Asset) SetID(id ID) {
	a.id = id
}

func (a *Asset) SetGroupID(groupID *GroupID) {
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

func (a *Asset) SetUser(user *accountdomain.UserID) {
	a.user = user
}

func (a *Asset) SetThread(thread *ThreadID) {
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

type IDList []ID

func (l IDList) Add(id ID) IDList {
	return append(l, id)
}

func (l IDList) Strings() []string {
	strings := make([]string, len(l))
	for i, id := range l {
		strings[i] = id.String()
	}
	return strings
}

type Filter struct {
	Sort         *usecasex.Sort
	Keyword      *string
	Pagination   *usecasex.Pagination
	ContentTypes []string
}
