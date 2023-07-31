//go:generate go run github.com/Khan/genqlient

package accountproxy

import (
	"context"

	_ "github.com/Khan/genqlient/generate"
	"github.com/Khan/genqlient/graphql"
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/account/accountusecase/accountinterfaces"
)

type Workspace struct {
	http     HTTPClient
	gql      graphql.Client
	endpoint string
}

func NewWorkspace(endpoint string, h HTTPClient) accountinterfaces.Workspace {
	return &Workspace{
		http:     h,
		endpoint: endpoint,
		gql:      graphql.NewClient(endpoint, h),
	}
}

func (w *Workspace) Fetch(ctx context.Context, ids accountdomain.WorkspaceIDList, op *accountusecase.Operator) ([]*workspace.Workspace, error) {
	return WorkspaceByIDsResponseTo(WorkspaceByIDs(ctx, w.gql, ids.Strings()))
}

func (w *Workspace) FindByUser(ctx context.Context, userID accountdomain.UserID, op *accountusecase.Operator) ([]*workspace.Workspace, error) {
	res, err := FindByUser(ctx, w.gql, userID.String())
	if err != nil {
		return nil, err
	}
	ws := make([]FragmentWorkspace, len(res.FindByUser))
	for i, w := range res.FindByUser {
		ws[i] = w.FragmentWorkspace
	}
	return ToWorkspaces(ws)
}

func (w *Workspace) FetchPolicy(ctx context.Context, policyID []workspace.PolicyID, op *accountusecase.Operator) ([]*workspace.Policy, error) {
	policyIDs := make([]string, len(policyID))
	for i, id := range policyID {
		policyIDs[i] = id.String()
	}
	res, err := FetchPolicy(ctx, w.gql, policyIDs)
	if err != nil {
		return nil, err
	}
	return ToPolicies(res), nil
}

func (w *Workspace) Create(ctx context.Context, name string, userID accountdomain.UserID, op *accountusecase.Operator) (*workspace.Workspace, error) {
	res, err := CreateWorkspace(ctx, w.gql, CreateWorkspaceInput{Name: name})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.CreateWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) Update(ctx context.Context, id accountdomain.WorkspaceID, name string, op *accountusecase.Operator) (*workspace.Workspace, error) {
	res, err := UpdateWorkspace(ctx, w.gql, UpdateWorkspaceInput{WorkspaceId: id.String(), Name: name})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.UpdateWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) AddUserMember(ctx context.Context, id accountdomain.WorkspaceID, users map[accountdomain.UserID]workspace.Role, op *accountusecase.Operator) (*workspace.Workspace, error) {
	members := []MemberInput{}
	for id, role := range users {
		members = append(members, MemberInput{UserId: id.String(), Role: Role(string(role))})
	}
	res, err := AddUsersToWorkspace(ctx, w.gql, AddUsersToWorkspaceInput{WorkspaceId: id.String(), Users: members})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.AddUsersToWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) AddIntegrationMember(ctx context.Context, id accountdomain.WorkspaceID, integrationId accountdomain.IntegrationID, role workspace.Role, op *accountusecase.Operator) (*workspace.Workspace, error) {
	res, err := AddIntegrationToWorkspace(ctx, w.gql, AddIntegrationToWorkspaceInput{WorkspaceId: id.String(), IntegrationId: integrationId.String(), Role: Role(string(role))})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.AddIntegrationToWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) UpdateUserMember(ctx context.Context, id accountdomain.WorkspaceID, userID accountdomain.UserID, role workspace.Role, op *accountusecase.Operator) (*workspace.Workspace, error) {
	res, err := UpdateUserOfWorkspace(ctx, w.gql, UpdateUserOfWorkspaceInput{WorkspaceId: id.String(), UserId: userID.String(), Role: Role(string(role))})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.UpdateUserOfWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) UpdateIntegration(ctx context.Context, id accountdomain.WorkspaceID, integrationID accountdomain.IntegrationID, role workspace.Role, op *accountusecase.Operator) (*workspace.Workspace, error) {
	res, err := UpdateIntegrationOfWorkspace(ctx, w.gql, UpdateIntegrationOfWorkspaceInput{WorkspaceId: id.String(), IntegrationId: integrationID.String(), Role: Role(string(role))})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.UpdateIntegrationOfWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) RemoveUserMember(ctx context.Context, id accountdomain.WorkspaceID, userID accountdomain.UserID, op *accountusecase.Operator) (*workspace.Workspace, error) {
	res, err := RemoveUserFromWorkspace(ctx, w.gql, RemoveUserFromWorkspaceInput{WorkspaceId: id.String(), UserId: userID.String()})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.RemoveUserFromWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) RemoveIntegration(ctx context.Context, id accountdomain.WorkspaceID, integrationID accountdomain.IntegrationID, op *accountusecase.Operator) (*workspace.Workspace, error) {
	res, err := RemoveIntegrationFromWorkspace(ctx, w.gql, RemoveIntegrationFromWorkspaceInput{WorkspaceId: id.String(), IntegrationId: integrationID.String()})
	if err != nil {
		return nil, err
	}
	return ToWorkspace(res.RemoveIntegrationFromWorkspace.Workspace.FragmentWorkspace)
}

func (w *Workspace) Remove(ctx context.Context, id accountdomain.WorkspaceID, op *accountusecase.Operator) error {
	_, err := DeleteWorkspace(ctx, w.gql, DeleteWorkspaceInput{WorkspaceId: id.String()})
	if err != nil {
		return err
	}
	return nil
}
