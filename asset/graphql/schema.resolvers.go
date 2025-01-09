package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.43

import (
	"context"

	"github.com/reearth/reearthx/asset/domain"
)

// UploadAsset is the resolver for the uploadAsset field.
func (r *mutationResolver) UploadAsset(ctx context.Context, input UploadAssetInput) (*UploadAssetPayload, error) {
	id, err := domain.IDFrom(input.ID)
	if err != nil {
		return nil, err
	}

	// Create asset metadata
	asset := domain.NewAsset(
		id,
		input.File.Filename,
		input.File.Size,
		input.File.ContentType,
	)

	// Create asset metadata first
	if err := r.assetUsecase.CreateAsset(ctx, asset); err != nil {
		return nil, err
	}

	// Upload file content
	if err := r.assetUsecase.UploadAssetContent(ctx, id, FileFromUpload(&input.File)); err != nil {
		return nil, err
	}

	// Update asset status to active
	asset.UpdateStatus(domain.StatusActive, "")
	if err := r.assetUsecase.UpdateAsset(ctx, asset); err != nil {
		return nil, err
	}

	return &UploadAssetPayload{
		Asset: AssetFromDomain(asset),
	}, nil
}

// GetAssetUploadURL is the resolver for the getAssetUploadURL field.
func (r *mutationResolver) GetAssetUploadURL(ctx context.Context, input GetAssetUploadURLInput) (*GetAssetUploadURLPayload, error) {
	id, err := domain.IDFrom(input.ID)
	if err != nil {
		return nil, err
	}

	url, err := r.assetUsecase.GetAssetUploadURL(ctx, id)
	if err != nil {
		return nil, err
	}

	return &GetAssetUploadURLPayload{
		UploadURL: url,
	}, nil
}

// UpdateAssetMetadata is the resolver for the updateAssetMetadata field.
func (r *mutationResolver) UpdateAssetMetadata(ctx context.Context, input UpdateAssetMetadataInput) (*UpdateAssetMetadataPayload, error) {
	id, err := domain.IDFrom(input.ID)
	if err != nil {
		return nil, err
	}

	asset, err := r.assetUsecase.GetAsset(ctx, id)
	if err != nil {
		return nil, err
	}

	asset.UpdateMetadata(input.Name, "", input.ContentType)
	asset.SetSize(int64(input.Size))
	if err := r.assetUsecase.UpdateAsset(ctx, asset); err != nil {
		return nil, err
	}

	return &UpdateAssetMetadataPayload{
		Asset: AssetFromDomain(asset),
	}, nil
}

// DeleteAsset is the resolver for the deleteAsset field.
func (r *mutationResolver) DeleteAsset(ctx context.Context, input DeleteAssetInput) (*DeleteAssetPayload, error) {
	id, err := domain.IDFrom(input.ID)
	if err != nil {
		return nil, err
	}

	if err := r.assetUsecase.DeleteAsset(ctx, id); err != nil {
		return nil, err
	}

	return &DeleteAssetPayload{
		AssetID: input.ID,
	}, nil
}

// DeleteAssets is the resolver for the deleteAssets field.
func (r *mutationResolver) DeleteAssets(ctx context.Context, input DeleteAssetsInput) (*DeleteAssetsPayload, error) {
	var assetIDs []domain.ID
	for _, idStr := range input.Ids {
		id, err := domain.IDFrom(idStr)
		if err != nil {
			return nil, err
		}
		assetIDs = append(assetIDs, id)
	}

	for _, id := range assetIDs {
		if err := r.assetUsecase.DeleteAsset(ctx, id); err != nil {
			return nil, err
		}
	}

	return &DeleteAssetsPayload{
		AssetIds: input.Ids,
	}, nil
}

// DeleteAssetsInGroup is the resolver for the deleteAssetsInGroup field.
func (r *mutationResolver) DeleteAssetsInGroup(ctx context.Context, input DeleteAssetsInGroupInput) (*DeleteAssetsInGroupPayload, error) {
	groupID, err := domain.GroupIDFrom(input.GroupID)
	if err != nil {
		return nil, err
	}

	if err := r.assetUsecase.DeleteAllAssetsInGroup(ctx, groupID); err != nil {
		return nil, err
	}

	return &DeleteAssetsInGroupPayload{
		GroupID: input.GroupID,
	}, nil
}

// MoveAsset is the resolver for the moveAsset field.
func (r *mutationResolver) MoveAsset(ctx context.Context, input MoveAssetInput) (*MoveAssetPayload, error) {
	id, err := domain.IDFrom(input.ID)
	if err != nil {
		return nil, err
	}

	asset, err := r.assetUsecase.GetAsset(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.ToWorkspaceID != nil {
		wsID, err := domain.WorkspaceIDFrom(*input.ToWorkspaceID)
		if err != nil {
			return nil, err
		}
		asset.MoveToWorkspace(wsID)
	}

	if input.ToProjectID != nil {
		projID, err := domain.ProjectIDFrom(*input.ToProjectID)
		if err != nil {
			return nil, err
		}
		asset.MoveToProject(projID)
	}

	if err := r.assetUsecase.UpdateAsset(ctx, asset); err != nil {
		return nil, err
	}

	return &MoveAssetPayload{
		Asset: AssetFromDomain(asset),
	}, nil
}

// Asset is the resolver for the asset field.
func (r *queryResolver) Asset(ctx context.Context, id string) (*Asset, error) {
	assetID, err := domain.IDFrom(id)
	if err != nil {
		return nil, err
	}

	asset, err := r.assetUsecase.GetAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}

	return AssetFromDomain(asset), nil
}

// Assets is the resolver for the assets field.
func (r *queryResolver) Assets(ctx context.Context) ([]*Asset, error) {
	assets, err := r.assetUsecase.ListAssets(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*Asset, len(assets))
	for i, asset := range assets {
		result[i] = AssetFromDomain(asset)
	}

	return result, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
