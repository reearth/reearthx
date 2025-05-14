package asset

import (
	"fmt"
)

type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeGCS   StorageType = "gcs"
)

type Config struct {
	StorageType    StorageType
	StorageBaseDir string
	StorageBaseURL string
	StorageBucket  string
	GCSCredentials string
}

type Factory struct {
	assetService  AssetService
	groupService  GroupService
	policyService PolicyService
	storage       Storage
}

func NewFactory(
	assetRepo AssetRepository,
	groupRepo GroupRepository,
	policyRepo PolicyRepository,
	storage Storage,
) (*Factory, error) {
	if assetRepo == nil {
		return nil, fmt.Errorf("asset repository is required")
	}
	if groupRepo == nil {
		return nil, fmt.Errorf("group repository is required")
	}
	if policyRepo == nil {
		return nil, fmt.Errorf("policy repository is required")
	}
	if storage == nil {
		return nil, fmt.Errorf("storage is required")
	}

	fileProcessor := NewFileProcessor()
	zipExtractor := NewZipExtractor(assetRepo, storage)

	assetService := NewAssetService(assetRepo, groupRepo, storage, fileProcessor, zipExtractor)
	groupService := NewGroupService(groupRepo)
	policyService := NewPolicyService(policyRepo)

	return &Factory{
		assetService:  assetService,
		groupService:  groupService,
		policyService: policyService,
		storage:       storage,
	}, nil
}

func (f *Factory) AssetService() AssetService {
	return f.assetService
}

func (f *Factory) GroupService() GroupService {
	return f.groupService
}

func (f *Factory) PolicyService() PolicyService {
	return f.policyService
}

func (f *Factory) Storage() Storage {
	return f.storage
}
