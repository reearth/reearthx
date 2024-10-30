package generator

type ResourceDefinition interface {
	GetResource() string
	GetActions() []ActionDefinition
}

type ActionDefinition interface {
	GetAction() string
	GetRoles() []string
}

type resourceDefinition struct {
	Resource string
	Actions  []actionDefinition
}

type actionDefinition struct {
	Action string
	Roles  []string
}

func (r resourceDefinition) GetResource() string {
	return r.Resource
}

func (r resourceDefinition) GetActions() []ActionDefinition {
	actions := make([]ActionDefinition, len(r.Actions))
	for i, a := range r.Actions {
		actions[i] = a
	}
	return actions
}

func (a actionDefinition) GetAction() string {
	return a.Action
}

func (a actionDefinition) GetRoles() []string {
	return a.Roles
}
