package importer

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/ars-standard/ars/internal/safepath"
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
`)

	repo, err := (&ClaudeImporter{}).Import(root)
	require.NoError(t, err)
	assert.Equal(t, "demo", repo.Manifest.Project.Name)
	require.Len(t, repo.Instructions, 1)
	require.Len(t, repo.Agents, 1)
	assert.Equal(t, "planner", repo.Agents[0].ID)
	assert.Contains(t, repo.Agents[0].Content, "Plans work.")
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
