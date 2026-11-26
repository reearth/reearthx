package mongodoc

import (
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/mongox"
	"github.com/samber/lo"
)

type WorkspaceMemberDocument struct {
	Role       string
	InvitedBy  string
	Disabled   bool
	ExternalID string `bson:"externalid,omitempty"`
}

type WorkspaceScimConfigDocument struct {
	Enabled          bool
	GroupRoleMapping map[string]string
	TokenHash        string
}

type WorkspaceMetadataDocument struct {
	Description  string
	Website      string
	Location     string
	BillingEmail string
	PhotoURL     string
}

type WorkspaceDocument struct {
	ID           string
	Name         string
	Alias        string
	Email        string
	Metadata     *WorkspaceMetadataDocument
	Members      map[string]WorkspaceMemberDocument
	Integrations map[string]WorkspaceMemberDocument
	Personal     bool
	Policy       string `bson:",omitempty"`
	// MembersHash is maintained for reearth-accounts, which shares this collection and has a unique index on
	// (alias, members_hash). See ComputeWorkspaceMembersHash.
	MembersHash string `bson:"members_hash,omitempty"`
	// Fields shared with reearth-accounts. ReplaceOne writes the whole document, so they must round-trip.
	ScimConfig *WorkspaceScimConfigDocument `bson:"scimconfig,omitempty"`
	CreatedAt  *time.Time                   `bson:"createdat,omitempty"`
	CreatedBy  string                       `bson:"createdby,omitempty"`
	UpdatedAt  time.Time                    `bson:"updatedat"`
	DeletedAt  *time.Time                   `bson:"deletedat,omitempty"`
}

func NewWorkspace(ws *workspace.Workspace) (*WorkspaceDocument, string) {
	membersDoc := map[string]WorkspaceMemberDocument{}
	for uId, m := range ws.Members().Users() {
		membersDoc[uId.String()] = WorkspaceMemberDocument{
			Role:       string(m.Role),
			Disabled:   m.Disabled,
			InvitedBy:  m.InvitedBy.String(),
			ExternalID: m.ExternalID,
		}
	}

	integrationsDoc := map[string]WorkspaceMemberDocument{}
	for iId, m := range ws.Members().Integrations() {
		integrationsDoc[iId.String()] = WorkspaceMemberDocument{
			Role:       string(m.Role),
			Disabled:   m.Disabled,
			InvitedBy:  m.InvitedBy.String(),
			ExternalID: m.ExternalID,
		}
	}

	metadataDoc := &WorkspaceMetadataDocument{}
	if ws.Metadata() != nil {
		metadataDoc = &WorkspaceMetadataDocument{
			Description:  ws.Metadata().Description(),
			Website:      ws.Metadata().Website(),
			Location:     ws.Metadata().Location(),
			BillingEmail: ws.Metadata().BillingEmail(),
			PhotoURL:     ws.Metadata().PhotoURL(),
		}
	}

	// A marshalling error cannot happen with these types; an empty hash is left out of the document.
	membersHash, _ := ComputeWorkspaceMembersHash(membersDoc, integrationsDoc)

	wId := ws.ID().String()
	return &WorkspaceDocument{
		ID:           wId,
		Name:         ws.Name(),
		Alias:        ws.Alias(),
		Email:        ws.Email(),
		Metadata:     metadataDoc,
		Members:      membersDoc,
		Integrations: integrationsDoc,
		Personal:     ws.IsPersonal(),
		Policy:       lo.FromPtr(ws.Policy()).String(),
		MembersHash:  membersHash,
		ScimConfig:   newScimConfigDocument(ws.ScimConfig()),
		CreatedAt:    createdAt(ws.CreatedAt(), ws.ID().Timestamp()),
		CreatedBy:    lo.FromPtr(ws.CreatedBy()).String(),
		UpdatedAt:    time.Now(),
		DeletedAt:    ws.DeletedAt(),
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
				Role:       workspace.Role(member.Role),
				Disabled:   member.Disabled,
				InvitedBy:  inviterID,
				ExternalID: member.ExternalID,
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
				Role:       workspace.Role(integrationDoc.Role),
				Disabled:   integrationDoc.Disabled,
				InvitedBy:  accountdomain.MustUserID(integrationDoc.InvitedBy),
				ExternalID: integrationDoc.ExternalID,
			}
		}
	}

	var policy *workspace.PolicyID
	if d.Policy != "" {
		policy = workspace.PolicyID(d.Policy).Ref()
	}

	var createdBy *accountdomain.UserID
	if d.CreatedBy != "" {
		if uid, err := accountdomain.UserIDFrom(d.CreatedBy); err == nil {
			createdBy = &uid
		}
	}

	metadata := &workspace.Metadata{}
	if d.Metadata != nil {
		metadata = workspace.MetadataFrom(d.Metadata.Description, d.Metadata.Website, d.Metadata.Location, d.Metadata.BillingEmail, d.Metadata.PhotoURL)
	}

	return workspace.New().
		ID(tid).
		Name(d.Name).
		Alias(d.Alias).
		Email(d.Email).
		Metadata(metadata).
		Members(members).
		Integrations(integrations).
		Personal(d.Personal).
		Policy(policy).
		ScimConfig(d.ScimConfig.Model()).
		CreatedAt(d.CreatedAt).
		CreatedBy(createdBy).
		DeletedAt(d.DeletedAt).
		Build()
}

func newScimConfigDocument(c *workspace.ScimConfig) *WorkspaceScimConfigDocument {
	if c == nil {
		return nil
	}
	m := make(map[string]string, len(c.GroupRoleMapping))
	for k, v := range c.GroupRoleMapping {
		m[k] = string(v)
	}
	return &WorkspaceScimConfigDocument{
		Enabled:          c.Enabled,
		GroupRoleMapping: m,
		TokenHash:        c.TokenHash,
	}
}

func (d *WorkspaceScimConfigDocument) Model() *workspace.ScimConfig {
	if d == nil {
		return nil
	}
	m := make(map[string]workspace.Role, len(d.GroupRoleMapping))
	for k, v := range d.GroupRoleMapping {
		m[k] = workspace.Role(v)
	}
	return &workspace.ScimConfig{
		Enabled:          d.Enabled,
		GroupRoleMapping: m,
		TokenHash:        d.TokenHash,
	}
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
	return NewConsumer[*WorkspaceDocument, *workspace.Workspace]()
}
