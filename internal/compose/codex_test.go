package compose

import (
	"path/filepath"
	"testing"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodexComposer_OutputFilename(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
	}

	require.NoError(t, (&CodexComposer{}).Compose(root, repo))

	path := filepath.Join(root, "AGENTS.md")
	require.FileExists(t, path)
	assert.Equal(t, root, filepath.Dir(path))
}

func TestCodexComposer_SourceMarker(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
	}

	require.NoError(t, (&CodexComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, "AGENTS.md")
	require.NoError(t, err)
	assert.Contains(t, string(data), "<!-- ars:source .ai/ -->")
}

func TestCodexComposer_Idempotent(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Instructions: []arslib.Instruction{{
			ID:      "rules",
			Content: "Repository rules.",
		}},
		Agents: []arslib.Agent{{
			ID:      "planner",
			Content: "Plans work.",
		}},
	}

	composer := &CodexComposer{}
	require.NoError(t, composer.Compose(root, repo))

	first, err := safepath.ReadFile(root, "AGENTS.md")
	require.NoError(t, err)

	require.NoError(t, composer.Compose(root, repo))
	second, err := safepath.ReadFile(root, "AGENTS.md")
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestCodexAgentSection_YAMLHeader(t *testing.T) {
	section := codexAgentSection(arslib.Agent{
		ID:      "planner",
		Content: "Plans.",
	}, nil)
	assert.Contains(t, section, "---\nagent: planner\n---")
}
