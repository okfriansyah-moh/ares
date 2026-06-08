package importer

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/markdown"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

const claudeSourceMarker = "<!-- ars:source .ai/ -->"

// ClaudeImporter imports from CLAUDE.md.
type ClaudeImporter struct{}

func (c *ClaudeImporter) Source() string {
	return "claude"
}

func (c *ClaudeImporter) Import(root string) (*arslib.Repository, error) {
	path, err := safepath.Join(root, "CLAUDE.md")
	if err != nil {
		return nil, fmt.Errorf("import claude: %w", err)
	}

	data, err := safepath.ReadFile(root, "CLAUDE.md")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("import claude: CLAUDE.md not found at %s", path)
		}
		return nil, fmt.Errorf("import claude: %w", err)
	}

	sections, err := markdown.ExtractSections(data)
	if err != nil {
		return nil, fmt.Errorf("import claude: %w", err)
	}

	projectName := inferProjectName(data, sections)
	if strings.Contains(string(data), claudeSourceMarker) {
		sections = normalizeClaudeGeneratedSections(sections, projectName)
	} else {
		slog.Warn("import claude: source marker missing", "path", path)
	}

	return sectionsToRepository(sections, projectName), nil
}

func normalizeClaudeGeneratedSections(sections []markdown.Section, projectName string) []markdown.Section {
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

func isProviderIndexHeading(heading string) bool {
	h := strings.ToLower(strings.TrimSpace(heading))
	return h == "claude skills" || h == "codex skills" || h == "github copilot files"
}

func isClaudeAgentBodyHeading(heading string) bool {
	switch strings.ToLower(strings.TrimSpace(heading)) {
	case "role", "responsibilities", "uses", "boundaries":
		return true
	default:
		return false
	}
}
