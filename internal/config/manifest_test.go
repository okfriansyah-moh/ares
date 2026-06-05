package config

import (
	"errors"
	"os"
	"testing"

	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidManifest(t *testing.T) {
	root := t.TempDir()

	manifestYAML := `version: "2.0"
project:
  name: ares
  description: AI Repository Standard
  repository: https://github.com/ars-standard/ars
defaults:
  agent: architect
`
	require.NoError(t, safepath.WriteFile(root, ".ai/manifest.yaml", []byte(manifestYAML), 0o644))

	got, err := Load(root)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "2.0", got.Version)
	assert.Equal(t, arslib.Project{
		Name:        "ares",
		Description: "AI Repository Standard",
		Repository:  "https://github.com/ars-standard/ars",
	}, got.Project)
	assert.Equal(t, arslib.Defaults{Agent: "architect"}, got.Defaults)
}

func TestLoad_MissingFile(t *testing.T) {
	root := t.TempDir()

	_, err := Load(root)
	require.Error(t, err)

	var pathErr *os.PathError
	require.ErrorAs(t, err, &pathErr)
}

func TestLoad_InvalidYAML(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, safepath.WriteFile(root, ".ai/manifest.yaml", []byte("version: [\n"), 0o644))

	_, err := Load(root)
	require.Error(t, err)
	assert.NotPanics(t, func() { _ = err.Error() })
}

func TestLoad_MissingProjectName(t *testing.T) {
	m := &arslib.Manifest{
		Version: "2.0",
		Project: arslib.Project{Name: ""},
	}

	err := Validate(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project.name")
}

func TestWrite_Roundtrip(t *testing.T) {
	root := t.TempDir()
	original := &arslib.Manifest{
		Version: "2.0",
		Project: arslib.Project{
			Name:        "roundtrip",
			Description: "test repo",
			Repository:  "https://example.com/ares",
		},
		Defaults: arslib.Defaults{Agent: "planner"},
	}

	require.NoError(t, Write(root, original))

	got, err := Load(root)
	require.NoError(t, err)
	assert.Equal(t, original, got)
}

func TestLoad_PathTraversal(t *testing.T) {
	root := t.TempDir()

	_, err := safepath.Join(root, "..", "..", "etc", "passwd")
	require.Error(t, err)
	assert.True(t, errors.Is(err, safepath.ErrPathEscape))
}
