package asset

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountusecase"
	"github.com/reearth/reearthx/idx"
)

// AssetOperator provides project and workspace access control
type AssetOperator struct {
	Integration          IntegrationID
	Machine              bool
	Lang                 string
	ReadableProjects     GroupIDList
	WritableProjects     GroupIDList
	OwningProjects       GroupIDList
	MaintainableProjects GroupIDList

	AcOperator *accountusecase.Operator
}

// Ownable interface represents objects that can be owned by a user or integration
type Ownable interface {
	User() *accountdomain.UserID
	Integration() *IntegrationID
	Project() GroupID
}

// Workspaces returns workspace IDs with the given role
func (o *AssetOperator) Workspaces(r workspace.Role) accountdomain.WorkspaceIDList {
	if o == nil || o.AcOperator == nil {
		return nil
	}
	if r == workspace.RoleReader {
		return o.AcOperator.ReadableWorkspaces
	}
	if r == workspace.RoleWriter {
		return o.AcOperator.WritableWorkspaces
	}
	if r == workspace.RoleMaintainer {
		return o.AcOperator.MaintainableWorkspaces
	}
	if r == workspace.RoleOwner {
		return o.AcOperator.OwningWorkspaces
	}
	return nil
}

// AllReadableWorkspaces returns all workspaces the operator can read
func (o *AssetOperator) AllReadableWorkspaces() accountdomain.WorkspaceIDList {
	if o == nil || o.AcOperator == nil {
		return nil
	}
	return append(o.AcOperator.ReadableWorkspaces, o.AllWritableWorkspaces()...)
}

// AllWritableWorkspaces returns all workspaces the operator can write
func (o *AssetOperator) AllWritableWorkspaces() accountdomain.WorkspaceIDList {
	if o == nil || o.AcOperator == nil {
		return nil
	}
	return append(o.AcOperator.WritableWorkspaces, o.AllMaintainingWorkspaces()...)
}

// AllMaintainingWorkspaces returns all workspaces the operator can maintain
func (o *AssetOperator) AllMaintainingWorkspaces() accountdomain.WorkspaceIDList {
	if o == nil || o.AcOperator == nil {
		return nil
	}
	return append(o.AcOperator.MaintainableWorkspaces, o.AllOwningWorkspaces()...)
}

// AllOwningWorkspaces returns all workspaces the operator owns
func (o *AssetOperator) AllOwningWorkspaces() accountdomain.WorkspaceIDList {
	if o == nil || o.AcOperator == nil {
		return nil
	}
	return o.AcOperator.OwningWorkspaces
}

// IsReadableWorkspace checks if the operator can read the given workspaces
func (o *AssetOperator) IsReadableWorkspace(workspace ...accountdomain.WorkspaceID) bool {
	return o.AllReadableWorkspaces().Intersect(workspace).Len() > 0
}

// IsWritableWorkspace checks if the operator can write to the given workspaces
func (o *AssetOperator) IsWritableWorkspace(workspace ...accountdomain.WorkspaceID) bool {
	return o.AllWritableWorkspaces().Intersect(workspace).Len() > 0
}

// IsMaintainingWorkspace checks if the operator can maintain the given workspaces
func (o *AssetOperator) IsMaintainingWorkspace(workspace ...accountdomain.WorkspaceID) bool {
	return o.AllMaintainingWorkspaces().Intersect(workspace).Len() > 0
}

// IsOwningWorkspace checks if the operator owns the given workspaces
func (o *AssetOperator) IsOwningWorkspace(workspace ...accountdomain.WorkspaceID) bool {
	return o.AllOwningWorkspaces().Intersect(workspace).Len() > 0
}

// AddNewWorkspace adds a new workspace to the operator's owned workspaces
func (o *AssetOperator) AddNewWorkspace(workspace accountdomain.WorkspaceID) {
	if o == nil || o.AcOperator == nil {
		return
	}
	o.AcOperator.OwningWorkspaces = append(o.AcOperator.OwningWorkspaces, workspace)
}

// Projects returns project IDs with the given role
func (o *AssetOperator) Projects(r workspace.Role) GroupIDList {
	if o == nil {
		return nil
	}
	if r == workspace.RoleReader {
		return o.ReadableProjects
	}
	if r == workspace.RoleWriter {
		return o.WritableProjects
	}
	if r == workspace.RoleMaintainer {
		return o.MaintainableProjects
	}
	if r == workspace.RoleOwner {
		return o.OwningProjects
	}
	return nil
}

