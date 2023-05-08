package accountproxy

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/util"
)

func WorkspaceByIDsResponseTo(r *WorkspaceByIDsResponse, err error) ([]*workspace.Workspace, error) {
	if err != nil || r == nil {
		return nil, err
	}
	return util.TryMap(r.Nodes, WorkspaceByIDsNodeTo)
}

func WorkspaceByIDsNodeTo(r WorkspaceByIDsNodesNode) (*workspace.Workspace, error) {
	if r == nil {
		return nil, nil
	}
	w, ok := r.(*WorkspaceByIDsNodesWorkspace)
	if !ok || w == nil {
		return nil, nil
	}
	return ToWorkspace(w.TemplateWorkspace)
}

func ToWorkspace(r TemplateWorkspace) (*workspace.Workspace, error) {
	id, err := workspace.IDFrom(r.Id)
	if err != nil {
		return nil, err
	}
	members := map[accountdomain.UserID]workspace.Member{}
	integrations := map[accountdomain.IntegrationID]workspace.Member{}

	for i := range r.Members {
		w, ok := r.Members[i].(*TemplateWorkspaceMembersWorkspaceUserMember)
		if ok || w != nil {
			id, err := user.IDFrom(w.UserId)
			if err != nil {
				return nil, err
			}

			members[id] = workspace.Member{
				Role: workspace.Role(w.Role),
			}
		}
		in, ok := r.Members[i].(*TemplateWorkspaceMembersWorkspaceIntegrationMember)
		if ok || in != nil {
			iid, err := accountdomain.IntegrationIDFrom(in.IntegrationId)
			if err != nil {
				return nil, err
			}

			uid, err := user.IDFrom(in.InvitedById)
			if err != nil {
				return nil, err
			}

			integrations[iid] = workspace.Member{
				Role:      workspace.Role(in.Role),
				InvitedBy: uid,
			}
		}

	}
	return workspace.New().ID(id).
		Name(r.Name).Personal(r.Personal).Members(members).Integrations(integrations).Build()
}
