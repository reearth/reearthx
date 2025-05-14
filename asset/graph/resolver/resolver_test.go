package resolver

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/reearth/reearthx/asset"
	"github.com/reearth/reearthx/asset/graph/generated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAssetService struct {
	mock.Mock
}

var _ asset.AssetService = &MockAssetService{}

func (m *MockAssetService) CreateAsset(ctx context.Context, param asset.CreateAssetParam) (*asset.Asset, error) {
	args := m.Called(ctx, param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetService) GetAsset(ctx context.Context, id asset.AssetID) (*asset.Asset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetService) GetAssetFile(ctx context.Context, id asset.AssetID) (*asset.AssetFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.AssetFile), args.Error(1)
}

func (m *MockAssetService) ListAssets(ctx context.Context, groupID asset.GroupID, filter asset.AssetFilter, sort asset.AssetSort, pagination asset.Pagination) ([]*asset.Asset, int64, error) {
	args := m.Called(ctx, groupID, filter, sort, pagination)
	return args.Get(0).([]*asset.Asset), args.Get(1).(int64), args.Error(2)
}

func (m *MockAssetService) UpdateAsset(ctx context.Context, param asset.UpdateAssetParam) (*asset.Asset, error) {
	args := m.Called(ctx, param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetService) DeleteAsset(ctx context.Context, id asset.AssetID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAssetService) DeleteAssets(ctx context.Context, ids []asset.AssetID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockAssetService) DecompressAsset(ctx context.Context, id asset.AssetID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAssetService) CreateAssetUpload(ctx context.Context, param asset.CreateAssetUploadParam) (*asset.AssetUploadInfo, error) {
	args := m.Called(ctx, param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.AssetUploadInfo), args.Error(1)
}

type MockGroupService struct {
	mock.Mock
}

func (m *MockGroupService) CreateGroup(ctx context.Context, param struct {
	Name string
}) (*asset.Group, error) {
	args := m.Called(ctx, param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Group), args.Error(1)
}

func (m *MockGroupService) GetGroup(ctx context.Context, id asset.GroupID) (*asset.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Group), args.Error(1)
}

func (m *MockGroupService) DeleteGroup(ctx context.Context, id asset.GroupID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGroupService) AssignPolicy(ctx context.Context, groupID asset.GroupID, policyID *asset.PolicyID) (*asset.Group, error) {
	args := m.Called(ctx, groupID, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Group), args.Error(1)
}

type MockPolicyService struct {
	mock.Mock
}

func (m *MockPolicyService) CreatePolicy(ctx context.Context, param struct {
	Name         string
	StorageLimit int64
}) (*asset.Policy, error) {
	args := m.Called(ctx, param)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Policy), args.Error(1)
}

func (m *MockPolicyService) GetPolicy(ctx context.Context, id asset.PolicyID) (*asset.Policy, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.Policy), args.Error(1)
}

func (m *MockPolicyService) DeletePolicy(ctx context.Context, id asset.PolicyID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupClient(t *testing.T) (*client.Client, *MockAssetService, *MockGroupService, *MockPolicyService) {
	t.Helper()
	mockAssetService := new(MockAssetService)
	mockGroupService := new(MockGroupService)
	mockPolicyService := new(MockPolicyService)

	resolverRoot := &Resolver{
		AssetService:  mockAssetService,
		GroupService:  mockGroupService,
		PolicyService: mockPolicyService,
	}

	gqlClient := client.New(handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolverRoot,
	})))

	return gqlClient, mockAssetService, mockGroupService, mockPolicyService
}

