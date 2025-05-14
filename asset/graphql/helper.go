package graphql

import (
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/reearth/reearthx/asset/domain/entity"
)

func FileFromUpload(file *graphql.Upload) io.Reader {
	return file.File
}

func AssetFromDomain(a *entity.Asset) *Asset {
	if a == nil {
		return nil
	}

	var err *string
	if e := a.Error(); e != "" {
		err = &e
	}

	var url *string
	if u := a.URL(); u != "" {
		url = &u
	}

	return &Asset{
		ID:          a.ID().String(),
		Name:        a.Name(),
		Size:        int(a.Size()),
		ContentType: a.ContentType(),
		URL:         url,
		Status:      AssetStatus(a.Status()),
		Error:       err,
		CreatedAt:   a.CreatedAt(),
		UpdatedAt:   a.UpdatedAt(),
	}
}
