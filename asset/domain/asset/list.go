package asset

import (
	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

type List []*Asset

func (l List) SortByID() List {
	m := slices.Clone(l)
	slices.SortFunc(m, func(a, b *Asset) int {
		return a.ID().Compare(b.ID())
	})
	return m
}

func (l List) SetAccessInfoResolver(r AccessInfoResolver) {
	lo.ForEach(l, func(a *Asset, _ int) {
		if a != nil {
			a.SetAccessInfoResolver(r)
		}
	})
}

func (l List) Clone() List {
	return util.Map(l, func(p *Asset) *Asset { return p.Clone() })
}

func (l List) Map() Map {
	return lo.SliceToMap(lo.Filter(l, func(a *Asset, _ int) bool {
		return a != nil
	}), func(a *Asset) (AssetID, *Asset) {
		return a.ID(), a
	})
}

func (l List) IDs() []AssetID {
	ids := make([]AssetID, 0, len(l))
	for _, a := range l {
		ids = append(ids, a.ID())
	}
	return ids
}
