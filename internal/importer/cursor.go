package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
)

const (
	cursorTypeAgentRequested = "agent-requested"
	cursorTypeAlways         = "always"
)

// CursorImporter imports from .cursor/rules/*.mdc.
type CursorImporter struct{}

func (c *CursorImporter) Source() string {
	return "cursor"
}

func (c *CursorImporter) Import(root string) (*arslib.Repository, error) {
	if _, err := safepath.Join(root, ".cursor", "rules"); err != nil {
		return nil, fmt.Errorf("import cursor: %w", err)
	}

	entries, err := safepath.ReadDir(root, ".cursor/rules")
	if err != nil {
		if os.IsNotExist(err) {
			return emptyImportedRepository("imported-project"), nil
		}
		return nil, fmt.Errorf("import cursor: %w", err)
	}

	repo := emptyImportedRepository("imported-project")
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".mdc" {
			continue
		}

		rel := filepath.ToSlash(filepath.Join(".cursor", "rules", entry.Name()))
		data, err := safepath.ReadFile(root, rel)
		if err != nil {
			return nil, fmt.Errorf("import cursor: %w", err)
		}

		ruleType, body := parseCursorRule(data)
		id := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		relBase := filepath.Join(".ai")

		switch ruleType {
		case cursorTypeAgentRequested:
			repo.Agents = append(repo.Agents, arslib.Agent{
				ID:      id,
				Path:    filepath.ToSlash(filepath.Join(relBase, "agents", id, "AGENT.md")),
				Content: body,
			})
		case cursorTypeAlways:
			repo.Instructions = append(repo.Instructions, arslib.Instruction{
				ID:      id,
				Path:    filepath.ToSlash(filepath.Join(relBase, "instructions", id+".md")),
				Content: body,
			})
		}
	}

	return repo, nil
}

func parseCursorRule(data []byte) (ruleType string, body string) {
	text := string(data)
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 || strings.TrimSpace(parts[0]) != "" {
		return "", strings.TrimSpace(text)
	}

	for _, line := range strings.Split(parts[1], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.TrimSpace(key) == "type" {
			ruleType = strings.TrimSpace(value)
			break
		}
	}

	return ruleType, strings.TrimSpace(parts[2])
}

func emptyImportedRepository(projectName string) *arslib.Repository {
	return &arslib.Repository{
		Manifest: arslib.Manifest{
			Version: "2.0",
			Project: arslib.Project{
				Name: projectName,
			},
		},
	}
}
