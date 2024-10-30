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
	Actions []string `yaml:"actions"`
	Effect  string   `yaml:"effect"`
	Roles   []string `yaml:"roles"`
}

type DefineResourcesFunc func(builder *ResourceBuilder) []ResourceDefinition

func GeneratePolicies(serviceName string, defineResources DefineResourcesFunc, outputDir string) error {
	builder := NewResourceBuilder(serviceName)
	resources := defineResources(builder)

	for _, resource := range resources {
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
				Actions: []string{action.Action},
				Effect:  "EFFECT_ALLOW",
				Roles:   action.Roles,
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
