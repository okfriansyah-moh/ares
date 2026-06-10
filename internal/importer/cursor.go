package importer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/content"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

const (
	cursorTypeAgentRequested = "agent-requested"
	cursorTypeAlways         = "always"
	cursorARSTypeAgent       = "agent"
	cursorARSTypeInstruction = "instruction"
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
			if !content.HasBody(body) {
				continue
			}
			repo.Agents = append(repo.Agents, arslib.Agent{
				ID:        id,
				Path:      filepath.ToSlash(filepath.Join(relBase, "agents", id, "AGENT.md")),
				Content:   body,
				SkillRefs: extractSkillRefs(body),
			})
		case cursorTypeAlways:
			if !content.HasBody(body) {
				continue
			}
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
		return "", cleanImportedMarkdownBody(text)
	}

	meta := map[string]string{}

	for _, line := range strings.Split(parts[1], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		meta[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	if kind, ok := meta["arsType"]; ok {
		switch strings.Trim(kind, `"'`) {
		case cursorARSTypeAgent:
			ruleType = cursorTypeAgentRequested
		case cursorARSTypeInstruction:
			ruleType = cursorTypeAlways
		}
	}

	if ruleType == "" {
		if typ, ok := meta["type"]; ok {
			ruleType = strings.TrimSpace(typ)
		}
	}

	if ruleType == "" {
		if always, ok := meta["alwaysApply"]; ok {
			switch strings.Trim(strings.ToLower(always), `"'`) {
			case "true":
				ruleType = cursorTypeAlways
			case "false":
				ruleType = cursorTypeAgentRequested
			}
		}
	}

	return ruleType, cleanImportedMarkdownBody(parts[2])
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
