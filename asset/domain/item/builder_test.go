package item

import (
	"testing"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/project"
	"github.com/reearth/reearthx/asset/domain/schema"
	"github.com/reearth/reearthx/asset/domain/value"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/util"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_ID(t *testing.T) {
	iid := NewID()
	b := New().ID(iid).
		Schema(id.NewSchemaID()).
		Model(id.NewModelID()).
		Project(id.NewProjectID()).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	assert.Equal(t, iid, b.id)
}

func TestBuilder_SchemaID(t *testing.T) {
	sid := schema.NewID()
	b := New().NewID().
		Schema(sid).
		Model(id.NewModelID()).
		Project(id.NewProjectID()).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	assert.Equal(t, sid, b.Schema())
}

func TestBuilder_Fields(t *testing.T) {
	fid := schema.NewFieldID()
	fields := Fields{NewField(fid, value.TypeBool.Value(true).AsMultiple(), nil)}
	b := New().NewID().
		Schema(id.NewSchemaID()).
		Model(id.NewModelID()).
		Project(id.NewProjectID()).
		Fields(fields).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	assert.Equal(t, fields, b.Fields())
	b = New().NewID().
		Schema(id.NewSchemaID()).
		Project(id.NewProjectID()).
		Model(id.NewModelID()).
		Fields(nil).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	assert.Nil(t, b.Fields())
}

func TestNew(t *testing.T) {
	res := New()
	assert.NotNil(t, res)
}

func TestBuilder_NewID(t *testing.T) {
	res, _ := New().NewID().
		Schema(id.NewSchemaID()).
		Model(id.NewModelID()).
		Project(id.NewProjectID()).
		Thread(id.NewThreadID().Ref()).
		Build()
	assert.NotNil(t, res.ID())
}

func TestBuilder_Project(t *testing.T) {
	pid := project.NewID()
	b := New().NewID().
		Project(pid).
		Model(id.NewModelID()).
		Schema(id.NewSchemaID()).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	assert.Equal(t, pid, b.Project())
}

func TestBuilder_Model(t *testing.T) {
	mid := id.NewModelID()
	b := New().NewID().
		Model(mid).
		Project(id.NewProjectID()).
		Schema(id.NewSchemaID()).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	assert.Equal(t, mid, b.Model())
}

func TestBuilder_User(t *testing.T) {
	uId := accountdomain.NewUserID()
	b := New().User(uId)
	assert.Equal(t, &uId, b.i.user)
}

func TestBuilder_Integration(t *testing.T) {
	iId := id.NewIntegrationID()
	b := New().Integration(iId)
	assert.Equal(t, &iId, b.i.integration)
}

func TestBuilder_Thread(t *testing.T) {
	tid := id.NewThreadID().Ref()
	b := New().NewID().
		Model(id.NewModelID()).
		Project(id.NewProjectID()).
		Schema(id.NewSchemaID()).
		Thread(tid).
		MustBuild()
	assert.Equal(t, tid, b.Thread())
}

func TestBuilder_Timestamp(t *testing.T) {
	tt := time.Now()
	b := New().NewID().
		Project(id.NewProjectID()).
		Schema(id.NewSchemaID()).
		Model(id.NewModelID()).
		Timestamp(tt).
		Schema(id.NewSchemaID()).
		Thread(id.NewThreadID().Ref()).
		MustBuild()
	assert.Equal(t, tt, b.Timestamp())
}

func TestBuilder_Build(t *testing.T) {
	iid := NewID()
	sid := id.NewSchemaID()
	mid := id.NewModelID()
	pid := id.NewProjectID()
	tid := id.NewThreadID().Ref()
	now := time.Now()
	defer util.MockNow(now)()

	type fields struct {
		i *Item
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Item
		wantErr error
	}{
		{
			name: "should build an item",
			fields: fields{
				i: &Item{
					id:      iid,
					schema:  sid,
					project: pid,
					model:   mid,
					thread:  tid,
				},
			},
			want: &Item{
				id:         iid,
				schema:     sid,
				project:    pid,
				model:      mid,
				thread:     tid,
				timestamp:  now,
				isMetadata: false,
			},
			wantErr: nil,
		},
		{
			name: "should fail: invalid item ID",
			fields: fields{
				i: &Item{},
			},
			want:    nil,
			wantErr: id.ErrInvalidID,
		},
		{
			name: "should fail: invalid schema ID",
			fields: fields{
				i: &Item{
					id:      iid,
					project: pid,
					model:   mid,
					thread:  tid,
				},
			},
			want:    nil,
			wantErr: id.ErrInvalidID,
		},
		{
			name: "should fail: invalid project ID",
			fields: fields{
				i: &Item{
					id:     iid,
					schema: sid,
					model:  mid,
					thread: tid,
				},
			},
			want:    nil,
			wantErr: id.ErrInvalidID,
		},
		{
			name: "should fail: invalid model ID",
			fields: fields{
				i: &Item{
					id:      iid,
					schema:  sid,
					project: pid,
					thread:  tid,
				},
			},
			want:    nil,
			wantErr: id.ErrInvalidID,
		},
		// {
		// 	name: "should fail: invalid thread ID",
		// 	fields: fields{
		// 		i: &Item{
		// 			id:      iid,
		// 			schema:  sid,
		// 			project: pid,
		// 			model:   mid,
		// 		},
		// 	},
		// 	want:    nil,
		// 	wantErr: id.ErrInvalidID,
		// },
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				i: tt.fields.i,
			}

			got, err := b.Build()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				assert.Panics(t, func() {
					_ = b.MustBuild()
				})
			} else {
				assert.Equal(t, tt.want, got)
				got = b.MustBuild()
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuilder_NewThread(t *testing.T) {
	b := New().NewThread()
	assert.NotNil(t, b.i.thread)
}

func TestBuilder_MetadataItem(t *testing.T) {
	iId := id.NewItemID().Ref()
	b := New().MetadataItem(iId)
	assert.Equal(t, iId, b.i.metadataItem)
}

func TestBuilder_IsMetadata(t *testing.T) {
	b := New().IsMetadata(true)
	assert.Equal(t, true, b.i.isMetadata)
}

func TestBuilder_UpdatedByUser(t *testing.T) {
	uId := accountdomain.NewUserID()
	uuid := New().UpdatedByUser(&uId)
	assert.Equal(t, &uId, uuid.i.updatedByUser)
}

func TestBuilder_UpdatedByIntegration(t *testing.T) {
	iid := id.NewIntegrationID()
	uuid := New().UpdatedByIntegration(&iid)
	assert.Equal(t, &iid, uuid.i.updatedByIntegration)
}

func TestBuilder_OriginalItem(t *testing.T) {
	iId := id.NewItemID().Ref()
	b := New().OriginalItem(iId)
	assert.Equal(t, iId, b.i.originalItem)
}
