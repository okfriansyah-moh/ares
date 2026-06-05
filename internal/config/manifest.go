package config

import (
	"fmt"

	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
	"gopkg.in/yaml.v3"
)

const maxYAMLDepth = 8

// Load reads and parses .ai/manifest.yaml under root and validates it.
func Load(root string) (*arslib.Manifest, error) {
	data, err := safepath.ReadFile(root, ".ai/manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	var m arslib.Manifest
	if err := unmarshalYAML(data, &m); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	if err := Validate(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

// Validate checks required manifest fields.
func Validate(m *arslib.Manifest) error {
	if m == nil {
		return fmt.Errorf("config: manifest is nil")
	}
	if m.Version == "" {
		return fmt.Errorf("config: version is required")
	}
	if m.Project.Name == "" {
		return fmt.Errorf("config: project.name is required")
	}
	return nil
}

// Write marshals m to YAML and writes .ai/manifest.yaml under root atomically.
func Write(root string, m *arslib.Manifest) error {
	if err := Validate(m); err != nil {
		return err
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if err := safepath.WriteFile(root, ".ai/manifest.yaml", data, 0o644); err != nil {
		return fmt.Errorf("config: %w", err)
	}

	return nil
}

func unmarshalYAML(data []byte, v any) error {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return err
	}
	if yamlNodeDepth(&node) > maxYAMLDepth {
		return fmt.Errorf("yaml document exceeds maximum depth of %d", maxYAMLDepth)
	}
	return node.Decode(v)
}

func yamlNodeDepth(n *yaml.Node) int {
	if n == nil || len(n.Content) == 0 {
		return 1
	}
	maxChild := 0
	for _, child := range n.Content {
		if d := yamlNodeDepth(child); d > maxChild {
			maxChild = d
		}
	}
	return maxChild + 1
}
