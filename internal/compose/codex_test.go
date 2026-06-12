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
		Instructions: []arslib.Instruction{{
			ID:      "repo",
			Path:    ".ai/instructions/repo.md",
			Content: "Repository rules.\n",
		}},
		Skills: []arslib.Skill{{
			ID:      "plan-management",
			Path:    ".ai/skills/plan-management/SKILL.md",
			Content: "# Plan Management\n\n## Purpose\nCreate reliable implementation plans.\n",
		}},
		Agents: []arslib.Agent{{
			ID:      "planner",
			Path:    ".ai/agents/planner/AGENT.md",
			Content: "## Role\nPlans work.\n",
			SkillRefs: []string{
				"skills/plan-management/SKILL.md",
			},
		}},
	}

	require.NoError(t, (&CodexComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, "AGENTS.md")
	require.NoError(t, err)
	assert.Contains(t, string(data), "<!-- ars:source .ai/ -->")
	assert.Contains(t, string(data), "## Codex Skills")
	assert.Contains(t, string(data), ".agents/skills/plan-management/SKILL.md")

	skillData, err := safepath.ReadFile(root, ".agents/skills/plan-management/SKILL.md")
	require.NoError(t, err)
	assert.Contains(t, string(skillData), "name: plan-management")
	assert.Contains(t, string(skillData), "description:")

	metaData, err := safepath.ReadFile(root, ".agents/skills/plan-management/agents/openai.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(metaData), "allow_implicit_invocation: true")

	subagentData, err := safepath.ReadFile(root, ".codex/agents/planner.toml")
	require.NoError(t, err)
	assert.Contains(t, string(subagentData), "name = \"planner\"")
	assert.Contains(t, string(subagentData), "description =")
	assert.Contains(t, string(subagentData), "developer_instructions = \"\"\"")
	assert.Contains(t, string(subagentData), "sandbox_mode = \"read-only\"")

	configData, err := safepath.ReadFile(root, ".codex/config.toml")
	require.NoError(t, err)
	assert.Contains(t, string(configData), "[agents]")
	assert.Contains(t, string(configData), "max_threads = 6")

	rulesData, err := safepath.ReadFile(root, ".codex/rules/ares.rules")
	require.NoError(t, err)
	assert.Contains(t, string(rulesData), "prefix_rule(")
	assert.Contains(t, string(rulesData), "decision = \"forbidden\"")
}

func TestCodexComposer_SkillExtraFiles(t *testing.T) {
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

	require.NoError(t, (&CodexComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, ".agents/skills/plan-management/reference/reference.md")
	require.NoError(t, err)
	assert.Equal(t, "# Reference\nDetailed content.\n", string(data))
}

func TestCodexComposer_PreservesExistingAGENTSMD(t *testing.T) {
	root := t.TempDir()

	existing := "# Custom AGENTS.md\nDo not overwrite this.\n"
	require.NoError(t, safepath.WriteFile(root, "AGENTS.md", []byte(existing), 0o644))

	repo := &arslib.Repository{
		Manifest:     arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Instructions: []arslib.Instruction{{ID: "rules", Content: "Some rules."}},
	}

	require.NoError(t, (&CodexComposer{}).Compose(root, repo))

	data, err := safepath.ReadFile(root, "AGENTS.md")
	require.NoError(t, err)
	assert.Equal(t, existing, string(data))

	configData, err := safepath.ReadFile(root, ".codex/config.toml")
	require.NoError(t, err)
	assert.Contains(t, string(configData), "[agents]")
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
			Path:    ".ai/agents/planner/AGENT.md",
			Content: "Plans work.",
		}},
		Skills: []arslib.Skill{{
			ID:      "task-review",
			Path:    ".ai/skills/task-review/SKILL.md",
			Content: "# Task Review\n\n## Purpose\nReview task output quality.\n",
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

func TestNormalizeCodexSkillName(t *testing.T) {
	assert.Equal(t, "plan-management", normalizeCodexSkillName("Plan_Management"))
	assert.Equal(t, "skill", normalizeCodexSkillName("***"))
}

func TestCodexComposer_FailsOnSkillNameCollision(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Skills: []arslib.Skill{
			{ID: "Plan_Management", Content: "first"},
			{ID: "plan-management", Content: "second"},
		},
	}

	err := (&CodexComposer{}).Compose(root, repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compose codex: skill id")
	assert.Contains(t, err.Error(), "normalizes to \"plan-management\"")
}

func TestCodexComposer_FailsOnAgentNameCollision(t *testing.T) {
	root := t.TempDir()
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Agents: []arslib.Agent{
			{ID: "Planner_Team", Content: "first"},
			{ID: "planner-team", Content: "second"},
		},
	}

	err := (&CodexComposer{}).Compose(root, repo)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compose codex: agent id")
	assert.Contains(t, err.Error(), "normalizes to \"planner-team\"")
}