func TestQueryAsset(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	assetID := asset.NewAssetID()
	groupID := asset.NewGroupID()
	now := time.Now()
	mockAsset := &asset.Asset{
		ID:          assetID,
		GroupID:     groupID,
		CreatedAt:   now,
		Size:        1024,
		ContentType: "image/jpeg",
		UUID:        "test-uuid",
		URL:         "http://example.com/test.jpg",
		FileName:    "test.jpg",
	}

	mockAssetService.On("GetAsset", mock.Anything, assetID).Return(mockAsset, nil)

	var resp struct {
		Asset struct {
			ID        string
			GroupID   string
			CreatedAt string
			Size      int
			FileName  string
			URL       string
		}
	}

	query := `
		query GetAsset($id: ID!) {
			asset(id: $id) {
				id
				groupId
				createdAt
				size
				fileName
				url
			}
		}
	`

	c.MustPost(query, &resp, client.Var("id", assetID.String()))

	assert.Equal(t, assetID.String(), resp.Asset.ID)
	assert.Equal(t, groupID.String(), resp.Asset.GroupID)
	assert.Equal(t, 1024, resp.Asset.Size)
	assert.Equal(t, "test.jpg", resp.Asset.FileName)
	assert.Equal(t, "http://example.com/test.jpg", resp.Asset.URL)

	mockAssetService.AssertExpectations(t)
}

func TestCreateAsset(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	assetID := asset.NewAssetID()
	groupID := asset.NewGroupID()
	now := time.Now()

	mockAsset := &asset.Asset{
		ID:          assetID,
		GroupID:     groupID,
		CreatedAt:   now,
		Size:        1024,
		ContentType: "image/jpeg",
		UUID:        "test-uuid",
		URL:         "http://example.com/test.jpg",
		FileName:    "test.jpg",
	}

	mockAssetService.On("CreateAsset", mock.Anything, mock.MatchedBy(func(param asset.CreateAssetParam) bool {
		return param.GroupID == groupID && param.URL == "http://example.com/test.jpg" && param.Token == "test-token"
	})).Return(mockAsset, nil)

	var resp struct {
		CreateAsset struct {
			Asset struct {
				ID       string
				GroupID  string
				Size     int
				FileName string
				URL      string
			}
		}
	}

	mutation := `
		mutation CreateAsset($input: CreateAssetInput!) {
			createAsset(input: $input) {
				asset {
					id
					groupId
					size
					fileName
					url
				}
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"groupId": groupID.String(),
		"url":     "http://example.com/test.jpg",
		"token":   "test-token",
	}))

	assert.Equal(t, assetID.String(), resp.CreateAsset.Asset.ID)
	assert.Equal(t, groupID.String(), resp.CreateAsset.Asset.GroupID)
	assert.Equal(t, 1024, resp.CreateAsset.Asset.Size)
	assert.Equal(t, "test.jpg", resp.CreateAsset.Asset.FileName)
	assert.Equal(t, "http://example.com/test.jpg", resp.CreateAsset.Asset.URL)

	mockAssetService.AssertExpectations(t)
}

func TestListAssets(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	groupID := asset.NewGroupID()
	assets := []*asset.Asset{
		{
			ID:          asset.NewAssetID(),
			GroupID:     groupID,
			CreatedAt:   time.Now(),
			Size:        1024,
			ContentType: "image/jpeg",
			UUID:        "test-uuid-1",
			URL:         "http://example.com/test1.jpg",
			FileName:    "test1.jpg",
		},
		{
			ID:          asset.NewAssetID(),
			GroupID:     groupID,
			CreatedAt:   time.Now(),
			Size:        2048,
			ContentType: "image/png",
			UUID:        "test-uuid-2",
			URL:         "http://example.com/test2.png",
			FileName:    "test2.png",
		},
	}

	mockAssetService.On("ListAssets", mock.Anything, groupID, mock.Anything, mock.Anything, mock.Anything).Return(assets, int64(2), nil)

	var resp struct {
		Assets struct {
			Nodes []struct {
				ID       string
				FileName string
				Size     int
			}
			TotalCount int
		}
	}

	query := `
		query ListAssets($groupId: ID!, $pagination: Pagination!) {
			assets(groupId: $groupId, pagination: $pagination) {
				nodes {
					id
					fileName
					size
				}
				totalCount
			}
		}
	`

	c.MustPost(query, &resp,
		client.Var("groupId", groupID.String()),
		client.Var("pagination", map[string]interface{}{
			"offset": 0,
			"limit":  10,
		}),
	)

	assert.Equal(t, 2, len(resp.Assets.Nodes))
	assert.Equal(t, 2, resp.Assets.TotalCount)
	assert.Equal(t, "test1.jpg", resp.Assets.Nodes[0].FileName)
	assert.Equal(t, 1024, resp.Assets.Nodes[0].Size)
	assert.Equal(t, "test2.png", resp.Assets.Nodes[1].FileName)
	assert.Equal(t, 2048, resp.Assets.Nodes[1].Size)

	mockAssetService.AssertExpectations(t)
}

func TestUpdateAsset(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	assetID := asset.NewAssetID()
	groupID := asset.NewGroupID()
	now := time.Now()
	previewType := asset.PreviewTypeImage

	mockAsset := &asset.Asset{
		ID:          assetID,
		GroupID:     groupID,
		CreatedAt:   now,
		Size:        1024,
		ContentType: "image/jpeg",
		UUID:        "test-uuid",
		URL:         "http://example.com/test.jpg",
		FileName:    "test.jpg",
		PreviewType: previewType,
	}

	mockAssetService.On("UpdateAsset", mock.Anything, mock.MatchedBy(func(param asset.UpdateAssetParam) bool {
		return param.ID == assetID && *param.PreviewType == previewType
	})).Return(mockAsset, nil)

	var resp struct {
		UpdateAsset struct {
			Asset struct {
				ID          string
				GroupID     string
				PreviewType string
				FileName    string
			}
		}
	}

	mutation := `
		mutation UpdateAsset($input: UpdateAssetInput!) {
			updateAsset(input: $input) {
				asset {
					id
					groupId
					previewType
					fileName
				}
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"id":          assetID.String(),
		"previewType": "IMAGE",
	}))

	assert.Equal(t, assetID.String(), resp.UpdateAsset.Asset.ID)
	assert.Equal(t, groupID.String(), resp.UpdateAsset.Asset.GroupID)
	assert.Equal(t, "IMAGE", resp.UpdateAsset.Asset.PreviewType)
	assert.Equal(t, "test.jpg", resp.UpdateAsset.Asset.FileName)

	mockAssetService.AssertExpectations(t)
}

