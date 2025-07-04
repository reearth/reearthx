package gateway

import (
	"context"
	"io"
	"mime"
	"net/url"
	"path"
	"time"

	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/file"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
)

var (
	ErrInvalidFile                error = rerror.NewE(i18n.T("invalid file"))
	ErrFailedToUploadFile         error = rerror.NewE(i18n.T("failed to upload file"))
	ErrFileTooLarge               error = rerror.NewE(i18n.T("file too large"))
	ErrFailedToDeleteFile         error = rerror.NewE(i18n.T("failed to delete file"))
	ErrFileNotFound               error = rerror.NewE(i18n.T("file not found"))
	ErrUnsupportedOperation       error = rerror.NewE(i18n.T("unsupported operation"))
	ErrUnsupportedContentEncoding error = rerror.NewE(i18n.T("unsupported content encoding"))
	ErrInvalidUUID                error = rerror.NewE(i18n.T("invalid uuid"))
	ErrInvalidInput               error = rerror.NewE(i18n.T("invalid input"))
)

type FileEntry struct {
	Name            string
	ContentType     string
	ContentEncoding string
	Size            int64
}

type UploadAssetLink struct {
	URL             string
	ContentType     string
	ContentEncoding string
	Next            string
	ContentLength   int64
}

type IssueUploadAssetParam struct {
	ExpiresAt time.Time

	UUID            string
	Filename        string
	ContentType     string
	ContentEncoding string

	Cursor string
	// ContentLength is the size of the file in bytes. It is required when S3 is used.
	ContentLength int64
}

func (p IssueUploadAssetParam) GetOrGuessContentType() string {
	if p.ContentType != "" {
		return p.ContentType
	}
	return mime.TypeByExtension(path.Ext(p.Filename))
}

type File interface {
	ReadAsset(
		context.Context,
		string,
		string,
		map[string]string,
	) (io.ReadCloser, map[string]string, error)
	GetAssetFiles(context.Context, string) ([]FileEntry, error)
	UploadAsset(context.Context, *file.File) (string, int64, error)
	Read(context.Context, string, map[string]string) (io.ReadCloser, map[string]string, error)
	Upload(context.Context, *file.File, string) (int64, error)
	DeleteAsset(context.Context, string, string) error
	DeleteAssets(context.Context, []string) error
	PublishAsset(context.Context, string, string) error
	UnpublishAsset(context.Context, string, string) error
	GetAccessInfoResolver() asset.AccessInfoResolver
	GetAccessInfo(*asset.Asset) *asset.AccessInfo
	GetBaseURL() string
	IssueUploadAssetLink(context.Context, IssueUploadAssetParam) (*UploadAssetLink, error)
	UploadedAsset(context.Context, *asset.Upload) (*file.File, error)

	RemoveAsset(context.Context, *url.URL) error // viz
}

func init() {
	// mime package depends on the OS, so adding the requited mime types to make sure about the results in different OS
	lo.Must0(mime.AddExtensionType(".zip", "application/zip"))
	lo.Must0(mime.AddExtensionType(".7z", "application/x-7z-compressed"))
	lo.Must0(mime.AddExtensionType(".gz", "application/gzip"))
	lo.Must0(mime.AddExtensionType(".bz2", "application/x-bzip2"))
	lo.Must0(mime.AddExtensionType(".tar", "application/x-tar"))
	lo.Must0(mime.AddExtensionType(".rar", "application/vnd.rar"))
}
