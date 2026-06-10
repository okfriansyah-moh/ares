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

// ClaudeComposer writes CLAUDE.md at the repository root.
type ClaudeComposer struct{}

func (c *ClaudeComposer) Target() string {
	return "claude"
}

func (c *ClaudeComposer) Compose(root string, repo *arslib.Repository) error {
	if repo == nil {
		return fmt.Errorf("compose claude: repository is nil")
	}

	if err := validateAgentIDs(root, repo.Agents); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}

	if _, err := safepath.Join(root, "CLAUDE.md"); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}
	if err := resetDir(root, ".claude"); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}
	if err := safepath.MkdirAll(root, ".claude/skills", 0o755); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}

	skills := append([]arslib.Skill(nil), repo.Skills...)
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].ID < skills[j].ID
	})

	skillNameByID := map[string]string{}
	seenSkillDirs := map[string]string{}
	for _, skill := range skills {
		skillDir := normalizeClaudeSkillName(skill.ID)
		if err := detectNormalizedCollision(seenSkillDirs, skillDir, skill.ID, "claude", "skill"); err != nil {
			return err
		}
		skillNameByID[skill.ID] = skillDir

		skillPath := filepath.ToSlash(filepath.Join(".claude", "skills", skillDir, "SKILL.md"))
		if err := safepath.MkdirAll(root, filepath.ToSlash(filepath.Join(".claude", "skills", skillDir)), 0o755); err != nil {
			return fmt.Errorf("compose claude: %w", err)
		}

		content := fmt.Sprintf("---\nname: %s\ndescription: \"%s\"\n---\n\n<!-- ars:source %s -->\n%s",
			skillDir,
			escapeYAMLDoubleQuoted(claudeSkillDescription(skill)),
			pathOrDefault(skill.Path, filepath.ToSlash(filepath.Join(".ai", "skills", skill.ID, "SKILL.md"))),
			ensureTrailingNewline(skill.Content),
		)
		if err := safepath.WriteFile(root, skillPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("compose claude: %w", err)
		}

		for _, ef := range skill.ExtraFiles {
			if err := safepath.ValidateExtraFileRel(ef.Rel); err != nil {
				return fmt.Errorf("compose claude: skill %q extra file: %w", skill.ID, err)
			}
			efRel := filepath.ToSlash(filepath.Join(".claude", "skills", skillDir, ef.Rel))
			if err := safepath.MkdirAll(root, filepath.ToSlash(filepath.Dir(efRel)), 0o755); err != nil {
				return fmt.Errorf("compose claude: %w", err)
			}
			if err := safepath.WriteFile(root, efRel, ef.Content, 0o644); err != nil {
				return fmt.Errorf("compose claude: %w", err)
			}
		}
	}

	content := buildClaudeRootOutput(repo, skillNameByID)

	if err := safepath.WriteFile(root, "CLAUDE.md", []byte(content), 0o644); err != nil {
		return fmt.Errorf("compose claude: %w", err)
	}

	return nil
}

