package validator

import (
	"testing"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/internal/scaffold"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validAgent = `## Role
Owns testing.

## Responsibilities
- Validate structure

## Uses
- skills/example/SKILL.md

## Boundaries
- No runtime execution
`

func writeValidTree(t *testing.T, root string) {
	t.Helper()
	require.NoError(t, scaffold.Run(scaffold.Options{Root: root}))

	require.NoError(t, safepath.WriteFile(root, ".ai/agents/planner/AGENT.md", []byte(validAgent), 0o644))
	require.NoError(t, safepath.WriteFile(root, ".ai/skills/example/SKILL.md", []byte("# Example\n"), 0o644))
}

func errorFindings(findings []arslib.Finding) []arslib.Finding {
	var out []arslib.Finding
	for _, f := range findings {
		if f.Level == arslib.Error {
			out = append(out, f)
		}
	}
	return out
}

func TestRun_ValidTree(t *testing.T) {
	root := t.TempDir()
	writeValidTree(t, root)

	findings, err := Run(root)
	require.NoError(t, err)
	assert.Empty(t, errorFindings(findings))
}

func TestRun_MissingManifest(t *testing.T) {
	root := t.TempDir()

	findings, err := Run(root)
	require.NoError(t, err)

	var manifestErrors []arslib.Finding
	for _, f := range findings {
		if f.Path == ".ai/manifest.yaml" && f.Level == arslib.Error {
			manifestErrors = append(manifestErrors, f)
		}
	}
	require.NotEmpty(t, manifestErrors)
}

func TestRun_MissingAgentSection(t *testing.T) {
	root := t.TempDir()
	writeValidTree(t, root)

	require.NoError(t, safepath.WriteFile(root, ".ai/agents/planner/AGENT.md", []byte("## Responsibilities\nOnly one section.\n"), 0o644))

	findings, err := Run(root)
	require.NoError(t, err)

	missing := map[string]bool{
		"Role":       false,
		"Uses":       false,
		"Boundaries": false,
	}
	for _, f := range findings {
		if f.Level == arslib.Error && f.Path == ".ai/agents/planner/AGENT.md" {
			for heading := range missing {
				if f.Message == "missing required section ## "+heading {
					missing[heading] = true
				}
			}
		}
	}

	assert.True(t, missing["Role"])
	assert.True(t, missing["Uses"])
	assert.True(t, missing["Boundaries"])
}

func TestRun_BrokenSkillRef(t *testing.T) {
	root := t.TempDir()
	writeValidTree(t, root)

	brokenAgent := `## Role
Owns testing.

## Responsibilities
- Validate structure

## Uses
- skills/foo/SKILL.md

## Boundaries
- No runtime execution
`
	require.NoError(t, safepath.WriteFile(root, ".ai/agents/planner/AGENT.md", []byte(brokenAgent), 0o644))

	findings, err := Run(root)
	require.NoError(t, err)

	found := false
	for _, f := range findings {
		if f.Level == arslib.Error && f.Path == ".ai/agents/planner/AGENT.md" && f.Message == "skill reference not found: skills/foo/SKILL.md" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestRun_DeterministicOrder(t *testing.T) {
	root := t.TempDir()
	writeValidTree(t, root)

	first, err := Run(root)
	require.NoError(t, err)

	second, err := Run(root)
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestRun_PathTraversalInSkillRef(t *testing.T) {
	root := t.TempDir()
	writeValidTree(t, root)

	traversalAgent := `## Role
Owns testing.

## Responsibilities
- Validate structure

## Uses
- ../../etc/passwd

## Boundaries
- No runtime execution
`
	require.NoError(t, safepath.WriteFile(root, ".ai/agents/planner/AGENT.md", []byte(traversalAgent), 0o644))

	findings, err := Run(root)
	require.NoError(t, err)

	found := false
	for _, f := range findings {
		if f.Level == arslib.Error && f.Path == ".ai/agents/planner/AGENT.md" && f.Message == "skill reference escapes repository root: ../../etc/passwd" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestLevelString(t *testing.T) {
	assert.Equal(t, "Error", levelString(arslib.Error))
}
