package mongodoc

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/model"
	"github.com/reearth/reearthx/asset/domain/project"
	"github.com/reearth/reearthx/asset/domain/schema"
	"github.com/stretchr/testify/assert"
)

func TestModelDocument_Model(t *testing.T) {
	now := time.Now()
	mId, pId, sId, smId := model.NewID(), project.NewID(), schema.NewID(), schema.NewID()
	tests := []struct {
		name    string
		mDoc    *ModelDocument
		want    *model.Model
		wantErr bool
	}{
		{
			name: "model",
			mDoc: &ModelDocument{
				ID:          mId.String(),
				Name:        "abc",
				Description: "xyz",
				Key:         "mmm123",
				Public:      true,
				Metadata:    smId.StringRef(),
				Project:     pId.String(),
				Schema:      sId.String(),
				UpdatedAt:   now,
			},
			want: model.New().ID(mId).
				Name("abc").
				Description("xyz").
				Key(id.NewKey("mmm123")).
				Public(true).
				Project(pId).
				Metadata(smId.Ref()).
				Schema(sId).
				UpdatedAt(now).
				MustBuild(),
			wantErr: false,
		},
		{
			name: "Invalid id 1",
			mDoc: &ModelDocument{
				ID:          "abc",
				Name:        "abc",
				Description: "xyz",
				Key:         "mmm123",
				Public:      true,
				Project:     pId.String(),
				Schema:      sId.String(),
				UpdatedAt:   now,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid id 2",
			mDoc: &ModelDocument{
				ID:          mId.String(),
				Name:        "abc",
				Description: "xyz",
				Key:         "mmm123",
				Public:      true,
				Project:     "abc",
				Schema:      sId.String(),
				UpdatedAt:   now,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid id 3",
			mDoc: &ModelDocument{
				ID:          mId.String(),
				Name:        "abc",
				Description: "xyz",
				Key:         "mmm123",
				Public:      true,
				Project:     pId.String(),
				Schema:      "abc",
				UpdatedAt:   now,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mDoc.Model()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewModel(t *testing.T) {
	now := time.Now()
	mId, pId, sId := model.NewID(), project.NewID(), schema.NewID()
	tests := []struct {
		name   string
		args   *model.Model
		want   *ModelDocument
		wantId string
	}{
		{
			name: "",
			args: model.New().ID(mId).
				Name("abc").
				Description("xyz").
				Key(id.NewKey("mmm123")).
				Public(true).
				Project(pId).
				Schema(sId).
				UpdatedAt(now).
				MustBuild(),
			want: &ModelDocument{
				ID:          mId.String(),
				Name:        "abc",
				Description: "xyz",
				Key:         "mmm123",
				Public:      true,
				Project:     pId.String(),
				Schema:      sId.String(),
				UpdatedAt:   now,
			},
			wantId: mId.String(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, gotId := NewModel(tt.args)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantId, gotId)
		})
	}
}

func TestNewModelConsumer(t *testing.T) {
	c := NewModelConsumer()
	assert.NotNil(t, c)
}