func TestDeleteAsset(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	assetID := asset.NewAssetID()

	mockAssetService.On("DeleteAsset", mock.Anything, assetID).Return(nil)

	var resp struct {
		DeleteAsset struct {
			AssetID string
		}
	}

	mutation := `
		mutation DeleteAsset($input: DeleteAssetInput!) {
			deleteAsset(input: $input) {
				assetId
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"id": assetID.String(),
	}))

	assert.Equal(t, assetID.String(), resp.DeleteAsset.AssetID)

	mockAssetService.AssertExpectations(t)
}

func TestDecompressAsset(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	assetID := asset.NewAssetID()
	groupID := asset.NewGroupID()
	now := time.Now()
	status := asset.ExtractionStatusDone

	mockAsset := &asset.Asset{
		ID:                      assetID,
		GroupID:                 groupID,
		CreatedAt:               now,
		Size:                    1024,
		ContentType:             "application/zip",
		UUID:                    "test-uuid",
		URL:                     "http://example.com/test.zip",
		FileName:                "test.zip",
		ArchiveExtractionStatus: &status,
	}

	mockAssetService.On("DecompressAsset", mock.Anything, assetID).Return(nil)
	mockAssetService.On("GetAsset", mock.Anything, assetID).Return(mockAsset, nil)

	var resp struct {
		DecompressAsset struct {
			Asset struct {
				ID                      string
				GroupID                 string
				FileName                string
				ArchiveExtractionStatus string
			}
		}
	}

	mutation := `
		mutation DecompressAsset($input: DecompressAssetInput!) {
			decompressAsset(input: $input) {
				asset {
					id
					groupId
					fileName
					archiveExtractionStatus
				}
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"id": assetID.String(),
	}))

	assert.Equal(t, assetID.String(), resp.DecompressAsset.Asset.ID)
	assert.Equal(t, groupID.String(), resp.DecompressAsset.Asset.GroupID)
	assert.Equal(t, "test.zip", resp.DecompressAsset.Asset.FileName)
	assert.Equal(t, "DONE", resp.DecompressAsset.Asset.ArchiveExtractionStatus)

	mockAssetService.AssertExpectations(t)
}

