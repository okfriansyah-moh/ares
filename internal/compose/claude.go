package compose

import (
	"fmt"

	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
)

// ClaudeComposer writes CLAUDE.md at the repository root.
type ClaudeComposer struct{}

func (c *ClaudeComposer) Target() string {
	return "claude"
}

func (c *ClaudeComposer) Compose(root string, repo *arslib.Repository) error {
	if repo == nil {
		return fmt.Errorf("compose claude: repository is nil")
	}

	if err := validateAgentIDs(root, repo.Agents); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}

	if _, err := safepath.Join(root, "CLAUDE.md"); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}

	content := buildMarkdownOutput(composerFormat{
		agentSection: claudeAgentSection,
	}, repo)

	if err := safepath.WriteFile(root, "CLAUDE.md", []byte(content), 0o644); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}

	return nil
}
