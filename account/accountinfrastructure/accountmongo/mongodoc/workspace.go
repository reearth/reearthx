package mongodoc

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/mongox"
)

type WorkspaceMemberDocument struct {
	Role      string
	InvitedBy string
	Disabled  bool
}

type WorkspaceDocument struct {
	ID           string
	Name         string
	Members      map[string]WorkspaceMemberDocument
	Integrations map[string]WorkspaceMemberDocument
	Personal     bool
}

func NewWorkspace(ws *workspace.Workspace) (*WorkspaceDocument, string) {
	membersDoc := map[string]WorkspaceMemberDocument{}
	for uId, m := range ws.Members().Users() {
		membersDoc[uId.String()] = WorkspaceMemberDocument{
			Role:      string(m.Role),
			Disabled:  m.Disabled,
			InvitedBy: m.InvitedBy.String(),
		}
	}
	integrationsDoc := map[string]WorkspaceMemberDocument{}
	for iId, m := range ws.Members().Integrations() {
		integrationsDoc[iId.String()] = WorkspaceMemberDocument{
			Role:      string(m.Role),
			Disabled:  m.Disabled,
			InvitedBy: m.InvitedBy.String(),
		}
	}
	wId := ws.ID().String()
	return &WorkspaceDocument{
		ID:           wId,
		Name:         ws.Name(),
		Members:      membersDoc,
		Integrations: integrationsDoc,
		Personal:     ws.IsPersonal(),
	}, wId
}

func (d *WorkspaceDocument) Model() (*workspace.Workspace, error) {
	tid, err := accountdomain.WorkspaceIDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	members := map[accountdomain.UserID]workspace.Member{}
	if d.Members != nil {
		for uid, member := range d.Members {
			uid, err := accountdomain.UserIDFrom(uid)
			if err != nil {
				return nil, err
			}
			inviterID, err := accountdomain.UserIDFrom(member.InvitedBy)
			if err != nil {
				inviterID = uid
			}
			members[uid] = workspace.Member{
				Role:      workspace.Role(member.Role),
				Disabled:  member.Disabled,
				InvitedBy: inviterID,
			}
		}
	}
	integrations := map[accountdomain.IntegrationID]workspace.Member{}
	if d.Integrations != nil {
		for iId, integrationDoc := range d.Integrations {
			iId, err := accountdomain.IntegrationIDFrom(iId)
			if err != nil {
				return nil, err
			}
			integrations[iId] = workspace.Member{
				Role:      workspace.Role(integrationDoc.Role),
				Disabled:  integrationDoc.Disabled,
				InvitedBy: accountdomain.MustUserID(integrationDoc.InvitedBy),
			}
		}
	}
	return workspace.NewWorkspace().
		ID(tid).
		Name(d.Name).
		Members(members).
		Integrations(integrations).
		Personal(d.Personal).
		Build()
}

func NewWorkspaces(workspaces []*workspace.Workspace) ([]*WorkspaceDocument, []string) {
	res := make([]*WorkspaceDocument, 0, len(workspaces))
	ids := make([]string, 0, len(workspaces))
	for _, d := range workspaces {
		if d == nil {
			continue
		}
		r, wId := NewWorkspace(d)
		res = append(res, r)
		ids = append(ids, wId)
	}
	return res, ids
}

type WorkspaceConsumer = mongox.SliceFuncConsumer[*WorkspaceDocument, *workspace.Workspace]

func NewWorkspaceConsumer() *WorkspaceConsumer {
	return NewComsumer[*WorkspaceDocument, *workspace.Workspace]()
}