func TestCreateAssetUpload(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	groupID := asset.NewGroupID()

	mockUploadInfo := &asset.AssetUploadInfo{
		Token:           "upload-token",
		URL:             "http://example.com/upload",
		ContentType:     "image/jpeg",
		ContentLength:   1024,
		ContentEncoding: "gzip",
		Next:            "next-cursor",
	}

	mockAssetService.On("CreateAssetUpload", mock.Anything, mock.MatchedBy(func(param asset.CreateAssetUploadParam) bool {
		return param.GroupID == groupID &&
			param.FileName == "test.jpg" &&
			param.ContentLength == 1024 &&
			param.ContentEncoding == "gzip" &&
			param.Cursor == "cursor"
	})).Return(mockUploadInfo, nil)

	var resp struct {
		CreateAssetUpload struct {
			Token           string
			URL             string
			ContentType     string
			ContentLength   int
			ContentEncoding string
			Next            string
		}
	}

	mutation := `
		mutation CreateAssetUpload($input: CreateAssetUploadInput!) {
			createAssetUpload(input: $input) {
				token
				url
				contentType
				contentLength
				contentEncoding
				next
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"groupId":         groupID.String(),
		"fileName":        "test.jpg",
		"contentLength":   1024,
		"contentEncoding": "gzip",
		"cursor":          "cursor",
	}))

	assert.Equal(t, "upload-token", resp.CreateAssetUpload.Token)
	assert.Equal(t, "http://example.com/upload", resp.CreateAssetUpload.URL)
	assert.Equal(t, "image/jpeg", resp.CreateAssetUpload.ContentType)
	assert.Equal(t, 1024, resp.CreateAssetUpload.ContentLength)
	assert.Equal(t, "gzip", resp.CreateAssetUpload.ContentEncoding)
	assert.Equal(t, "next-cursor", resp.CreateAssetUpload.Next)

	mockAssetService.AssertExpectations(t)
}

func TestDeleteAssets(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	assetID1 := asset.NewAssetID()
	assetID2 := asset.NewAssetID()

	mockAssetService.On("DeleteAssets", mock.Anything, mock.MatchedBy(func(ids []asset.AssetID) bool {
		if len(ids) != 2 {
			return false
		}
		return ids[0] == assetID1 && ids[1] == assetID2
	})).Return(nil)

	var resp struct {
		DeleteAssets struct {
			AssetIds []string
		}
	}

	mutation := `
		mutation DeleteAssets($input: DeleteAssetsInput!) {
			deleteAssets(input: $input) {
				assetIds
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"ids": []string{assetID1.String(), assetID2.String()},
	}))

	assert.Equal(t, 2, len(resp.DeleteAssets.AssetIds))
	assert.Equal(t, assetID1.String(), resp.DeleteAssets.AssetIds[0])
	assert.Equal(t, assetID2.String(), resp.DeleteAssets.AssetIds[1])

	mockAssetService.AssertExpectations(t)
}

