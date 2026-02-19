package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefineResources(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		resourceRules []ResourceRule
		expected      []ResourceDefinition
	}{
		{
			name:        "single resource with single action",
			serviceName: "flow",
			resourceRules: []ResourceRule{
				{
					Resource: "project",
					Actions: map[string]ActionRule{
						"read": {Roles: []string{"owner", "reader"}},
					},
				},
			},
			expected: []ResourceDefinition{
				{
					Resource: "flow:project",
					Actions: []ActionDefinition{
						{Action: "read", Roles: []string{"owner", "reader"}},
					},
				},
			},
		},
		{
			name:        "single resource with multiple actions sorted",
			serviceName: "flow",
			resourceRules: []ResourceRule{
				{
					Resource: "project",
					Actions: map[string]ActionRule{
						"write": {Roles: []string{"owner"}},
						"read":  {Roles: []string{"owner", "reader"}},
					},
				},
			},
			expected: []ResourceDefinition{
				{
					Resource: "flow:project",
					Actions: []ActionDefinition{
						{Action: "read", Roles: []string{"owner", "reader"}},
						{Action: "write", Roles: []string{"owner"}},
					},
				},
			},
		},
		{
			name:        "multiple resources",
			serviceName: "flow",
			resourceRules: []ResourceRule{
				{
					Resource: "project",
					Actions: map[string]ActionRule{
						"read": {Roles: []string{"owner"}},
					},
				},
				{
					Resource: "workflow",
					Actions: map[string]ActionRule{
						"read": {Roles: []string{"viewer"}},
					},
				},
			},
			expected: []ResourceDefinition{
				{
					Resource: "flow:project",
					Actions: []ActionDefinition{
						{Action: "read", Roles: []string{"owner"}},
					},
				},
				{
					Resource: "flow:workflow",
					Actions: []ActionDefinition{
						{Action: "read", Roles: []string{"viewer"}},
					},
				},
			},
		},
		{
			name:        "action with condition",
			serviceName: "flow",
			resourceRules: []ResourceRule{
				{
					Resource: "document",
					Actions: map[string]ActionRule{
						"approve": {
							Roles:     []string{"manager"},
							Condition: SimpleExpr(`R.attr.status == "PENDING"`),
						},
					},
				},
			},
			expected: []ResourceDefinition{
				{
					Resource: "flow:document",
					Actions: []ActionDefinition{
						{
							Action:    "approve",
							Roles:     []string{"manager"},
							Condition: SimpleExpr(`R.attr.status == "PENDING"`),
						},
					},
				},
			},
		},
		{
			name:        "action without condition",
			serviceName: "flow",
			resourceRules: []ResourceRule{
				{
					Resource: "project",
					Actions: map[string]ActionRule{
						"read": {Roles: []string{"owner"}},
					},
				},
			},
			expected: []ResourceDefinition{
				{
					Resource: "flow:project",
					Actions: []ActionDefinition{
						{Action: "read", Roles: []string{"owner"}},
					},
				},
			},
		},
		{
			name:          "empty resource rules",
			serviceName:   "flow",
			resourceRules: []ResourceRule{},
			expected:      []ResourceDefinition{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewResourceBuilder(tt.serviceName)
			result := DefineResources(builder, tt.resourceRules)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefineResources_NilBuilder(t *testing.T) {
	assert.Panics(t, func() {
		DefineResources(nil, []ResourceRule{
			{Resource: "project", Actions: map[string]ActionRule{}},
		})
	})
}
