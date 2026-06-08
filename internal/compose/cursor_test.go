package compose

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorComposer_BasicOutput(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{
			Project: arslib.Project{Name: "demo"},
		},
		Instructions: []arslib.Instruction{{
			ID:      "repo-rules",
			Content: "## Scope\nRepository rules.\n",
		}},
		Agents: []arslib.Agent{{
			ID:      "planner",
			Content: "## Role\nPlans work.\n",
		}},
	}

	require.NoError(t, (&CursorComposer{}).Compose(root, repo))

	rulePath := filepath.Join(root, ".cursor", "rules", "repo-rules.mdc")
	agentRulePath := filepath.Join(root, ".cursor", "rules", "planner.mdc")

	require.FileExists(t, rulePath)
	require.FileExists(t, agentRulePath)

	ruleData, err := safepath.ReadFile(root, ".cursor/rules/repo-rules.mdc")
	require.NoError(t, err)
	assert.Contains(t, string(ruleData), "type: always")
	assert.Contains(t, string(ruleData), "<!-- project: demo -->")
	assert.Contains(t, string(ruleData), "Repository rules.")

	agentData, err := safepath.ReadFile(root, ".cursor/rules/planner.mdc")
	require.NoError(t, err)
	assert.Contains(t, string(agentData), "type: agent-requested")
	assert.Contains(t, string(agentData), "Plans work.")
}

func TestCursorComposer_SkillInlined(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Agents: []arslib.Agent{{
			ID:        "architect",
			Content:   "## Role\nOwns architecture.\n",
			SkillRefs: []string{"skills/architecture-management/SKILL.md"},
		}},
		Skills: []arslib.Skill{{
			ID:      "architecture-management",
			Content: "## Skill\nArchitecture guidance.\n",
		}},
	}

	require.NoError(t, (&CursorComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, ".cursor/rules/architect.mdc")
	require.NoError(t, err)
	body := string(data)
	assert.Contains(t, body, "Architecture guidance.")
	assert.Contains(t, body, "### Context: architecture-management")
}

func TestCursorComposer_NoPrompts(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Instructions: []arslib.Instruction{{
			ID:      "repo-rules",
			Content: "rules",
		}},
	}

	require.NoError(t, (&CursorComposer{}).Compose(root, repo))

	exists, err := safepath.Exists(root, ".cursor/prompts")
	require.NoError(t, err)
	require.True(t, exists)

	entries, err := safepath.ReadDir(root, ".cursor/prompts")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestCursorComposer_Idempotent(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Instructions: []arslib.Instruction{{
			ID:      "repo-rules",
			Content: "rules",
		}},
		Agents: []arslib.Agent{{
			ID:      "planner",
			Content: "agent",
		}},
		Prompts: []arslib.Prompt{{
			ID:      "plan",
			Content: "prompt body",
		}},
	}

	composer := &CursorComposer{}
	require.NoError(t, composer.Compose(root, repo))
	first := treeChecksum(t, root, ".cursor")

	require.NoError(t, composer.Compose(root, repo))
	second := treeChecksum(t, root, ".cursor")

	assert.Equal(t, first, second)
}

func TestCursorComposer_PathTraversal(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Agents: []arslib.Agent{{
			ID:      "../evil",
			Content: "malicious",
		}},
	}

	require.NoError(t, (&CursorComposer{}).Compose(root, repo))

	entries, err := safepath.ReadDir(root, ".cursor/rules")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "evil.mdc", entries[0].Name())

	exists, err := safepath.Exists(root, "evil.mdc")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCompose_UnknownTarget(t *testing.T) {
	err := Compose(t.TempDir(), "unknown", &arslib.Repository{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownTarget)
}

func TestCursorRuleHeader(t *testing.T) {
	assert.Equal(t, "---\ntype: always\n---\n", cursorRuleHeader("always"))
}

func treeChecksum(t *testing.T, root, relDir string) string {
	t.Helper()
	h := sha256.New()
	dir, err := safepath.Join(root, relDir)
	require.NoError(t, err)
	err = safepath.WalkDir(root, relDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := safepath.ReadFile(root, filepath.ToSlash(filepath.Join(relDir, rel)))
		if err != nil {
			return err
		}
		h.Write([]byte(filepath.ToSlash(rel)))
		h.Write(data)
		return nil
	})
	require.NoError(t, err)
	return hex.EncodeToString(h.Sum(nil))
}

func TestRegistry_Targets(t *testing.T) {
	targets := DefaultRegistry.Targets()
	require.Contains(t, targets, "cursor")
	assert.True(t, strings.EqualFold(targets[0], targets[0]))
}
