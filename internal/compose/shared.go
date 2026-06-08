package compose

import (
	"fmt"
	"sort"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

// composerFormat configures shared single-file markdown compose output.
type composerFormat struct {
	agentSection func(agent arslib.Agent, skills map[string]arslib.Skill) string
}

func buildMarkdownOutput(format composerFormat, repo *arslib.Repository) string {
	var b strings.Builder
	b.WriteString(arsSourceMarker)
	b.WriteString("# ")
	b.WriteString(repo.Manifest.Project.Name)
	b.WriteString("\n\n")

	instructions := append([]arslib.Instruction(nil), repo.Instructions...)
	sort.Slice(instructions, func(i, j int) bool {
		return instructions[i].ID < instructions[j].ID
	})

	if len(instructions) > 0 {
		b.WriteString("## Repository Instructions\n\n")
		for _, inst := range instructions {
			b.WriteString(inst.Content)
			if !strings.HasSuffix(inst.Content, "\n") {
				b.WriteByte('\n')
			}
			b.WriteByte('\n')
		}
	}

	skillByID := indexSkills(repo.Skills)
	agents := append([]arslib.Agent(nil), repo.Agents...)
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	for _, agent := range agents {
		b.WriteString(format.agentSection(agent, skillByID))
		b.WriteByte('\n')
	}

	return b.String()
}

func claudeAgentSection(agent arslib.Agent, skills map[string]arslib.Skill) string {
	var b strings.Builder
	b.WriteString("## ")
	b.WriteString(strings.ToLower(agent.ID))
	b.WriteString("\n\n")
	b.WriteString(buildAgentRule(agent, skills))
	return b.String()
}

func codexAgentSection(agent arslib.Agent, skills map[string]arslib.Skill) string {
	var b strings.Builder
	b.WriteString("---\nagent: ")
	b.WriteString(agent.ID)
	b.WriteString("\n---\n\n")
	b.WriteString(buildAgentRule(agent, skills))
	return b.String()
}

func validateAgentIDs(root string, agents []arslib.Agent) error {
	for _, agent := range agents {
		if agent.ID != sanitizeRuleName(agent.ID) {
			return fmt.Errorf("compose: invalid agent id %q", agent.ID)
		}
		if _, err := safepath.Join(root, ".ai", "agents", agent.ID); err != nil {
			return err
		}
	}
	return nil
}
