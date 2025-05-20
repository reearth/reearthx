package workspace

import "github.com/reearth/reearthx/util"

type Workspace struct {
	id           ID
	name         string
	alias        string
	description  string
	website      string
	email        string
	billingEmail string
	members      *Members
	policy       *PolicyID
	location     string
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

func (w *Workspace) Description() string {
	return w.description
}

func (w *Workspace) Website() string {
	return w.website
}

func (w *Workspace) Email() string {
	return w.email
}

func (w *Workspace) BillingEmail() string {
	return w.billingEmail
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

func (w *Workspace) UpdateAlias(alias string) {
	w.alias = alias
}

func (w *Workspace) UpdateDescription(description string) {
	w.description = description
}

func (w *Workspace) UpdateWebsite(website string) {
	w.website = website
}

func (w *Workspace) UpdateEmail(email string) {
	w.email = email
}

func (w *Workspace) UpdateBillingEmail(billingEmail string) {
	w.billingEmail = billingEmail
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
