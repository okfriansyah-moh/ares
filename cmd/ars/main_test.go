package main

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	"github.com/ars-standard/ars/internal/safepath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeTestCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestInit_CreatesAIDir(t *testing.T) {
	root := t.TempDir()

	_, err := executeTestCommand(t, "init", "--root", root)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(root, ".ai", "manifest.yaml"))
}

func TestValidate_ExitZeroOnValid(t *testing.T) {
	root := t.TempDir()
	_, err := executeTestCommand(t, "init", "--root", root)
	require.NoError(t, err)

	_, err = executeTestCommand(t, "validate", "--root", root)
	require.NoError(t, err)
}

func TestValidate_ExitOneOnError(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".ai/manifest.yaml", "version: \"2.0\"\nproject:\n  name: demo\n")
	writeFile(t, root, ".ai/agents/planner/AGENT.md", "## Role\nPlans work.\n")

	_, err := executeTestCommand(t, "validate", "--root", root)
	require.Error(t, err)
	assert.True(t, errors.Is(err, errValidationFailed))
}

func TestCompose_CursorTarget(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".ai/manifest.yaml", "version: \"2.0\"\nproject:\n  name: demo\n")
	writeFile(t, root, ".ai/instructions/repo-rules.md", "Repository rules.\n")

	_, err := executeTestCommand(t, "compose", "--target", "cursor", "--root", root)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(root, ".cursor", "rules", "repo-rules.mdc"))
}

func TestImport_GitHubSource(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".github/copilot-instructions.md", `# demo

## Agent: planner

Plans work.
`)

	_, err := executeTestCommand(t, "import", "github", "--root", root)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(root, ".ai", "manifest.yaml"))
	require.FileExists(t, filepath.Join(root, ".ai", "agents", "planner", "AGENT.md"))
}

func TestVersion_Flag(t *testing.T) {
	out, err := executeTestCommand(t, "--version")
	require.NoError(t, err)
	assert.Contains(t, out, "ars vdev")
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	require.NoError(t, safepath.WriteFile(root, rel, []byte(content), 0o644))
}
