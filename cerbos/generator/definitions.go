package generator

import (
	"slices"
)

type ResourceRule struct {
	Resource string
	Actions  map[string]ActionRule
}

type ActionRule struct {
	Roles     []string
	Condition *Condition
}

func DefineResources(builder *ResourceBuilder, resourceRules []ResourceRule) []ResourceDefinition {
	if builder == nil {
		panic("ResourceBuilder cannot be nil")
	}

	for _, r := range resourceRules {
		var actions []ActionDefinition
		// Sort action keys to ensure deterministic output
		actionKeys := make([]string, 0, len(r.Actions))
		for action := range r.Actions {
			actionKeys = append(actionKeys, action)
		}
		slices.Sort(actionKeys)

		for _, action := range actionKeys {
			actionRule := r.Actions[action]
			if actionRule.Condition != nil {
				actions = append(actions, NewActionDefinitionWithCondition(action, actionRule.Roles, actionRule.Condition))
			} else {
				actions = append(actions, NewActionDefinition(action, actionRule.Roles))
			}
		}
		builder.AddResource(r.Resource, actions)
	}

	return builder.Build()
}
