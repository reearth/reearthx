package asset

import (
	"time"

	"github.com/google/uuid"
	"github.com/reearth/reearthx/account/accountdomain"
)

type Builder struct {
	a *Asset
}

func New() *Builder {
	a := &Asset{}
	a.public = false
	return &Builder{a: a}
}

func (b *Builder) Build() (*Asset, error) {
	if b.a.id.IsNil() {
		return nil, ErrInvalidID
	}
	if b.a.project.IsNil() {
		return nil, ErrNoProjectID
	}
	if b.a.user.IsNil() && b.a.integration.IsNil() {
		return nil, ErrNoUser
	}
	if b.a.size == 0 {
		return nil, ErrZeroSize
	}
	if b.a.uuid == "" {
		return nil, ErrNoUUID
	}
	if b.a.createdAt.IsZero() {
		b.a.createdAt = b.a.id.Timestamp()
	}
	return b.a, nil
}

func (b *Builder) MustBuild() *Asset {
	r, err := b.Build()
	if err != nil {
		panic(err)
	}
	return r
}

func (b *Builder) ID(id ID) *Builder {
	b.a.id = id
	return b
}

func (b *Builder) NewID() *Builder {
	b.a.id = NewID()
	return b
}

func (b *Builder) Name(name string) *Builder {
	b.a.name = name
	return b
}

func (b *Builder) CoreSupport(coreSupport bool) *Builder {
	b.a.coreSupport = coreSupport
	return b
}

func (b *Builder) ContentType(contentType string) *Builder {
	b.a.contentType = contentType
	return b
}

func (b *Builder) CreatedBy(createdBy accountdomain.UserID) *Builder {
	b.a.user = &createdBy
	b.a.integration = nil
	return b
}

func (b *Builder) Workspace(workspace accountdomain.WorkspaceID) *Builder {
	b.a.workspace = workspace
	return b
}

func (b *Builder) Project(pid ProjectID) *Builder {
	b.a.project = pid
	return b
}

func (b *Builder) CreatedAt(createdAt time.Time) *Builder {
	b.a.createdAt = createdAt
	return b
}

func (b *Builder) CreatedByUser(createdBy accountdomain.UserID) *Builder {
	b.a.user = &createdBy
	b.a.integration = nil
	return b
}

func (b *Builder) CreatedByIntegration(createdBy IntegrationID) *Builder {
	b.a.integration = &createdBy
	b.a.user = nil
	return b
}

func (b *Builder) FileName(name string) *Builder {
	b.a.fileName = name
	return b
}

func (b *Builder) Size(size uint64) *Builder {
	b.a.size = size
	return b
}

func (b *Builder) URL(url string) *Builder {
	b.a.url = url
	return b
}

func (b *Builder) Type(t *PreviewType) *Builder {
	b.a.previewType = t
	return b
}

func (b *Builder) UUID(uuid string) *Builder {
	b.a.uuid = uuid
	return b
}

func (b *Builder) NewUUID() *Builder {
	b.a.uuid = uuid.NewString()
	return b
}

func (b *Builder) Thread(th *ThreadID) *Builder {
	b.a.thread = th
	return b
}

func (b *Builder) ArchiveExtractionStatus(s *ArchiveExtractionStatus) *Builder {
	b.a.archiveExtractionStatus = s
	return b
}

func (b *Builder) FlatFiles(flatFiles bool) *Builder {
	b.a.flatFiles = flatFiles
	return b
}

func (b *Builder) Public(public bool) *Builder {
	b.a.public = public
	return b
}
