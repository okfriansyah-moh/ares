package compose

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/markdown"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

const (
	cursorARSTypeInstruction = "instruction"
	cursorARSTypeAgent       = "agent"
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

	if err := resetDir(root, ".cursor"); err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	if err := safepath.MkdirAll(root, ".cursor/rules", 0o755); err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	if err := safepath.MkdirAll(root, ".cursor/prompts", 0o755); err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	if err := safepath.MkdirAll(root, ".cursor/skills", 0o755); err != nil {
		return fmt.Errorf("compose cursor: %w", err)
	}
	if err := safepath.MkdirAll(root, ".cursor/agents", 0o755); err != nil {
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
		body := sourceMarker(inst.Path) + inst.Content
		if firstRule {
			body = projectComment + body
			firstRule = false
		}
		description := fmt.Sprintf("Repository rule: %s", name)
		content := cursorRuleHeader(true, description, cursorARSTypeInstruction) + body
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
		body := sourceMarker(agent.Path) + buildCursorAgentRule(agent, skillByID)
		if firstRule {
			body = projectComment + body
			firstRule = false
		}
		description := fmt.Sprintf("Use when %s", strings.ToLower(subagentDescription(agent.Content, name)))
		content := cursorRuleHeader(false, description, cursorARSTypeAgent) + body
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
		content := prompt.Content
		if err := safepath.WriteFile(root, rel, []byte(content), 0o644); err != nil {
			return fmt.Errorf("compose cursor: %w", err)
		}
	}

	skills := append([]arslib.Skill(nil), repo.Skills...)
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].ID < skills[j].ID
	})

	for _, skill := range skills {
		name := sanitizeRuleName(skill.ID)
		rel := filepath.ToSlash(filepath.Join(".cursor", "skills", name, "SKILL.md"))
		content := sourceMarker(skill.Path) + skill.Content
		if err := safepath.MkdirAll(root, filepath.ToSlash(filepath.Join(".cursor", "skills", name)), 0o755); err != nil {
			return fmt.Errorf("compose cursor: %w", err)
		}
		if err := safepath.WriteFile(root, rel, []byte(content), 0o644); err != nil {
			return fmt.Errorf("compose cursor: %w", err)
		}
	}

	for _, agent := range agents {
		name := sanitizeRuleName(agent.ID)
		rel := filepath.ToSlash(filepath.Join(".cursor", "agents", name+".md"))
		content := fmt.Sprintf("---\nname: %s\ndescription: %s\nmodel: inherit\n---\n\n%s",
			name,
			subagentDescription(agent.Content, name),
			buildCursorSubagent(agent, skillByID),
		)
		if err := safepath.WriteFile(root, rel, []byte(content), 0o644); err != nil {
			return fmt.Errorf("compose cursor: %w", err)
		}
	}

	return nil
}

func cursorRuleHeader(alwaysApply bool, description, arsType string) string {
	return fmt.Sprintf("---\ndescription: \"%s\"\nalwaysApply: %t\narsType: %s\n---\n", escapeYAMLDoubleQuoted(description), alwaysApply, arsType)
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

func buildCursorAgentRule(agent arslib.Agent, skills map[string]arslib.Skill) string {
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
		b.WriteString(sourceMarker(skill.Path))
		b.WriteString(skill.Content)
	}
	return b.String()
}

