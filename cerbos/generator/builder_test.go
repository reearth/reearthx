package generator

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceBuilder(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
	}{
		{
			name:        "create new builder",
			serviceName: "flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewResourceBuilder(tt.serviceName)
			assert.NotNil(t, builder)
			assert.Equal(t, tt.serviceName, builder.serviceName)
			assert.Empty(t, builder.resources)
		})
	}
}

func TestNewActionDefinition(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		roles    []string
		expected ActionDefinition
	}{
		{
			name:   "create action definition",
			action: "read",
			roles:  []string{"owner", "reader"},
			expected: ActionDefinition{
				Action: "read",
				Roles:  []string{"owner", "reader"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewActionDefinition(tt.action, tt.roles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceBuilder_AddResource(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		resource    string
		actions     []ActionDefinition
		expected    map[string][]ActionDefinition
	}{
		{
			name:        "add single resource",
			serviceName: "flow",
			resource:    "project",
			actions: []ActionDefinition{
				{
					Action: "read",
					Roles:  []string{"owner", "reader"},
				},
			},
			expected: map[string][]ActionDefinition{
				"project": {
					{
						Action: "read",
						Roles:  []string{"owner", "reader"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewResourceBuilder(tt.serviceName)
			builder.AddResource(tt.resource, tt.actions)
			assert.Equal(t, tt.expected, builder.resources)
		})
	}
}

func TestResourceBuilder_Build(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		resources   map[string][]ActionDefinition
		expected    []ResourceDefinition
	}{
		{
			name:        "build single resource",
			serviceName: "flow",
			resources: map[string][]ActionDefinition{
				"project": {
					{
						Action: "read",
						Roles:  []string{"owner", "reader"},
					},
				},
			},
			expected: []ResourceDefinition{
				{
					Resource: "flow:project",
					Actions: []ActionDefinition{
						{
							Action: "read",
							Roles:  []string{"owner", "reader"},
						},
					},
				},
			},
		},
		{
			name:        "build multiple resources",
			serviceName: "flow",
			resources: map[string][]ActionDefinition{
				"project": {
					{
						Action: "read",
						Roles:  []string{"owner", "reader"},
					},
					{
						Action: "write",
						Roles:  []string{"owner"},
					},
				},
				"workflow": {
					{
						Action: "read",
						Roles:  []string{"owner", "viewer"},
					},
				},
			},
			expected: []ResourceDefinition{
				{
					Resource: "flow:project",
					Actions: []ActionDefinition{
						{
							Action: "read",
							Roles:  []string{"owner", "reader"},
						},
						{
							Action: "write",
							Roles:  []string{"owner"},
						},
					},
				},
				{
					Resource: "flow:workflow",
					Actions: []ActionDefinition{
						{
							Action: "read",
							Roles:  []string{"owner", "viewer"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &ResourceBuilder{
				serviceName: tt.serviceName,
				resources:   tt.resources,
			}
			result := builder.Build()
			sort.Slice(result, func(i, j int) bool {
				return result[i].Resource < result[j].Resource
			})
			sort.Slice(tt.expected, func(i, j int) bool {
				return tt.expected[i].Resource < tt.expected[j].Resource
			})
			assert.Equal(t, tt.expected, result)
		})
	}
}