func buildClaudeRootOutput(repo *arslib.Repository, skillNameByID map[string]string) string {
	var b strings.Builder
	b.WriteString(arsSourceMarker)
	b.WriteString("# ")
	b.WriteString(repo.Manifest.Project.Name)
	b.WriteString("\n\n")

	instructions := append([]arslib.Instruction(nil), repo.Instructions...)
	sort.Slice(instructions, func(i, j int) bool {
		return instructions[i].ID < instructions[j].ID
	})

	nonEmptyInstructions := make([]arslib.Instruction, 0, len(instructions))
	for _, inst := range instructions {
		if hasContentBody(inst.Content) {
			nonEmptyInstructions = append(nonEmptyInstructions, inst)
		}
	}

	if len(nonEmptyInstructions) > 0 {
		b.WriteString("## Repository Instructions\n\n")
		for _, inst := range nonEmptyInstructions {
			b.WriteString("<!-- ars:source ")
			b.WriteString(pathOrDefault(inst.Path, filepath.ToSlash(filepath.Join(".ai", "instructions", inst.ID+".md"))))
			b.WriteString(" -->\n")
			b.WriteString(ensureTrailingNewline(inst.Content))
			b.WriteByte('\n')
		}
	}

	if len(skillNameByID) > 0 {
		b.WriteString("## Claude Skills\n\n")
		orderedSkills := make([]string, 0, len(skillNameByID))
		for _, dir := range skillNameByID {
			orderedSkills = append(orderedSkills, dir)
		}
		sort.Strings(orderedSkills)
		for _, dir := range orderedSkills {
			b.WriteString("- .claude/skills/")
			b.WriteString(dir)
			b.WriteString("/SKILL.md\n")
		}
		b.WriteByte('\n')
	}

	agents := append([]arslib.Agent(nil), repo.Agents...)
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	for _, agent := range agents {
		b.WriteString("## ")
		b.WriteString(strings.ToLower(agent.ID))
		b.WriteString("\n\n")
		b.WriteString("<!-- ars:source ")
		b.WriteString(pathOrDefault(agent.Path, filepath.ToSlash(filepath.Join(".ai", "agents", agent.ID, "AGENT.md"))))
		b.WriteString(" -->\n")
		b.WriteString(ensureTrailingNewline(agent.Content))

		resolved := extractReferencedSkillDirs(agent.SkillRefs, skillNameByID)
		if len(resolved) > 0 {
			b.WriteString("\n### Skills\n")
			for _, dir := range resolved {
				b.WriteString("- .claude/skills/")
				b.WriteString(dir)
				b.WriteString("/SKILL.md\n")
			}
		}
		b.WriteByte('\n')
	}

	return b.String()
}

func extractReferencedSkillDirs(refs []string, skillNameByID map[string]string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		skill, ok := resolveSkill(ref, mapSkillsByDir(skillNameByID))
		if !ok {
			continue
		}
		dir := skillNameByID[skill.ID]
		if dir == "" {
			continue
		}
		if _, exists := seen[dir]; exists {
			continue
		}
		seen[dir] = struct{}{}
		out = append(out, dir)
	}
	sort.Strings(out)
	return out
}

func mapSkillsByDir(skillNameByID map[string]string) map[string]arslib.Skill {
	out := map[string]arslib.Skill{}
	for id, dir := range skillNameByID {
		out[id] = arslib.Skill{ID: id}
		out[dir] = arslib.Skill{ID: id}
	}
	return out
}

func normalizeClaudeSkillName(id string) string {
	raw := strings.ToLower(strings.TrimSpace(id))
	if raw == "" {
		return "skill"
	}

	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		isAlpha := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if isAlpha || isDigit {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	name := strings.Trim(b.String(), "-")
	if name == "" {
		name = "skill"
	}
	name = strings.ReplaceAll(name, "anthropic", "ars")
	name = strings.ReplaceAll(name, "claude", "ars")
	name = strings.Trim(name, "-")
	if name == "" {
		name = "skill"
	}
	if len(name) > 64 {
		name = strings.Trim(name[:64], "-")
		if name == "" {
			name = "skill"
		}
	}
	return name
}

func claudeSkillDescription(skill arslib.Skill) string {
	sections, err := markdown.ExtractSections([]byte(skill.Content))
	if err == nil {
		if purpose, ok := markdown.FindSection(sections, "Purpose"); ok {
			line := strings.TrimSpace(strings.SplitN(purpose.Content, "\n", 2)[0])
			if line != "" {
				return trimToMax(line, 1024)
			}
		}
	}
	return trimToMax(fmt.Sprintf("Reusable workflow and best practices for %s tasks. Use when relevant.", skill.ID), 1024)
}

func trimToMax(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max])
}

func pathOrDefault(path, fallback string) string {
	if strings.TrimSpace(path) == "" {
		return fallback
	}
	return filepath.ToSlash(path)
}

func ensureTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}
