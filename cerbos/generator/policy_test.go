package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePolicies(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name            string
		serviceName     string
		defineResources func(builder *ResourceBuilder) []ResourceDefinition
		wantFiles       map[string]string
		outputDir       string
		wantErr         string
	}{
		{
			name:        "success generate single policy",
			serviceName: "flow",
			defineResources: func(builder *ResourceBuilder) []ResourceDefinition {
				return builder.
					AddResource("project", []ActionDefinition{
						NewActionDefinition("read", []string{"owner", "reader"}),
					}).
					Build()
			},
			wantFiles: map[string]string{
				"flow_project.yaml": `apiVersion: api.cerbos.dev/v1
resourcePolicy:
  version: default
  resource: flow:project
  rules:
  - actions:
    - read
    effect: EFFECT_ALLOW
    roles:
    - owner
    - reader
`,
			},
		},
		{
			name:        "success generate multiple policies",
			serviceName: "flow",
			defineResources: func(builder *ResourceBuilder) []ResourceDefinition {
				return builder.
					AddResource("project", []ActionDefinition{
						NewActionDefinition("read", []string{"owner", "reader"}),
						NewActionDefinition("write", []string{"owner"}),
					}).
					AddResource("workflow", []ActionDefinition{
						NewActionDefinition("read", []string{"owner", "viewer"}),
					}).
					Build()
			},
			wantFiles: map[string]string{
				"flow_project.yaml": `apiVersion: api.cerbos.dev/v1
resourcePolicy:
  version: default
  resource: flow:project
  rules:
  - actions:
    - read
    effect: EFFECT_ALLOW
    roles:
    - owner
    - reader
  - actions:
    - write
    effect: EFFECT_ALLOW
    roles:
    - owner
`,
				"flow_workflow.yaml": `apiVersion: api.cerbos.dev/v1
resourcePolicy:
  version: default
  resource: flow:workflow
  rules:
  - actions:
    - read
    effect: EFFECT_ALLOW
    roles:
    - owner
    - viewer
`,
			},
		},
		{
			name:        "invalid resource definition",
			serviceName: "flow",
			defineResources: func(b *ResourceBuilder) []ResourceDefinition {
				return []ResourceDefinition{{
					Resource: "",
					Actions:  []ActionDefinition{},
				}}
			},
			outputDir: "test",
			wantErr:   "invalid resource name",
		},
		{
			name:            "nil define resources func",
			serviceName:     "flow",
			defineResources: nil,
			outputDir:       "test",
			wantErr:         "define resources function is required",
		},
		{
			name:        "success generate policy with simple condition",
			serviceName: "flow",
			defineResources: func(builder *ResourceBuilder) []ResourceDefinition {
				return builder.
					AddResource("document", []ActionDefinition{
						NewActionDefinitionWithCondition(
							"approve",
							[]string{"manager"},
							SimpleExpr(`R.attr.status == "PENDING_APPROVAL"`),
						),
					}).
					Build()
			},
			wantFiles: map[string]string{
				"flow_document.yaml": `apiVersion: api.cerbos.dev/v1
resourcePolicy:
  version: default
  resource: flow:document
  rules:
  - actions:
    - approve
    effect: EFFECT_ALLOW
    roles:
    - manager
    condition:
      match:
        expr: R.attr.status == "PENDING_APPROVAL"
`,
			},
		},
		{
			name:        "success generate policy with AllOf condition",
			serviceName: "flow",
			defineResources: func(builder *ResourceBuilder) []ResourceDefinition {
				return builder.
					AddResource("document", []ActionDefinition{
						NewActionDefinitionWithCondition(
							"approve",
							[]string{"manager"},
							AllOf(
								`R.attr.status == "PENDING_APPROVAL"`,
								`"GB" in R.attr.geographies`,
							),
						),
					}).
					Build()
			},
			wantFiles: map[string]string{
				"flow_document.yaml": `apiVersion: api.cerbos.dev/v1
resourcePolicy:
  version: default
  resource: flow:document
  rules:
  - actions:
    - approve
    effect: EFFECT_ALLOW
    roles:
    - manager
    condition:
      match:
        all:
          of:
          - expr: R.attr.status == "PENDING_APPROVAL"
          - expr: '"GB" in R.attr.geographies'
`,
			},
		},
		{
			name:        "success generate policy with AnyOf condition",
			serviceName: "flow",
			defineResources: func(builder *ResourceBuilder) []ResourceDefinition {
				return builder.
					AddResource("resource", []ActionDefinition{
						NewActionDefinitionWithCondition(
							"delete",
							[]string{"admin", "owner"},
							AnyOf(
								`P.attr.role == "admin"`,
								`R.attr.owner == P.id`,
							),
						),
					}).
					Build()
			},
			wantFiles: map[string]string{
				"flow_resource.yaml": `apiVersion: api.cerbos.dev/v1
resourcePolicy:
  version: default
  resource: flow:resource
  rules:
  - actions:
    - delete
    effect: EFFECT_ALLOW
    roles:
    - admin
    - owner
    condition:
      match:
        any:
          of:
          - expr: P.attr.role == "admin"
          - expr: R.attr.owner == P.id
`,
			},
		},
		{
			name:        "success generate policy with NoneOf condition",
			serviceName: "flow",
			defineResources: func(builder *ResourceBuilder) []ResourceDefinition {
				return builder.
					AddResource("project", []ActionDefinition{
						NewActionDefinitionWithCondition(
							"view",
							[]string{"user"},
							NoneOf(
								`R.attr.archived == true`,
								`R.attr.deleted == true`,
							),
						),
					}).
					Build()
			},
			wantFiles: map[string]string{
				"flow_project.yaml": `apiVersion: api.cerbos.dev/v1
resourcePolicy:
  version: default
  resource: flow:project
  rules:
  - actions:
    - view
    effect: EFFECT_ALLOW
    roles:
    - user
    condition:
      match:
        none:
          of:
          - expr: R.attr.archived == true
          - expr: R.attr.deleted == true
`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := tt.outputDir
			if testDir == "" {
				testDir = filepath.Join(tmpDir, tt.name)
			}

			err := GeneratePolicies(tt.serviceName, tt.defineResources, testDir)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)

			for filename, expectedContent := range tt.wantFiles {
				content, err := os.ReadFile(filepath.Join(testDir, filename))
				assert.NoError(t, err)
				assert.Equal(t, expectedContent, string(content))
			}

			files, err := os.ReadDir(testDir)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.wantFiles), len(files))
		})
	}
}
