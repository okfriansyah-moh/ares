package compose

import (
	"path/filepath"
	"testing"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeComposer_OutputFilename(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
	}

	require.NoError(t, (&ClaudeComposer{}).Compose(root, repo))

	path := filepath.Join(root, "CLAUDE.md")
	require.FileExists(t, path)
	assert.Equal(t, root, filepath.Dir(path))
}

func TestClaudeComposer_SourceMarker(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
	}

	require.NoError(t, (&ClaudeComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, "CLAUDE.md")
	require.NoError(t, err)
	assert.Contains(t, string(data), "<!-- ars:source .ai/ -->")
}

func TestClaudeComposer_Idempotent(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Instructions: []arslib.Instruction{{
			ID:      "rules",
			Content: "Repository rules.",
		}},
		Agents: []arslib.Agent{{
			ID:      "Planner",
			Content: "Plans work.",
		}},
	}

	composer := &ClaudeComposer{}
	require.NoError(t, composer.Compose(root, repo))

	first, err := safepath.ReadFile(root, "CLAUDE.md")
	require.NoError(t, err)

	require.NoError(t, composer.Compose(root, repo))
	second, err := safepath.ReadFile(root, "CLAUDE.md")
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestClaudeComposer_PathTraversal(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Agents: []arslib.Agent{{
			ID:      "../evil",
			Content: "malicious",
		}},
	}

	err := (&ClaudeComposer{}).Compose(root, repo)
	require.Error(t, err)

	exists, err := safepath.Exists(root, "CLAUDE.md")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = safepath.Exists(root, "evil.md")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestClaudeAgentSection_LowercaseHeading(t *testing.T) {
	section := claudeAgentSection(arslib.Agent{
		ID:      "Planner",
		Content: "Plans.",
	}, nil)
	assert.Contains(t, section, "## planner\n")
}
