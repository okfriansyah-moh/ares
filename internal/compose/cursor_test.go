package compose

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ars-standard/ars/pkg/arslib"
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

	ruleData, err := os.ReadFile(rulePath)
	require.NoError(t, err)
	assert.Contains(t, string(ruleData), "type: always")
	assert.Contains(t, string(ruleData), "<!-- project: demo -->")
	assert.Contains(t, string(ruleData), "Repository rules.")

	agentData, err := os.ReadFile(agentRulePath)
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

	data, err := os.ReadFile(filepath.Join(root, ".cursor", "rules", "architect.mdc"))
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

	promptsDir := filepath.Join(root, ".cursor", "prompts")
	info, err := os.Stat(promptsDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())

	entries, err := os.ReadDir(promptsDir)
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
	first := treeChecksum(t, filepath.Join(root, ".cursor"))

	require.NoError(t, composer.Compose(root, repo))
	second := treeChecksum(t, filepath.Join(root, ".cursor"))

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

	rulesDir := filepath.Join(root, ".cursor", "rules")
	entries, err := os.ReadDir(rulesDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "evil.mdc", entries[0].Name())

	outside := filepath.Join(root, "evil.mdc")
	_, err = os.Stat(outside)
	assert.True(t, os.IsNotExist(err))
}

func TestCompose_UnknownTarget(t *testing.T) {
	err := Compose(t.TempDir(), "unknown", &arslib.Repository{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownTarget)
}

func TestCursorRuleHeader(t *testing.T) {
	assert.Equal(t, "---\ntype: always\n---\n", cursorRuleHeader("always"))
}

func treeChecksum(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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
		data, err := os.ReadFile(path)
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
