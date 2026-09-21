package interactor

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/labstack/gommon/log"
	"github.com/reearth/reearthx/asset/domain/integrationapi"
	"github.com/reearth/reearthx/asset/domain/item"
	"github.com/reearth/reearthx/asset/domain/schema"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
)

var pointFieldIsNotSupportedError = rerror.NewE(
	i18n.T("point type is not supported in any geometry field in this model"),
)

// GeoJSON
func featureCollectionFromItems(
	ver item.VersionedList,
	s *schema.Schema,
) (*integrationapi.FeatureCollection, error) {
	return integrationapi.FeatureCollectionFromItems(ver, s)
}

// CSV
//
// ctx must be the request context, not the transaction one: it is what stops
// the generator goroutine when the client walks away mid-download.
func csvFromItems(
	ctx context.Context,
	pw *io.PipeWriter,
	l item.VersionedList,
	s *schema.Schema,
) error {
	if !s.IsPointFieldSupported() {
		return pointFieldIsNotSupportedError
	}
	go handleCSVGeneration(ctx, pw, l, s)
	return nil
}

func handleCSVGeneration(
	ctx context.Context,
	pw *io.PipeWriter,
	l item.VersionedList,
	s *schema.Schema,
) {
	done := make(chan struct{})
	defer close(done)

	// a pipe write blocks until someone reads it, so the generator would park
	// forever if the reader is abandoned; closing the pipe unblocks it
	go func() {
		select {
		case <-ctx.Done():
			_ = pw.CloseWithError(ctx.Err())
		case <-done:
		}
	}()

	err := generateCSV(pw, l, s)
	if err != nil {
		log.Errorf("failed to generate CSV: %v", err)
		_ = pw.CloseWithError(err)
	} else {
		_ = pw.Close()
	}
}

func generateCSV(pw *io.PipeWriter, l item.VersionedList, s *schema.Schema) error {
	w := csv.NewWriter(pw)
	defer w.Flush()
	headers := integrationapi.BuildCSVHeaders(s)
	if err := w.Write(headers); err != nil {
		return err
	}
	nonGeoFields := lo.Filter(s.Fields(), func(f *schema.Field, _ int) bool {
		return !f.IsGeometryField()
	})
	for _, ver := range l {
		row, ok := integrationapi.RowFromItem(ver.Value(), nonGeoFields)
		if ok {
			if err := w.Write(row); err != nil {
				return err
			}
		}
	}

	return w.Error()
}
