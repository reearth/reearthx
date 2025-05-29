package mongodoc

import (
	asset2 "github.com/reearth/reearthx/asset/domain/asset"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/idx"
	"github.com/reearth/reearthx/mongox"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
)

type AssetDocument struct {
	ID                      string    `bson:"id"`
	Project                 string    `bson:"groupid"`
	CreatedAt               time.Time `bson:"createdat"`
	User                    *string   `bson:"user,omitempty"`
	Integration             *string   `bson:"integrationid,omitempty"`
	FileName                string    `bson:"filename"`
	Size                    uint64    `bson:"size"`
	ContentType             string    `bson:"contenttype"`
	ContentEncoding         string    `bson:"contentencoding,omitempty"`
	PreviewType             string    `bson:"previewtype,omitempty"`
	UUID                    string    `bson:"uuid"`
	URL                     string    `bson:"url"`
	Thread                  *string   `bson:"thread,omitempty"`
	ArchiveExtractionStatus string    `bson:"archiveextractionstatus,omitempty"`
	FlatFiles               bool      `bson:"flatfiles"`
	Public                  bool      `bson:"public"`
}

type AssetAndFileDocument struct {
	ID        string
	File      *AssetFileDocument
	FlatFiles bool
}

type AssetFileDocument struct {
	Name            string
	Size            uint64
	ContentType     string
	ContentEncoding string
	Path            string
	Children        []*AssetFileDocument
}

type AssetConsumer = mongox.SliceFuncConsumer[*AssetDocument, *asset2.Asset]

type AssetAndFileConsumer = mongox.SliceConsumer[*AssetAndFileDocument]

type AssetFilesConsumer struct {
	c mongox.SliceConsumer[*AssetFilesPageDocument]
}

func (a *AssetFilesConsumer) Consume(raw bson.Raw) error {
	return a.c.Consume(raw)
}

func (a *AssetFilesConsumer) Result() AssetFilesDocument {
	return a.c.Result
}

func NewAssetConsumer() *AssetConsumer {
	return NewConsumer[*AssetDocument, *asset2.Asset]()
}

func NewAsset(a *asset2.Asset) (*AssetDocument, string) {
	aid := a.ID().String()

	previewType := ""
	if pt := a.PreviewType(); pt != nil {
		previewType = string(*pt)
	}

	archiveExtractionStatus := ""
	if s := a.ArchiveExtractionStatus(); s != nil {
		archiveExtractionStatus = string(*s)
	}

	var uid, tid *string
	if a.User() != nil {
		userStr := a.User().String()
		uid = &userStr
	}
	if a.Thread() != nil {
		threadStr := a.Thread().String()
		tid = &threadStr
	}

	// Integration is exposed but it's not a pointer
	var iid *string
	integrationStr := a.Integration().String()
	if integrationStr != "" {
		iid = &integrationStr
	}

	return &AssetDocument{
		ID:                      aid,
		Project:                 a.GroupID().String(), // GroupID serves as Project
		CreatedAt:               a.CreatedAt(),
		User:                    uid,
		Integration:             iid,
		FileName:                a.FileName(),
		Size:                    uint64(a.Size()),
		ContentType:             a.ContentType(),
		ContentEncoding:         a.ContentEncoding(),
		PreviewType:             previewType,
		UUID:                    a.UUID(),
		URL:                     a.URL(),
		Thread:                  tid,
		ArchiveExtractionStatus: archiveExtractionStatus,
		FlatFiles:               a.FlatFiles(),
		Public:                  a.Public(),
	}, aid
}

func (d *AssetDocument) Model() (*asset2.Asset, error) {
	aid, err := asset2.IdFrom(d.ID)
	if err != nil {
		return nil, err
	}
	groupID, err := asset2.GroupIDFrom(d.Project)
	if err != nil {
		return nil, err
	}

	a := asset2.NewAsset(aid, &groupID, d.CreatedAt, int64(d.Size), d.ContentType)

	a.SetFileName(d.FileName)
	a.SetUUID(d.UUID)
	a.SetURL(d.URL)
	a.SetContentEncoding(d.ContentEncoding)
	a.SetFlatFiles(d.FlatFiles)
	a.SetPublic(d.Public)

	if d.PreviewType != "" {
		pt := asset2.PreviewType(d.PreviewType)
		a.SetPreviewType(pt)
	}

	if d.ArchiveExtractionStatus != "" {
		status := asset2.ExtractionStatus(d.ArchiveExtractionStatus)
		a.SetArchiveExtractionStatus(&status)
	}

	if d.Integration != nil {
		iid, err := idx.From[asset2.IntegrationIDType](*d.Integration)
		if err != nil {
			return nil, err
		}
		a.AddIntegration(iid)
	}

	if d.User != nil {
		uid, err := accountdomain.UserIDFrom(*d.User)
		if err != nil {
			return nil, err
		}
		a.SetUser(&uid)
	}

	if d.Thread != nil {
		tid, err := idx.From[asset2.ThreadIDType](*d.Thread)
		if err != nil {
			return nil, err
		}
		a.SetThread(&tid)
	}

	return a, nil
}

func NewFile(f *asset2.File) *AssetFileDocument {
	if f == nil {
		return nil
	}

	var c []*AssetFileDocument
	if len(f.Children()) > 0 {
		for _, v := range f.Children() {
			c = append(c, NewFile(v))
		}
	}

	return &AssetFileDocument{
		Name:            f.Name(),
		Size:            f.Size(),
		ContentType:     f.ContentType(),
		ContentEncoding: f.ContentEncoding(),
		Path:            f.Path(),
		Children:        c,
	}
}

func (f *AssetFileDocument) Model() *asset2.File {
	if f == nil {
		return nil
	}

	var c []*asset2.File
	if len(f.Children) > 0 {
		for _, v := range f.Children {
			childFile := v.Model()
			c = append(c, childFile)
		}
	}

	af := asset2.NewFile().
		Name(f.Name).
		Size(f.Size).
		ContentType(f.ContentType).
		Path(f.Path).
		Children(c).
		Build()

	return af
}

type AssetFilesDocument []*AssetFilesPageDocument

func (d AssetFilesDocument) totalFiles() int {
	size := 0
	for _, page := range d {
		size += len(page.Files)
	}
	return size
}

func (d AssetFilesDocument) Model() []*asset2.File {
	files := make([]*asset2.File, 0, d.totalFiles())
	for _, page := range d {
		files = append(files, lo.Map(page.Files, func(f *AssetFileDocument, _ int) *asset2.File {
			return f.Model()
		})...)
	}
	return files
}

type AssetFilesPageDocument struct {
	AssetID string
	Page    int
	Files   []*AssetFileDocument
}

const assetFilesPageSize = 1000

func NewFiles(assetID asset2.ID, fs []*asset2.File) AssetFilesDocument {
	pageCount := (len(fs) + assetFilesPageSize - 1) / assetFilesPageSize
	pages := make([]*AssetFilesPageDocument, 0, pageCount)
	for i := 0; i < pageCount; i++ {
		offset := i * assetFilesPageSize
		chunk := fs[offset:]
		if len(chunk) > assetFilesPageSize {
			chunk = chunk[:assetFilesPageSize]
		}
		pages = append(pages, &AssetFilesPageDocument{
			AssetID: assetID.String(),
			Page:    i,
			Files: lo.Map(chunk, func(f *asset2.File, _ int) *AssetFileDocument {
				return NewFile(f)
			}),
		})
	}
	return pages
}

func NewAssetAndFileConsumer() *AssetAndFileConsumer {
	return &AssetAndFileConsumer{}
}