func TestQueryGroup(t *testing.T) {
	c, _, mockGroupService, _ := setupClient(t)

	groupID := asset.NewGroupID()
	policyID := asset.NewPolicyID()
	now := time.Now()
	mockGroup := &asset.Group{
		ID:        groupID,
		Name:      "test-group",
		CreatedAt: now,
		UpdatedAt: now,
		PolicyID:  &policyID,
	}

	mockGroupService.On("GetGroup", mock.Anything, groupID).Return(mockGroup, nil)

	var resp struct {
		Group struct {
			ID     string
			Name   string
			Policy struct {
				ID string
			}
		}
	}

	query := `
		query GetGroup($id: ID!) {
			group(id: $id) {
				id
				name
				policy {
					id
				}
			}
		}
	`

	c.MustPost(query, &resp, client.Var("id", groupID.String()))

	assert.Equal(t, groupID.String(), resp.Group.ID)
	assert.Equal(t, "test-group", resp.Group.Name)
	assert.Equal(t, policyID.String(), resp.Group.Policy.ID)

	mockGroupService.AssertExpectations(t)
}

func TestCreateGroup(t *testing.T) {
	c, _, mockGroupService, _ := setupClient(t)

	groupID := asset.NewGroupID()
	now := time.Now()
	mockGroup := &asset.Group{
		ID:        groupID,
		Name:      "new-group",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockGroupService.On("CreateGroup", mock.Anything, mock.MatchedBy(func(param struct{ Name string }) bool {
		return param.Name == "new-group"
	})).Return(mockGroup, nil)

	var resp struct {
		CreateGroup struct {
			Group struct {
				ID   string
				Name string
			}
		}
	}

	mutation := `
		mutation CreateGroup($input: CreateGroupInput!) {
			createGroup(input: $input) {
				group {
					id
					name
				}
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"name": "new-group",
	}))

	assert.Equal(t, groupID.String(), resp.CreateGroup.Group.ID)
	assert.Equal(t, "new-group", resp.CreateGroup.Group.Name)

	mockGroupService.AssertExpectations(t)
}

func TestQueryPolicy(t *testing.T) {
	c, _, _, mockPolicyService := setupClient(t)

	policyID := asset.NewPolicyID()
	now := time.Now()
	mockPolicy := &asset.Policy{
		ID:           policyID,
		Name:         "test-policy",
		StorageLimit: 10240,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockPolicyService.On("GetPolicy", mock.Anything, policyID).Return(mockPolicy, nil)

	var resp struct {
		Policy struct {
			ID           string
			Name         string
			StorageLimit int
		}
	}

	query := `
		query GetPolicy($id: ID!) {
			policy(id: $id) {
				id
				name
				storageLimit
			}
		}
	`

	c.MustPost(query, &resp, client.Var("id", policyID.String()))

	assert.Equal(t, policyID.String(), resp.Policy.ID)
	assert.Equal(t, "test-policy", resp.Policy.Name)
	assert.Equal(t, 10240, resp.Policy.StorageLimit)

	mockPolicyService.AssertExpectations(t)
}

func TestCreatePolicy(t *testing.T) {
	c, _, _, mockPolicyService := setupClient(t)

	policyID := asset.NewPolicyID()
	now := time.Now()
	mockPolicy := &asset.Policy{
		ID:           policyID,
		Name:         "new-policy",
		StorageLimit: 20480,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockPolicyService.On("CreatePolicy", mock.Anything, mock.MatchedBy(func(param struct {
		Name         string
		StorageLimit int64
	}) bool {
		return param.Name == "new-policy" && param.StorageLimit == 20480
	})).Return(mockPolicy, nil)

	var resp struct {
		CreatePolicy struct {
			Policy struct {
				ID           string
				Name         string
				StorageLimit int
			}
		}
	}

	mutation := `
		mutation CreatePolicy($input: CreatePolicyInput!) {
			createPolicy(input: $input) {
				policy {
					id
					name
					storageLimit
				}
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"name":         "new-policy",
		"storageLimit": 20480,
	}))

	assert.Equal(t, policyID.String(), resp.CreatePolicy.Policy.ID)
	assert.Equal(t, "new-policy", resp.CreatePolicy.Policy.Name)
	assert.Equal(t, 20480, resp.CreatePolicy.Policy.StorageLimit)

	mockPolicyService.AssertExpectations(t)
}

