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

type ResourceDefiner interface {
	DefineResources(builder *ResourceBuilder) []ResourceDefinition
}

func GeneratePolicies(definer ResourceDefiner, outputDir string) error {
	builder := NewResourceBuilder("")
	resources := definer.DefineResources(builder)

	for _, resource := range resources {
		policy := CerbosPolicy{
			APIVersion: "api.cerbos.dev/v1",
			ResourcePolicy: ResourcePolicy{
				Version:  "default",
				Resource: resource.GetResource(),
				Rules:    make([]Rule, 0, len(resource.GetActions())),
			},
		}

		for _, action := range resource.GetActions() {
			rule := Rule{
				Actions: []string{action.GetAction()},
				Effect:  "EFFECT_ALLOW",
				Roles:   action.GetRoles(),
			}
			policy.ResourcePolicy.Rules = append(policy.ResourcePolicy.Rules, rule)
		}

		filename := strings.ReplaceAll(resource.GetResource(), ":", "_")
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
