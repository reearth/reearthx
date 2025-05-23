package asset

import "github.com/samber/lo"

type Map map[AssetID]*Asset

func (m Map) List() List {
	return lo.MapToSlice(m, func(_ AssetID, v *Asset) *Asset {
		return v
	})
}

func (m Map) ListFrom(ids AssetIDList) (res List) {
	for _, id := range ids {
		if a, ok := m[id]; ok {
			res = append(res, a)
		}
	}
	return
}
