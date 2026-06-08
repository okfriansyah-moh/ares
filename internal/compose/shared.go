package compose

import (
	"fmt"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

const arsSourceMarker = "<!-- ars:source .ai/ -->\n"

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

func detectNormalizedCollision(seen map[string]string, normalized, original, target, kind string) error {
	if prev, ok := seen[normalized]; ok && prev != original {
		return fmt.Errorf("compose %s: %s id %q normalizes to %q which collides with %q", target, kind, original, normalized, prev)
	}
	seen[normalized] = original
	return nil
}
