package graphql

import (
	"github.com/reearth/reearthx/asset/service"
)

type Resolver struct {
	assetService *service.Service
}

func NewResolver(assetService *service.Service) *Resolver {
	return &Resolver{
		assetService: assetService,
	}
}
