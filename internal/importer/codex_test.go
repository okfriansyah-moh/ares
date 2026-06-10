package importer

import (
	"testing"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeCodexFile(t *testing.T, root, content string) {
	t.Helper()
	require.NoError(t, safepath.WriteFile(root, "AGENTS.md", []byte(content), 0o644))
}

func TestCodexImporter_WithMarker(t *testing.T) {
	root := t.TempDir()
	writeCodexFile(t, root, `<!-- ars:source .ai/ -->
# demo

## planner

## Role
Plans work.

## Uses
- skills/plan-management/SKILL.md
`)

	repo, err := (&CodexImporter{}).Import(root)
	require.NoError(t, err)
	assert.Equal(t, "demo", repo.Manifest.Project.Name)
	require.Len(t, repo.Agents, 1)
	assert.Equal(t, "planner", repo.Agents[0].ID)
	assert.Equal(t, []string{"skills/plan-management/SKILL.md"}, repo.Agents[0].SkillRefs)
}

func TestCodexImporter_PreservesReadableInstructionBoundaries(t *testing.T) {
	root := t.TempDir()
	writeCodexFile(t, root, `<!-- ars:source .ai/ -->
# demo

## Repository Instructions

<!-- ars:source .ai/instructions/stack.md -->
Backend: Go 1.26, modular monolith
Agent binary: Go 1.26, github.com/a2aproject/a2a-go/v2
Frontend: SvelteKit, TypeScript, TailwindCSS
`)

	repo, err := (&CodexImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Instructions, 1)

	content := repo.Instructions[0].Content
	assert.Contains(t, content, "monolith\nAgent binary")
	assert.Contains(t, content, "a2a-go/v2\nFrontend")
	assert.NotContains(t, content, "<!-- ars:source")
	assert.NotContains(t, content, "monolithAgent")
	assert.NotContains(t, content, "a2a-go/v2Frontend")
}

func TestCodexImporter_MissingFile(t *testing.T) {
	root := t.TempDir()

	_, err := (&CodexImporter{}).Import(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AGENTS.md not found")
}
