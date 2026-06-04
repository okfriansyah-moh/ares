package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
)

const (
	cursorRulesTypeAlways         = "always"
	cursorRulesTypeAgentRequested = "agent-requested"
	arsSourceMarker               = "<!-- ars:source .ai/ -->\n"
)

// CursorComposer writes .cursor/ artifacts from a Repository.
type CursorComposer struct{}

func (c *CursorComposer) Target() string {
	return "cursor"
}

func (c *CursorComposer) Compose(root string, repo *arslib.Repository) error {
	if repo == nil {
		return fmt.Errorf("compose cursor: repository is nil")
	}

	cursorDir, err := safepath.Join(root, ".cursor")
	if err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	rulesDir, err := safepath.Join(root, ".cursor", "rules")
	if err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	promptsDir, err := safepath.Join(root, ".cursor", "prompts")
	if err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}

	if err := resetDir(cursorDir); err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}

	skillByID := indexSkills(repo.Skills)
	projectComment := fmt.Sprintf("<!-- project: %s -->\n", repo.Manifest.Project.Name)

	instructions := append([]arslib.Instruction(nil), repo.Instructions...)
	sort.Slice(instructions, func(i, j int) bool {
		return instructions[i].ID < instructions[j].ID
	})

	firstRule := true
	for _, inst := range instructions {
		name := sanitizeRuleName(inst.ID)
		body := arsSourceMarker + inst.Content
		if firstRule {
			body = projectComment + body
			firstRule = false
		}
		content := cursorRuleHeader(cursorRulesTypeAlways) + body
		rel := filepath.ToSlash(filepath.Join(".cursor", "rules", name+".mdc"))
		if err := safepath.WriteFile(root, rel, []byte(content), 0o644); err != nil {
			return fmt.Errorf("compose cursor: %w", err)
		}
	}

	agents := append([]arslib.Agent(nil), repo.Agents...)
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	for _, agent := range agents {
		name := sanitizeRuleName(agent.ID)
		body := arsSourceMarker + buildAgentRule(agent, skillByID)
		if firstRule {
			body = projectComment + body
			firstRule = false
		}
		content := cursorRuleHeader(cursorRulesTypeAgentRequested) + body
		rel := filepath.ToSlash(filepath.Join(".cursor", "rules", name+".mdc"))
		if err := safepath.WriteFile(root, rel, []byte(content), 0o644); err != nil {
			return fmt.Errorf("compose cursor: %w", err)
		}
	}

	prompts := append([]arslib.Prompt(nil), repo.Prompts...)
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].ID < prompts[j].ID
	})

	for _, prompt := range prompts {
		name := sanitizeRuleName(prompt.ID)
		rel := filepath.ToSlash(filepath.Join(".cursor", "prompts", name+".prompt"))
		content := arsSourceMarker + prompt.Content
		if err := safepath.WriteFile(root, rel, []byte(content), 0o644); err != nil {
			return fmt.Errorf("compose cursor: %w", err)
		}
	}

	return nil
}

func cursorRuleHeader(agentType string) string {
	return fmt.Sprintf("---\ntype: %s\n---\n", agentType)
}

func buildAgentRule(agent arslib.Agent, skills map[string]arslib.Skill) string {
	var b strings.Builder
	b.WriteString(agent.Content)
	for _, ref := range agent.SkillRefs {
		skill, ok := resolveSkill(ref, skills)
		if !ok {
			continue
		}
		b.WriteString("\n\n### Context: ")
		b.WriteString(skill.ID)
		b.WriteString("\n\n")
		b.WriteString(skill.Content)
	}
	return b.String()
}

func indexSkills(skills []arslib.Skill) map[string]arslib.Skill {
	out := make(map[string]arslib.Skill, len(skills))
	for _, skill := range skills {
		out[skill.ID] = skill
	}
	return out
}

func resolveSkill(ref string, skills map[string]arslib.Skill) (arslib.Skill, bool) {
	ref = strings.TrimSpace(ref)
	ref = strings.Trim(ref, `"'`)
	ref = strings.TrimPrefix(ref, "- ")
	if skill, ok := skills[ref]; ok {
		return skill, true
	}
	if skill, ok := skills[filepath.Base(filepath.Dir(ref))]; ok {
		return skill, true
	}
	for id, skill := range skills {
		if strings.Contains(ref, id) {
			return skill, true
		}
	}
	return arslib.Skill{}, false
}

func sanitizeRuleName(name string) string {
	clean := filepath.Clean(name)
	base := filepath.Base(clean)
	if base == "." || base == ".." || base == "" {
		return "invalid"
	}
	return base
}

func resetDir(path string) error {
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
