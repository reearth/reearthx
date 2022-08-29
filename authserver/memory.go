package authserver

import (
	"context"

	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/util"
)

type Memory struct {
	data *util.SyncMap[RequestID, *Request]
}

var _ RequestRepo = (*Memory)(nil)

func NewMemory() *Memory {
	return &Memory{
		data: &util.SyncMap[RequestID, *Request]{},
	}
}

func (r *Memory) FindByID(_ context.Context, id RequestID) (*Request, error) {
	d, ok := r.data.Load(id)
	if ok {
		return d, nil
	}
	return nil, rerror.ErrNotFound
}

func (r *Memory) FindByCode(_ context.Context, s string) (*Request, error) {
	a := r.data.Find(func(_ RequestID, ar *Request) bool {
		return ar.GetCode() == s
	})
	if a == nil {
		return nil, rerror.ErrNotFound
	}
	return a, nil
}

func (r *Memory) FindBySubject(_ context.Context, s string) (*Request, error) {
	a := r.data.Find(func(_ RequestID, ar *Request) bool {
		return ar.GetSubject() == s
	})
	if a == nil {
		return nil, rerror.ErrNotFound
	}
	return a, nil
}

func (r *Memory) Save(_ context.Context, request *Request) error {
	r.data.Store(request.ID(), request)
	return nil
}

func (r *Memory) Remove(_ context.Context, requestID RequestID) error {
	r.data.Delete(requestID)
	return nil
}
