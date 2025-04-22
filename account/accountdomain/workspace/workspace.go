package workspace

import "github.com/reearth/reearthx/util"

type Workspace struct {
	id          ID
	name        string
	displayName string
	members     *Members
	policy      *PolicyID
	location    string
}

func (w *Workspace) ID() ID {
	return w.id
}

func (w *Workspace) Name() string {
	return w.name
}

func (w *Workspace) IsValidName(name string) bool {
	return util.IsValidName(name)
}

func (w *Workspace) DisplayName() string {
	return w.displayName
}

func (w *Workspace) Members() *Members {
	return w.members
}

func (w *Workspace) IsPersonal() bool {
	return w.members.Fixed()
}

func (w *Workspace) Location() string {
	return w.location
}

func (w *Workspace) LocationOr(def string) string {
	if w.location == "" {
		return def
	}
	return w.location
}

func (w *Workspace) Rename(name string) {
	w.name = name
}

func (w *Workspace) UpdateDisplayName(displayName string) {
	w.displayName = displayName
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
