package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type CerbosPolicy struct {
	APIVersion     string         `yaml:"apiVersion"`
	ResourcePolicy ResourcePolicy `yaml:"resourcePolicy"`
}

type ResourcePolicy struct {
	Version  string `yaml:"version"`
	Resource string `yaml:"resource"`
	Rules    []Rule `yaml:"rules"`
}

type Rule struct {
	Actions   []string   `yaml:"actions"`
	Effect    string     `yaml:"effect"`
	Roles     []string   `yaml:"roles"`
	Condition *Condition `yaml:"condition,omitempty"`
}

type Condition struct {
	Match Match `yaml:"match"`
}

type Match struct {
	All  *MatchExpressions `yaml:"all,omitempty"`
	Any  *MatchExpressions `yaml:"any,omitempty"`
	None *MatchExpressions `yaml:"none,omitempty"`
	Expr *string           `yaml:"expr,omitempty"`
}

type MatchExpressions struct {
	Of []Match `yaml:"of"`
}

type DefineResourcesFunc func(builder *ResourceBuilder) []ResourceDefinition

func GeneratePolicies(serviceName string, defineResources DefineResourcesFunc, outputDir string) error {
	if defineResources == nil {
		return fmt.Errorf("define resources function is required")
	}

	builder := NewResourceBuilder(serviceName)
	resources := defineResources(builder)

	for _, resource := range resources {
		if resource.Resource == "" {
			return fmt.Errorf("invalid resource name")
		}

		policy := CerbosPolicy{
			APIVersion: "api.cerbos.dev/v1",
			ResourcePolicy: ResourcePolicy{
				Version:  "default",
				Resource: resource.Resource,
				Rules:    make([]Rule, 0, len(resource.Actions)),
			},
		}

		for _, action := range resource.Actions {
			rule := Rule{
				Actions:   []string{action.Action},
				Effect:    "EFFECT_ALLOW",
				Roles:     action.Roles,
				Condition: action.Condition,
			}
			policy.ResourcePolicy.Rules = append(policy.ResourcePolicy.Rules, rule)
		}

		filename := strings.ReplaceAll(resource.Resource, ":", "_")
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.yaml", filename))

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		data, err := yaml.Marshal(policy)
		if err != nil {
			return fmt.Errorf("failed to marshal policy: %w", err)
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	return nil
}
