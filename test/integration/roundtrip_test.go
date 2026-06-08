package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var arsBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ars-integration-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	name := "ars"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	arsBin = filepath.Join(dir, name)

	build := exec.Command("go", "build", "-trimpath", "-o", arsBin, "./cmd/ars")
	build.Dir = repoRoot()
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		panic(string(out))
	}

	os.Exit(m.Run())
}

func TestRoundTrip_CursorComposeThenImport(t *testing.T) {
	source := sampleRepo(t)
	runARS(t, source, "compose", "--target", "cursor", "--root", source)

	importRoot := t.TempDir()
	copyDir(t, filepath.Join(source, ".cursor"), filepath.Join(importRoot, ".cursor"))
	runARS(t, importRoot, "import", "cursor", "--root", importRoot)

	assert.Contains(t, readFile(t, importRoot, ".ai/agents/planner/AGENT.md"), "Plans implementation tasks.")
	assert.Contains(t, readFile(t, importRoot, ".ai/instructions/repo.md"), "Keep changes scoped.")
}

func TestRoundTrip_CopilotComposeThenImport(t *testing.T) {
	source := sampleRepo(t)
	runARS(t, source, "compose", "--target", "copilot", "--root", source)

	importRoot := t.TempDir()
	copyDir(t, filepath.Join(source, ".github"), filepath.Join(importRoot, ".github"))
	runARS(t, importRoot, "import", "github", "--root", importRoot)

	assert.Contains(t, readFile(t, importRoot, ".ai/agents/planner/AGENT.md"), "Plans implementation tasks.")
	assert.Contains(t, readFile(t, importRoot, ".ai/instructions/repository-instructions.md"), "Keep changes scoped.")
	assert.Contains(t, readFile(t, importRoot, ".ai/prompts/review-task.md"), "Review Task")
}

func TestRoundTrip_ClaudeComposeThenImport(t *testing.T) {
	source := sampleRepo(t)
	runARS(t, source, "compose", "--target", "claude", "--root", source)

	importRoot := t.TempDir()
	copyFile(t, filepath.Join(source, "CLAUDE.md"), filepath.Join(importRoot, "CLAUDE.md"))
	runARS(t, importRoot, "import", "claude", "--root", importRoot)

	assert.Contains(t, readFile(t, importRoot, ".ai/agents/planner/AGENT.md"), "Plans implementation tasks.")
	assert.Contains(t, readFile(t, importRoot, ".ai/instructions/repository-instructions.md"), "Keep changes scoped.")
}

func TestRoundTrip_CodexComposeThenImport(t *testing.T) {
	source := sampleRepo(t)
	runARS(t, source, "compose", "--target", "codex", "--root", source)

	importRoot := t.TempDir()
	copyFile(t, filepath.Join(source, "AGENTS.md"), filepath.Join(importRoot, "AGENTS.md"))
	runARS(t, importRoot, "import", "codex", "--root", importRoot)

	assert.Contains(t, readFile(t, importRoot, ".ai/agents/planner/AGENT.md"), "Plans implementation tasks.")
	assert.Contains(t, readFile(t, importRoot, ".ai/instructions/repository-instructions.md"), "Keep changes scoped.")
}

func TestRoundTrip_AllTargets(t *testing.T) {
	root := sampleRepo(t)

	for _, target := range []string{"cursor", "copilot", "claude", "codex"} {
		runARS(t, root, "compose", "--target", target, "--root", root)
	}

	require.FileExists(t, filepath.Join(root, ".cursor", "rules", "repo.mdc"))
	require.FileExists(t, filepath.Join(root, ".github", "copilot-instructions.md"))
	require.FileExists(t, filepath.Join(root, "CLAUDE.md"))
	require.FileExists(t, filepath.Join(root, "AGENTS.md"))

	assert.Contains(t, readFile(t, root, ".cursor/rules/repo.mdc"), "<!-- ars:source .ai/")
	assert.Contains(t, readFile(t, root, ".github/copilot-instructions.md"), "Source: .ai/")
	assert.Contains(t, readFile(t, root, "CLAUDE.md"), "<!-- ars:source .ai/ -->")
	assert.Contains(t, readFile(t, root, "AGENTS.md"), "<!-- ars:source .ai/ -->")
}

func TestRoundTrip_EmptyRepo(t *testing.T) {
	root := t.TempDir()

	runARS(t, root, "init", "--root", root)
	out := runARS(t, root, "validate", "--root", root)

	assert.NotContains(t, out, "Error")
	require.FileExists(t, filepath.Join(root, ".ai", "manifest.yaml"))
}

func TestRoundTrip_AddAgentComposeValidate(t *testing.T) {
	root := t.TempDir()
	runARS(t, root, "init", "--root", root)
	writeFile(t, root, ".ai/agents/planner/AGENT.md", plannerAgent)
	writeFile(t, root, ".ai/skills/task-implementation/SKILL.md", "# Task Implementation\n")

	runARS(t, root, "compose", "--target", "cursor", "--root", root)
	out := runARS(t, root, "validate", "--root", root)

	require.FileExists(t, filepath.Join(root, ".cursor", "rules", "planner.mdc"))
	assert.NotContains(t, out, "Error")
}

func sampleRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, ".ai/manifest.yaml", `version: "2.0"
project:
  name: demo
  description: Integration fixture
defaults:
  agent: planner
`)
	writeFile(t, root, ".ai/instructions/repo.md", "Keep changes scoped.\n")
	writeFile(t, root, ".ai/agents/planner/AGENT.md", plannerAgent)
	writeFile(t, root, ".ai/skills/task-implementation/SKILL.md", "# Task Implementation\n")
	writeFile(t, root, ".ai/prompts/review-task.md", "# Review Task\n")
	return root
}

const plannerAgent = `## Role
Plans implementation tasks.

## Responsibilities
- Keep work scoped.

## Uses
- skills/task-implementation/SKILL.md

## Boundaries
- Does not implement future tasks.
`

func runARS(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(arsBin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	require.NoError(t, err)
	return string(data)
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	require.NoError(t, filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		copyFile(t, path, target)
		return nil
	}))
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o755))
	require.NoError(t, os.WriteFile(dst, data, 0o644))
}

func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot resolve repository root")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if strings.TrimSpace(root) == "" {
		panic("empty repository root")
	}
	return root
}
