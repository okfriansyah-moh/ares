package compose

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/content"
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

func filterContentfulSkills(skills []arslib.Skill) []arslib.Skill {
	out := make([]arslib.Skill, 0, len(skills))
	for _, skill := range skills {
		if content.HasBody(skill.Content) {
			out = append(out, skill)
		}
	}
	return out
}

func resolveSkill(ref string, skills map[string]arslib.Skill) (arslib.Skill, bool) {
	for _, candidate := range skillRefCandidates(ref) {
		if skill, ok := skills[candidate]; ok {
			return skill, true
		}
	}
	return arslib.Skill{}, false
}

func skillRefCandidates(ref string) []string {
	ref = strings.TrimSpace(ref)
	ref = strings.TrimLeft(ref, "-*")
	ref = strings.TrimSpace(strings.Trim(ref, "\"'`"))
	if ref == "" {
		return nil
	}

	candidates := []string{ref}
	slashed := filepath.ToSlash(ref)
	if slashed != ref {
		candidates = append(candidates, slashed)
	}
	if strings.Contains(slashed, "/") {
		dir := filepath.Base(filepath.Dir(slashed))
		if dir != "." && dir != "/" && dir != "" {
			candidates = append(candidates, dir)
		}
	}
	return candidates
}
