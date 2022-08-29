package authserver

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"
	"github.com/stretchr/testify/assert"
)

func TestNewMemory(t *testing.T) {
	assert.Equal(t, &Memory{
		data: &util.SyncMap[RequestID, *Request]{},
	}, NewMemory())
}

func TestMemory_FindByID(t *testing.T) {
	ctx := context.Background()
	r := NewRequest().NewID().MustBuild()

	got, err := (&Memory{
		data: &util.SyncMap[RequestID, *Request]{},
	}).FindByID(ctx, r.ID())
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	got, err = (&Memory{
		data: util.SyncMapFrom(map[RequestID]*Request{
			r.ID(): r,
		}),
	}).FindByID(ctx, r.ID())
	assert.Same(t, r, got)
	assert.NoError(t, err)
}

func TestMemory_FindByCode(t *testing.T) {
	ctx := context.Background()
	r := NewRequest().NewID().Code("aaa").MustBuild()

	got, err := (&Memory{
		data: &util.SyncMap[RequestID, *Request]{},
	}).FindByCode(ctx, "aaa")
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	got, err = (&Memory{
		data: util.SyncMapFrom(map[RequestID]*Request{
			r.ID(): r,
		}),
	}).FindByCode(ctx, "aaa")
	assert.Same(t, r, got)
	assert.NoError(t, err)
}

func TestMemory_FindBySubject(t *testing.T) {
	ctx := context.Background()
	r := NewRequest().NewID().Subject("sss").MustBuild()

	got, err := (&Memory{
		data: &util.SyncMap[RequestID, *Request]{},
	}).FindBySubject(ctx, "sss")
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	got, err = (&Memory{
		data: util.SyncMapFrom(map[RequestID]*Request{
			r.ID(): r,
		}),
	}).FindBySubject(ctx, "sss")
	assert.Same(t, r, got)
	assert.NoError(t, err)
}

func TestMemory_Save(t *testing.T) {
	ctx := context.Background()
	r := NewRequest().NewID().MustBuild()

	m := &Memory{
		data: &util.SyncMap[RequestID, *Request]{},
	}
	assert.NoError(t, m.Save(ctx, r))
	_, ok := m.data.Load(r.ID())
	assert.True(t, ok)
}

func TestMemory_Remove(t *testing.T) {
	ctx := context.Background()
	r := NewRequest().NewID().MustBuild()

	err := (&Memory{
		data: &util.SyncMap[RequestID, *Request]{},
	}).Remove(ctx, r.ID())
	assert.NoError(t, err)

	m := &Memory{
		data: util.SyncMapFrom(map[RequestID]*Request{
			r.ID(): r,
		}),
	}
	err = m.Remove(ctx, r.ID())
	assert.NoError(t, err)
	_, ok := m.data.Load(r.ID())
	assert.False(t, ok)
}
