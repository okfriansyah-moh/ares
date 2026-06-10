package importer

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeClaudeFile(t *testing.T, root, content string) {
	t.Helper()
	require.NoError(t, safepath.WriteFile(root, "CLAUDE.md", []byte(content), 0o644))
}

func TestClaudeImporter_WithMarker(t *testing.T) {
	root := t.TempDir()
	writeClaudeFile(t, root, `<!-- ars:source .ai/ -->
# demo

## Repository Instructions

Repository rules.

## planner

## Role
Plans work.

## Uses
- skills/plan-management/SKILL.md
`)

	repo, err := (&ClaudeImporter{}).Import(root)
	require.NoError(t, err)
	assert.Equal(t, "demo", repo.Manifest.Project.Name)
	require.Len(t, repo.Instructions, 1)
	require.Len(t, repo.Agents, 1)
	assert.Equal(t, "planner", repo.Agents[0].ID)
	assert.Contains(t, repo.Agents[0].Content, "Plans work.")
	assert.Equal(t, []string{"skills/plan-management/SKILL.md"}, repo.Agents[0].SkillRefs)
}

func TestClaudeImporter_PreservesReadableInstructionBoundaries(t *testing.T) {
	root := t.TempDir()
	writeClaudeFile(t, root, `<!-- ars:source .ai/ -->
# demo

## Repository Instructions

<!-- ars:source .ai/instructions/security-invariants.md -->
API keys never in source/config/logs - CredentialRef = env var name only
os.Getenv() only in backend/internal/platform/config/config.go and agent/internal/config/config.go
Never log resolved credential values
`)

	repo, err := (&ClaudeImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Instructions, 1)

	content := repo.Instructions[0].Content
	assert.Contains(t, content, "only\nos.Getenv")
	assert.Contains(t, content, "config.go\nNever log")
	assert.NotContains(t, content, "<!-- ars:source")
	assert.NotContains(t, content, "onlyos.Getenv")
	assert.NotContains(t, content, "config.goNever")
}

func TestClaudeImporter_WithoutMarker(t *testing.T) {
	root := t.TempDir()
	writeClaudeFile(t, root, `# demo

## Agent: reviewer

Reviews changes.
`)

	var warnings bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&warnings, nil)))
	defer slog.SetDefault(originalLogger)

	repo, err := (&ClaudeImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Agents, 1)
	assert.Equal(t, "reviewer", repo.Agents[0].ID)
	assert.Contains(t, repo.Agents[0].Content, "Reviews changes.")
	assert.Contains(t, warnings.String(), "source marker missing")
}

func TestClaudeImporter_MissingFile(t *testing.T) {
	root := t.TempDir()

	_, err := (&ClaudeImporter{}).Import(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CLAUDE.md not found")
}