func TestDeleteGroup(t *testing.T) {
	c, _, mockGroupService, _ := setupClient(t)

	groupID := asset.NewGroupID()

	mockGroupService.On("DeleteGroup", mock.Anything, groupID).Return(nil)

	var resp struct {
		DeleteGroup struct {
			GroupID string
		}
	}

	mutation := `
		mutation DeleteGroup($input: DeleteGroupInput!) {
			deleteGroup(input: $input) {
				groupId
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"id": groupID.String(),
	}))

	assert.Equal(t, groupID.String(), resp.DeleteGroup.GroupID)

	mockGroupService.AssertExpectations(t)
}

func TestDeletePolicy(t *testing.T) {
	c, _, _, mockPolicyService := setupClient(t)

	policyID := asset.NewPolicyID()

	mockPolicyService.On("DeletePolicy", mock.Anything, policyID).Return(nil)

	var resp struct {
		DeletePolicy struct {
			PolicyID string
		}
	}

	mutation := `
		mutation DeletePolicy($input: DeletePolicyInput!) {
			deletePolicy(input: $input) {
				policyId
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"id": policyID.String(),
	}))

	assert.Equal(t, policyID.String(), resp.DeletePolicy.PolicyID)

	mockPolicyService.AssertExpectations(t)
}

func TestAssignPolicy(t *testing.T) {
	c, _, mockGroupService, _ := setupClient(t)

	groupID := asset.NewGroupID()
	policyID := asset.NewPolicyID()
	now := time.Now()
	mockGroup := &asset.Group{
		ID:        groupID,
		Name:      "test-group",
		CreatedAt: now,
		UpdatedAt: now,
		PolicyID:  &policyID,
	}

	mockGroupService.On("AssignPolicy", mock.Anything, groupID, mock.MatchedBy(func(pid *asset.PolicyID) bool {
		return pid != nil && *pid == policyID
	})).Return(mockGroup, nil)

	var resp struct {
		AssignPolicy struct {
			Group struct {
				ID     string
				Name   string
				Policy struct {
					ID string
				}
			}
		}
	}

	mutation := `
		mutation AssignPolicy($input: AssignPolicyInput!) {
			assignPolicy(input: $input) {
				group {
					id
					name
					policy {
						id
					}
				}
			}
		}
	`

	c.MustPost(mutation, &resp, client.Var("input", map[string]interface{}{
		"groupId":  groupID.String(),
		"policyId": policyID.String(),
	}))

	assert.Equal(t, groupID.String(), resp.AssignPolicy.Group.ID)
	assert.Equal(t, "test-group", resp.AssignPolicy.Group.Name)
	assert.Equal(t, policyID.String(), resp.AssignPolicy.Group.Policy.ID)

	mockGroupService.AssertExpectations(t)
}

func TestQueryAssetFile(t *testing.T) {
	c, mockAssetService, _, _ := setupClient(t)

	assetID := asset.NewAssetID()
	mockAssetFile := &asset.AssetFile{
		Name:            "test.jpg",
		Size:            1024,
		ContentType:     "image/jpeg",
		ContentEncoding: "gzip",
		Path:            "/path/to/file",
		FilePaths:       []string{"/path/to/file"},
	}

	mockAssetService.On("GetAssetFile", mock.Anything, assetID).Return(mockAssetFile, nil)

	var resp struct {
		AssetFile struct {
			Name            string
			Size            int
			ContentType     string
			ContentEncoding string
		}
	}

	query := `
		query GetAssetFile($assetId: ID!) {
			assetFile(assetId: $assetId) {
				name
				size
				contentType
				contentEncoding
			}
		}
	`

	c.MustPost(query, &resp, client.Var("assetId", assetID.String()))

	assert.Equal(t, "test.jpg", resp.AssetFile.Name)
	assert.Equal(t, 1024, resp.AssetFile.Size)
	assert.Equal(t, "image/jpeg", resp.AssetFile.ContentType)
	assert.Equal(t, "gzip", resp.AssetFile.ContentEncoding)

	mockAssetService.AssertExpectations(t)
}
