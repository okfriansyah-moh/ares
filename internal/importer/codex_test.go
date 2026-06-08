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

func TestCodexImporter_MissingFile(t *testing.T) {
	root := t.TempDir()

	_, err := (&CodexImporter{}).Import(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AGENTS.md not found")
}
