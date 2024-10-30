package generator

type ResourceBuilder struct {
	serviceName string
	resources   map[string][]actionDefinition
}

func NewResourceBuilder(serviceName string) *ResourceBuilder {
	return &ResourceBuilder{
		serviceName: serviceName,
		resources:   make(map[string][]actionDefinition),
	}
}

func (b *ResourceBuilder) AddResource(resource string, actions []actionDefinition) *ResourceBuilder {
	b.resources[resource] = actions
	return b
}

func (b *ResourceBuilder) Build() []ResourceDefinition {
	result := make([]ResourceDefinition, 0, len(b.resources))
	for resource, actions := range b.resources {
		result = append(result, &resourceDefinition{
			Resource: b.serviceName + ":" + resource,
			Actions:  actions,
		})
	}
	return result
}

func NewActionDefinition(action string, roles []string) actionDefinition {
	return actionDefinition{
		Action: action,
		Roles:  roles,
	}
}
