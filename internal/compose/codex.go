package compose

import (
	"fmt"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

// CodexComposer writes AGENTS.md at the repository root.
type CodexComposer struct{}

func (c *CodexComposer) Target() string {
	return "codex"
}

func (c *CodexComposer) Compose(root string, repo *arslib.Repository) error {
	if repo == nil {
		return fmt.Errorf("compose codex: repository is nil")
	}

	if err := validateAgentIDs(root, repo.Agents); err != nil {
		return fmt.Errorf("compose codex: %w", err)
	}

	if _, err := safepath.Join(root, "AGENTS.md"); err != nil {
		return fmt.Errorf("compose codex: %w", err)
	}

	content := buildMarkdownOutput(composerFormat{
		agentSection: codexAgentSection,
	}, repo)

	if err := safepath.WriteFile(root, "AGENTS.md", []byte(content), 0o644); err != nil {
		return fmt.Errorf("compose codex: %w", err)
	}

	return nil
}
