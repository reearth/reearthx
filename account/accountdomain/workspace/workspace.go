package workspace

import (
	"time"

	"github.com/reearth/reearthx/util"
)

type Workspace struct {
	id       ID
	name     string
	alias    string
	email    string
	metadata *Metadata
	members  *Members
	policy   *PolicyID

	// Fields owned by reearth-accounts, which shares the workspace collection. reearthx only carries them so
	// that saving a workspace does not erase them.
	createdAt  *time.Time
	createdBy  *UserID
	deletedAt  *time.Time
	scimConfig *ScimConfig
}

func (w *Workspace) ID() ID {
	return w.id
}

func (w *Workspace) Name() string {
	return w.name
}

func (w *Workspace) Alias() string {
	return w.alias
}

func (w *Workspace) Email() string {
	return w.email
}

func (w *Workspace) Metadata() *Metadata {
	return w.metadata
}

func (w *Workspace) Members() *Members {
	return w.members
}

func (w *Workspace) IsPersonal() bool {
	return w.members.Fixed()
}

func (w *Workspace) Rename(name string) {
	w.name = name
}

func (w *Workspace) UpdateAlias(alias string) {
	w.alias = alias
}

func (w *Workspace) UpdateEmail(email string) {
	w.email = email
}

func (w *Workspace) SetMetadata(metadata *Metadata) {
	w.metadata = util.CloneRef(metadata)
}

func (w *Workspace) Policy() *PolicyID {
	return util.CloneRef(w.policy)
}

func (w *Workspace) PolicytOr(def PolicyID) PolicyID {
	if w.policy == nil {
		return def
	}
	return *w.policy
}

func (w *Workspace) SetPolicy(policy *PolicyID) {
	w.policy = util.CloneRef(policy)
}

// CreatedAt returns when the workspace was created, or nil if that is unknown.
func (w *Workspace) CreatedAt() *time.Time {
	return util.CloneRef(w.createdAt)
}

// CreatedBy returns the user who created the workspace, or nil if that is unknown.
func (w *Workspace) CreatedBy() *UserID {
	return util.CloneRef(w.createdBy)
}

// DeletedAt returns when reearth-accounts deleted the workspace, or nil if it is not deleted.
func (w *Workspace) DeletedAt() *time.Time {
	return util.CloneRef(w.deletedAt)
}

// ScimConfig returns the SCIM provisioning settings managed by reearth-accounts, or nil.
func (w *Workspace) ScimConfig() *ScimConfig {
	return w.scimConfig.Clone()
}
