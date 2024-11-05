package generator

type ResourceDefinition struct {
	Resource string
	Actions  []ActionDefinition
}

type ActionDefinition struct {
	Action string
	Roles  []string
}

type ResourceBuilder struct {
	serviceName string
	resources   map[string][]ActionDefinition
}

func NewResourceBuilder(serviceName string) *ResourceBuilder {
	return &ResourceBuilder{
		serviceName: serviceName,
		resources:   make(map[string][]ActionDefinition),
	}
}

func NewActionDefinition(action string, roles []string) ActionDefinition {
	return ActionDefinition{
		Action: action,
		Roles:  roles,
	}
}

func (b *ResourceBuilder) AddResource(resource string, actions []ActionDefinition) *ResourceBuilder {
	b.resources[resource] = actions
	return b
}

func (b *ResourceBuilder) Build() []ResourceDefinition {
	result := make([]ResourceDefinition, 0, len(b.resources))
	for resource, actions := range b.resources {
		result = append(result, ResourceDefinition{
			Resource: b.serviceName + ":" + resource,
			Actions:  actions,
		})
	}
	return result
}
