package scaffold

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/okfriansyah-moh/ares/internal/config"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_FreshDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "fresh-repo")

	err := Run(Options{Root: root})
	require.NoError(t, err)

	manifestPath := filepath.Join(root, ".ai", "manifest.yaml")
	readmePath := filepath.Join(root, ".ai", "instructions", "README.md")

	require.FileExists(t, manifestPath)
	require.FileExists(t, readmePath)

	m, err := config.Load(root)
	require.NoError(t, err)
	assert.Equal(t, "2.0", m.Version)
	assert.Equal(t, "fresh-repo", m.Project.Name)
}

func TestRun_AlreadyExists(t *testing.T) {
	root := t.TempDir()

	original := []byte("version: \"2.0\"\nproject:\n  name: keep-me\n")
	require.NoError(t, safepath.WriteFile(root, ".ai/manifest.yaml", original, 0o644))

	err := Run(Options{Root: root})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAlreadyInitialised))

	got, readErr := safepath.ReadFile(root, ".ai/manifest.yaml")
	require.NoError(t, readErr)
	assert.Equal(t, original, got)
}

func TestRun_Force(t *testing.T) {
	root := filepath.Join(t.TempDir(), "force-repo")
	require.NoError(t, Run(Options{Root: root}))

	require.NoError(t, safepath.WriteFile(root, ".ai/manifest.yaml", []byte("version: \"2.0\"\nproject:\n  name: stale\n"), 0o644))

	err := Run(Options{Root: root, Force: true})
	require.NoError(t, err)

	m, err := config.Load(root)
	require.NoError(t, err)
	assert.Equal(t, "force-repo", m.Project.Name)
}

func TestRun_ProjectNameInferred(t *testing.T) {
	root := filepath.Join(t.TempDir(), "my-project")

	require.NoError(t, Run(Options{Root: root}))

	m, err := config.Load(root)
	require.NoError(t, err)
	assert.Equal(t, filepath.Base(root), m.Project.Name)
	assert.Equal(t, "my-project", m.Project.Name)
}