func buildCursorSubagent(agent arslib.Agent, skills map[string]arslib.Skill) string {
	policy := subagentPolicyFor(agent.ID)

	var b strings.Builder
	b.WriteString(sourceMarker(agent.Path))
	b.WriteString(agent.Content)

	b.WriteString("\n\n## Operating Policy\n")
	b.WriteString("- Focus on your role only and avoid taking work owned by other agents.\n")
	b.WriteString("- Execute the smallest useful change and keep recommendations concrete.\n")
	b.WriteString("- Before concluding, report assumptions and any unresolved risks.\n")

	b.WriteString("\n## Tooling Policy\n")
	for _, line := range policy.Tooling {
		b.WriteString("- ")
		b.WriteString(line)
		b.WriteByte('\n')
	}

	b.WriteString("\n## Output Contract\n")
	for _, line := range policy.Output {
		b.WriteString("- ")
		b.WriteString(line)
		b.WriteByte('\n')
	}

	for _, ref := range agent.SkillRefs {
		skill, ok := resolveSkill(ref, skills)
		if !ok {
			continue
		}
		b.WriteString("\n\n### Skill Context: ")
		b.WriteString(skill.ID)
		b.WriteString("\n\n")
		b.WriteString(sourceMarker(skill.Path))
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
	ref = strings.Trim(ref, "\"'`")
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

func resetDir(root, rel string) error {
	return safepath.RemoveAll(root, rel)
}

func sourceMarker(path string) string {
	if strings.TrimSpace(path) == "" {
		path = ".ai/"
	}
	return fmt.Sprintf("<!-- ars:source %s -->\n", filepath.ToSlash(path))
}

func subagentDescription(agentContent, fallback string) string {
	sections, err := markdown.ExtractSections([]byte(agentContent))
	if err == nil {
		if role, ok := markdown.FindSection(sections, "Role"); ok {
			line := strings.TrimSpace(strings.SplitN(role.Content, "\n", 2)[0])
			if line != "" {
				return line
			}
		}
	}
	return fmt.Sprintf("Subagent for %s tasks.", fallback)
}

type subagentPolicy struct {
	Tooling []string
	Output  []string
}

func subagentPolicyFor(agentID string) subagentPolicy {
	id := strings.ToLower(strings.TrimSpace(agentID))

	if strings.Contains(id, "architect") {
		return subagentPolicy{
			Tooling: []string{
				"Prefer read-first workflow: inspect architecture docs and ADRs before suggesting edits.",
				"Do not perform broad implementation edits unless explicitly requested.",
				"Use terminal commands only for validation and architecture evidence gathering.",
			},
			Output: []string{
				"Lead with architecture findings and tradeoffs.",
				"Provide a recommended option with rationale and risks.",
				"List concrete file-level actions when changes are required.",
			},
		}
	}

	if strings.Contains(id, "plan") {
		return subagentPolicy{
			Tooling: []string{
				"Prioritize reading source specs, existing plans, and architecture context.",
				"Avoid implementation edits; focus on sequencing and validation design.",
				"Use terminal commands only to verify references, targets, or required checks.",
			},
			Output: []string{
				"Return a sequenced, dependency-aware plan.",
				"Include explicit validation commands for each execution phase.",
				"Call out blockers, assumptions, and scope boundaries.",
			},
		}
	}

	if strings.Contains(id, "review") || strings.Contains(id, "audit") {
		return subagentPolicy{
			Tooling: []string{
				"Prefer read-only analysis and avoid mutating files unless asked to remediate findings.",
				"Use tests and static checks to confirm suspected issues.",
				"Collect precise evidence before reporting a finding.",
			},
			Output: []string{
				"Report findings first, ordered by severity.",
				"Include file references and concise reproduction evidence.",
				"If no findings, explicitly state no findings and mention residual risk.",
			},
		}
	}

	return subagentPolicy{
		Tooling: []string{
			"Read relevant files before making edits.",
			"Run focused validation for changed areas, then broader checks when needed.",
			"Avoid unrelated refactors and keep changes minimal.",
		},
		Output: []string{
			"Summarize what changed and why.",
			"Include validation results and any skipped checks.",
			"List blockers or follow-up actions when unresolved.",
		},
	}
}

func escapeYAMLDoubleQuoted(in string) string {
	out := strings.ReplaceAll(in, "\\", "\\\\")
	out = strings.ReplaceAll(out, "\"", "\\\"")
	return out
}
