package graphql

import (
	"github.com/reearth/reearthx/asset/repository"
	assetusecase "github.com/reearth/reearthx/asset/usecase"
)

type Resolver struct {
	assetUsecase assetusecase.Usecase
	pubsub       repository.PubSubRepository
}

func NewResolver(assetUsecase assetusecase.Usecase, pubsub repository.PubSubRepository) *Resolver {
	return &Resolver{
		assetUsecase: assetUsecase,
		pubsub:       pubsub,
	}
}
