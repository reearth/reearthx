package graphql

import (
	"github.com/reearth/reearthx/asset/repository"
	"github.com/reearth/reearthx/asset/service"
)

type Resolver struct {
	assetService *service.Service
	pubsub       repository.PubSubRepository
}

func NewResolver(assetService *service.Service, pubsub repository.PubSubRepository) *Resolver {
	return &Resolver{
		assetService: assetService,
		pubsub:       pubsub,
	}
}
