package asset

import (
	asset2 "github.com/reearth/reearth-cms/server/pkg/asset"
	"time"
)

type Upload struct {
	uuid            string
	project         asset2.ProjectID
	fileName        string
	expiresAt       time.Time
	contentLength   int64
	contentType     string
	contentEncoding string
}

func (u *Upload) UUID() string {
	return u.uuid
}

func (u *Upload) Project() asset2.ProjectID {
	return u.project
}

func (u *Upload) FileName() string {
	return u.fileName
}

func (u *Upload) ExpiresAt() time.Time {
	return u.expiresAt
}

func (u *Upload) Expired(t time.Time) bool {
	return t.After(u.expiresAt)
}

func (u *Upload) ContentLength() int64 {
	return u.contentLength
}

func (u *Upload) ContentType() string {
	return u.contentType
}

func (u *Upload) ContentEncoding() string {
	return u.contentEncoding
}
