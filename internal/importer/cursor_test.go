package importer

import (
	"path/filepath"
	"testing"

	"github.com/okfriansyah-moh/ares/internal/compose"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorImporter_MissingRulesDir(t *testing.T) {
	root := t.TempDir()

	_, err := (&CursorImporter{}).Import(root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "import cursor")
}

func writeCursorRule(t *testing.T, root, name, content string) {
	t.Helper()
	require.NoError(t, safepath.WriteFile(root, filepath.ToSlash(filepath.Join(".cursor", "rules", name)), []byte(content), 0o644))
}

func TestCursorImporter_AgentRule(t *testing.T) {
	root := t.TempDir()
	writeCursorRule(t, root, "planner.mdc", "---\ndescription: \"Use when planning\"\nalwaysApply: false\narsType: agent\n---\n<!-- ars:source .ai/ -->\n## Role\nPlans work.\n\n## Uses\n- skills/task-implementation/SKILL.md\n")

	repo, err := (&CursorImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Agents, 1)
	assert.Equal(t, "planner", repo.Agents[0].ID)
	assert.Contains(t, repo.Agents[0].Content, "Plans work.")
	assert.Equal(t, []string{"skills/task-implementation/SKILL.md"}, repo.Agents[0].SkillRefs)
}

func TestCursorImporter_InstructionRule(t *testing.T) {
	root := t.TempDir()
	writeCursorRule(t, root, "repo-rules.mdc", "---\ndescription: \"Repository guidance\"\nalwaysApply: true\narsType: instruction\n---\nRepository rules.\n")

	repo, err := (&CursorImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Instructions, 1)
	assert.Equal(t, "repo-rules", repo.Instructions[0].ID)
	assert.Equal(t, "Repository rules.", repo.Instructions[0].Content)
}

func TestCursorImporter_EmptyRulesDir(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, safepath.MkdirAll(root, ".cursor/rules", 0o755))

	repo, err := (&CursorImporter{}).Import(root)
	require.NoError(t, err)
	assert.Empty(t, repo.Agents)
	assert.Empty(t, repo.Instructions)
}

func TestCursorImporter_FrontMatterStripped(t *testing.T) {
	root := t.TempDir()
	writeCursorRule(t, root, "planner.mdc", "---\ndescription: \"Use when planning\"\nalwaysApply: false\narsType: agent\n---\n## Role\nPlans work.\n")

	repo, err := (&CursorImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Agents, 1)
	assert.NotContains(t, repo.Agents[0].Content, "alwaysApply: false")
	assert.NotContains(t, repo.Agents[0].Content, "---")
}

func TestCursorImporter_LegacyTypeStillSupported(t *testing.T) {
	root := t.TempDir()
	writeCursorRule(t, root, "planner.mdc", "---\ntype: agent-requested\n---\n## Role\nPlans work.\n")

	repo, err := (&CursorImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Agents, 1)
	assert.Equal(t, "planner", repo.Agents[0].ID)
}

func TestCursorImporter_ComposeRoundTripAgentSections(t *testing.T) {
	root := t.TempDir()
	source := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Agents: []arslib.Agent{{
			ID:      "planner",
			Content: "## Role\nPlans work.\n",
		}},
	}

	require.NoError(t, (&compose.CursorComposer{}).Compose(root, source))

	repo, err := (&CursorImporter{}).Import(root)
	require.NoError(t, err)
	require.Len(t, repo.Agents, 1)
	assert.Equal(t, "planner", repo.Agents[0].ID)
	assert.Contains(t, repo.Agents[0].Content, "Plans work.")
}

func TestRegistry_SourcesIncludesCursorAndClaude(t *testing.T) {
	sources := DefaultRegistry.Sources()
	assert.Contains(t, sources, "cursor")
	assert.Contains(t, sources, "claude")
	assert.Contains(t, sources, "codex")
}