// AllReadableProjects returns all projects the operator can read
func (o *AssetOperator) AllReadableProjects() GroupIDList {
	return append(o.ReadableProjects, o.AllWritableProjects()...)
}

// AllWritableProjects returns all projects the operator can write
func (o *AssetOperator) AllWritableProjects() GroupIDList {
	return append(o.WritableProjects, o.AllMaintainableProjects()...)
}

// AllMaintainableProjects returns all projects the operator can maintain
func (o *AssetOperator) AllMaintainableProjects() GroupIDList {
	return append(o.MaintainableProjects, o.AllOwningProjects()...)
}

// AllOwningProjects returns all projects the operator owns
func (o *AssetOperator) AllOwningProjects() GroupIDList {
	return o.OwningProjects
}

// IsReadableProject checks if the operator can read the given projects
func (o *AssetOperator) IsReadableProject(projects ...GroupID) bool {
	return o.AllReadableProjects().Intersect(projects).Len() > 0
}

// IsWritableProject checks if the operator can write to the given projects
func (o *AssetOperator) IsWritableProject(projects ...GroupID) bool {
	return o.AllWritableProjects().Intersect(projects).Len() > 0
}

// IsMaintainingProject checks if the operator can maintain the given projects
func (o *AssetOperator) IsMaintainingProject(projects ...GroupID) bool {
	return o.AllMaintainableProjects().Intersect(projects).Len() > 0
}

// IsOwningProject checks if the operator owns the given projects
func (o *AssetOperator) IsOwningProject(projects ...GroupID) bool {
	return o.AllOwningProjects().Intersect(projects).Len() > 0
}

// AddNewProject adds a new project to the operator's owned projects
func (o *AssetOperator) AddNewProject(p GroupID) {
	o.OwningProjects = append(o.OwningProjects, p)
}

// Helper functions for operator ID

// OperatorFromUser creates an operator from a user ID
func OperatorFromUser(userID accountdomain.UserID) idx.ID[OperatorIDType] {
	id, _ := idx.From[OperatorIDType]("user:" + userID.String())
	return id
}

// OperatorFromIntegration creates an operator from an integration ID
func OperatorFromIntegration(integrationID IntegrationID) idx.ID[OperatorIDType] {
	id, _ := idx.From[OperatorIDType]("integration:" + integrationID.String())
	return id
}

// OperatorFromMachine creates a machine operator
func OperatorFromMachine() idx.ID[OperatorIDType] {
	id, _ := idx.From[OperatorIDType]("machine")
	return id
}

// Operator returns an OperatorID representing this operator
func (o *AssetOperator) Operator() idx.ID[OperatorIDType] {
	if o == nil || o.AcOperator == nil {
		return idx.ID[OperatorIDType]{}
	}

	var eOp idx.ID[OperatorIDType]
	if o.AcOperator.User != nil {
		eOp = OperatorFromUser(*o.AcOperator.User)
	}
	if o.Integration != (IntegrationID{}) {
		eOp = OperatorFromIntegration(o.Integration)
	}
	if o.Machine {
		eOp = OperatorFromMachine()
	}
	return eOp
}

// CanUpdate checks if the operator can update the given object
func (o *AssetOperator) CanUpdate(obj Ownable) bool {
	isWriter := o.IsWritableProject(obj.Project())
	isMaintainer := o.IsMaintainingProject(obj.Project())
	return isMaintainer || (isWriter && o.Owns(obj)) || o.Machine
}

// Owns checks if the operator owns the given object
func (o *AssetOperator) Owns(obj Ownable) bool {
	if o == nil || o.AcOperator == nil {
		return false
	}

	return (o.AcOperator.User != nil && obj.User() != nil && *o.AcOperator.User == *obj.User()) ||
		(o.Integration != (IntegrationID{}) && obj.Integration() != nil && o.Integration == *obj.Integration())
}

// RoleByProject returns the role of the operator for the given project
func (o *AssetOperator) RoleByProject(pid GroupID) workspace.Role {
	if o.IsOwningProject(pid) {
		return workspace.RoleOwner
	}
	if o.IsMaintainingProject(pid) {
		return workspace.RoleMaintainer
	}
	if o.IsWritableProject(pid) {
		return workspace.RoleWriter
	}
	if o.IsReadableProject(pid) {
		return workspace.RoleReader
	}
	return ""
}
