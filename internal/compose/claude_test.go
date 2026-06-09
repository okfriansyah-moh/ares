package compose

import (
	"path/filepath"
	"strings"
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
		Instructions: []arslib.Instruction{{
			ID:      "architecture",
			Path:    ".ai/instructions/architecture.md",
			Content: "# Architecture\n",
		}},
		Skills: []arslib.Skill{{
			ID:      "plan-management",
			Path:    ".ai/skills/plan-management/SKILL.md",
			Content: "# Plan Management\n\n## Purpose\nCreate implementation plans.\n",
		}},
	}

	require.NoError(t, (&ClaudeComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, "CLAUDE.md")
	require.NoError(t, err)
	assert.Contains(t, string(data), "<!-- ars:source .ai/ -->")
	assert.Contains(t, string(data), "## Claude Skills")
	assert.Contains(t, string(data), ".claude/skills/plan-management/SKILL.md")

	skillData, err := safepath.ReadFile(root, ".claude/skills/plan-management/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(skillData), "name: plan-management")
	assert.Contains(t, string(skillData), "description: \"Create implementation plans.\"")
	assert.Contains(t, string(skillData), "<!-- ars:source .ai/skills/plan-management/SKILL.md -->")
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
		Skills: []arslib.Skill{{
			ID:      "task-review",
			Path:    ".ai/skills/task-review/SKILL.md",
			Content: "# Task Review\n\n## Purpose\nReview tasks.\n",
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

func TestNormalizeClaudeSkillName_Compliance(t *testing.T) {
	assert.Equal(t, "plan-management", normalizeClaudeSkillName("Plan_Management"))
	assert.Equal(t, "ars-helper", normalizeClaudeSkillName("claude-helper"))
	assert.Equal(t, "skill", normalizeClaudeSkillName("***"))
	assert.LessOrEqual(t, len(normalizeClaudeSkillName(strings.Repeat("a", 90))), 64)
}

func TestClaudeAgentSection_LowercaseHeading(t *testing.T) {
	section := claudeAgentSection(arslib.Agent{
		ID:      "Planner",
		Content: "Plans.",
	}, nil)
	assert.Contains(t, section, "## planner\n")
}

func TestClaudeComposer_SkillExtraFiles(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Skills: []arslib.Skill{{
			ID:      "plan-management",
			Path:    ".ai/skills/plan-management/SKILL.md",
			Content: "# Plan Management\n\n## Purpose\nManage plans.\n",
			ExtraFiles: []arslib.ExtraFile{
				{Rel: "reference/reference.md", Content: []byte("# Reference\nDetailed content.\n")},
			},
		}},
	}

	require.NoError(t, (&ClaudeComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, ".claude/skills/plan-management/reference/reference.md")
	require.NoError(t, err)
	assert.Equal(t, "# Reference\nDetailed content.\n", string(data))
}

func TestClaudeComposer_FailsOnSkillNameCollision(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Skills: []arslib.Skill{
			{ID: "claude-helper", Content: "first"},
			{ID: "anthropic_helper", Content: "second"},
		},
	}

	err := (&ClaudeComposer{}).Compose(root, repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compose claude: skill id")
	assert.Contains(t, err.Error(), "normalizes to \"ars-helper\"")
}
