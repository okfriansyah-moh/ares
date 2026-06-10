package compose

import (
	"testing"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeTargetsPreserveReadableInstructionBoundaries(t *testing.T) {
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Instructions: []arslib.Instruction{{
			ID:   "development-rules",
			Path: ".ai/instructions/development-rules.md",
			Content: "Design before code - brainstorming skill\n" +
				"Vertical slice per module - .github/skills/vertical-slice/SKILL.md\n" +
				"Commands and todo requirements: AGENTS.md Validation Requirement.\n" +
				"Use rtk for verbose command output.\n",
		}},
	}

	cases := map[string]string{
		"cursor":  ".cursor/rules/development-rules.mdc",
		"copilot": ".github/instructions/development-rules.instructions.md",
		"claude":  "CLAUDE.md",
		"codex":   "AGENTS.md",
	}

	for target, rel := range cases {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, Compose(root, target, repo))

			data, err := safepath.ReadFile(root, rel)
			require.NoError(t, err)
			body := string(data)

			assert.Contains(t, body, "skill\nVertical slice")
			assert.Contains(t, body, "Requirement.\nUse rtk")
			assert.NotContains(t, body, "skillVertical")
			assert.NotContains(t, body, "Requirement.Use")
		})
	}
}

func TestComposeTargetsSkipEmptyContentArtifacts(t *testing.T) {
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{Project: arslib.Project{Name: "demo"}},
		Instructions: []arslib.Instruction{
			{
				ID:      "empty-rules",
				Path:    ".ai/instructions/empty-rules.md",
				Content: " \n\t",
			},
			{
				ID:      "repo-rules",
				Path:    ".ai/instructions/repo-rules.md",
				Content: "Keep changes scoped.\n",
			},
		},
		Agents: []arslib.Agent{
			{
				ID:        "empty-agent",
				Path:      ".ai/agents/empty-agent/AGENT.md",
				Content:   "\n",
				SkillRefs: []string{"skills/empty-skill/SKILL.md"},
			},
		},
		Skills: []arslib.Skill{
			{
				ID:      "empty-skill",
				Path:    ".ai/skills/empty-skill/SKILL.md",
				Content: "",
			},
		},
		Prompts: []arslib.Prompt{
			{
				ID:      "empty-prompt",
				Path:    ".ai/prompts/empty-prompt.md",
				Content: " ",
			},
		},
	}

	for _, target := range []string{"cursor", "copilot", "claude", "codex"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			require.NoError(t, Compose(root, target, repo))

			switch target {
			case "cursor":
				exists, err := safepath.Exists(root, ".cursor/rules/empty-rules.mdc")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".cursor/rules/empty-agent.mdc")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".cursor/agents/empty-agent.md")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".cursor/skills/empty-skill/SKILL.md")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".cursor/prompts/empty-prompt.prompt")
				require.NoError(t, err)
				assert.False(t, exists)
			case "copilot":
				exists, err := safepath.Exists(root, ".github/instructions/empty-rules.instructions.md")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".github/agents/empty-agent.agent.md")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".github/skills/empty-skill/SKILL.md")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".github/prompts/empty-prompt.prompt.md")
				require.NoError(t, err)
				assert.False(t, exists)

				data, err := safepath.ReadFile(root, ".github/copilot-instructions.md")
				require.NoError(t, err)
				assert.NotContains(t, string(data), "empty-agent")
				assert.NotContains(t, string(data), "empty-skill")
			case "claude":
				data, err := safepath.ReadFile(root, "CLAUDE.md")
				require.NoError(t, err)
				assert.NotContains(t, string(data), "empty-rules")
				assert.NotContains(t, string(data), "empty-agent")
				assert.NotContains(t, string(data), "empty-skill")
				assert.Contains(t, string(data), "Keep changes scoped.")
				exists, err := safepath.Exists(root, ".claude/skills/empty-skill/SKILL.md")
				require.NoError(t, err)
				assert.False(t, exists)
			case "codex":
				data, err := safepath.ReadFile(root, "AGENTS.md")
				require.NoError(t, err)
				assert.NotContains(t, string(data), "empty-rules")
				assert.NotContains(t, string(data), "empty-agent")
				assert.NotContains(t, string(data), "empty-skill")
				assert.Contains(t, string(data), "Keep changes scoped.")
				exists, err := safepath.Exists(root, ".agents/skills/empty-skill/SKILL.md")
				require.NoError(t, err)
				assert.False(t, exists)
				exists, err = safepath.Exists(root, ".codex/agents/empty-agent.toml")
				require.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}
