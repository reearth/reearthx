package workspace

import "github.com/reearth/reearthx/util"

type Workspace struct {
	id      ID
	name    string
	members *Members
	policy  *PolicyID
}

func (t *Workspace) ID() ID {
	return t.id
}

func (t *Workspace) Name() string {
	return t.name
}

func (t *Workspace) Members() *Members {
	return t.members
}

func (t *Workspace) IsPersonal() bool {
	return t.members.Fixed()
}

func (t *Workspace) Rename(name string) {
	t.name = name
}

func (w *Workspace) Policy() *PolicyID {
	return util.CloneRef(w.policy)
}

func (w *Workspace) SetPolicy(policy *PolicyID) {
	w.policy = util.CloneRef(policy)
}
