package asset

import (
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/samber/lo"
)

type Map map[id.ID]*Asset

func (m Map) List() List {
	return lo.MapToSlice(m, func(_ id.ID, v *Asset) *Asset {
		return v
	})
}

func (m Map) ListFrom(ids []id.ID) (res List) {
	for _, id := range ids {
		if a, ok := m[id]; ok {
			res = append(res, a)
		}
	}
	return
}
