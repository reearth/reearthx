package asset

import (
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/samber/lo"
)

type Map map[id.AssetID]*Asset

func (m Map) List() List {
	return lo.MapToSlice(m, func(_ id.AssetID, v *Asset) *Asset {
		return v
	})
}

func (m Map) ListFrom(ids []id.AssetID) (res List) {
	for _, id := range ids {
		if a, ok := m[id]; ok {
			res = append(res, a)
		}
	}
	return
}
