package importer

import (
	"fmt"
	"os"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/markdown"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

const codexSourceMarker = "<!-- ars:source .ai/ -->"

// CodexImporter imports from AGENTS.md.
type CodexImporter struct{}

func (c *CodexImporter) Source() string {
	return "codex"
}

func (c *CodexImporter) Import(root string) (*arslib.Repository, error) {
	path, err := safepath.Join(root, "AGENTS.md")
	if err != nil {
		return nil, fmt.Errorf("import codex: %w", err)
	}

	data, err := safepath.ReadFile(root, "AGENTS.md")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("import codex: AGENTS.md not found at %s", path)
		}
		return nil, fmt.Errorf("import codex: %w", err)
	}

	sections, err := markdown.ExtractSections(data)
	if err != nil {
		return nil, fmt.Errorf("import codex: %w", err)
	}

	projectName := inferProjectName(data, sections)
	if strings.Contains(string(data), codexSourceMarker) {
		sections = normalizeCodexGeneratedSections(sections, projectName)
	}

	repo := sectionsToRepository(sections, projectName)
	return repo, nil
}

func normalizeCodexGeneratedSections(sections []markdown.Section, projectName string) []markdown.Section {
	out := append([]markdown.Section(nil), sections...)
	for i := range out {
		heading := strings.TrimSpace(out[i].Heading)
		if heading == "" ||
			out[i].Level != 2 ||
			strings.EqualFold(heading, projectName) ||
			isTopLevelInstruction(heading) ||
			isProviderIndexHeading(heading) ||
			isClaudeAgentBodyHeading(heading) ||
			ClassifySection(heading) != classInstruction {
			continue
		}
		out[i].Heading = "Agent: " + heading
	}
	return out
}
