package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/markdown"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

var h1LineRe = regexp.MustCompile(`(?m)^#\s+(.+)\s*$`)

// GitHubImporter imports from .github/copilot-instructions.md.
type GitHubImporter struct{}

func (g *GitHubImporter) Source() string {
	return "github"
}

func (g *GitHubImporter) Import(root string) (*arslib.Repository, error) {
	path, err := safepath.Join(root, ".github", "copilot-instructions.md")
	if err != nil {
		return nil, fmt.Errorf("import github: %w", err)
	}

	data, err := safepath.ReadFile(root, ".github/copilot-instructions.md")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("import github: copilot-instructions.md not found at %s", path)
		}
		return nil, fmt.Errorf("import github: %w", err)
	}

	sections, err := markdown.ExtractSections(data)
	if err != nil {
		return nil, fmt.Errorf("import github: %w", err)
	}

	projectName := inferProjectName(data, sections)
	return sectionsToRepository(sections, projectName), nil
}

func inferProjectName(raw []byte, sections []markdown.Section) string {
	for _, sec := range sections {
		if sec.Level == 1 && strings.TrimSpace(sec.Heading) != "" {
			return strings.TrimSpace(sec.Heading)
		}
	}
	if match := h1LineRe.FindSubmatch(raw); len(match) > 1 {
		return strings.TrimSpace(string(match[1]))
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "<!--") {
			continue
		}
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimPrefix(line, "#"))
		}
		return line
	}
	return "imported-project"
}

func sectionsToRepository(sections []markdown.Section, projectName string) *arslib.Repository {
	repo := &arslib.Repository{
		Manifest: arslib.Manifest{
			Version: "2.0",
			Project: arslib.Project{Name: projectName},
		},
	}

	used := make(map[string]bool)
	for i := 0; i < len(sections); i++ {
		sec := sections[i]
		heading := strings.TrimSpace(sec.Heading)
		if heading == "" {
			continue
		}
		if sec.Level == 1 && strings.EqualFold(heading, projectName) {
			continue
		}

		id := uniqueSlug(heading, used)
		content := strings.TrimSpace(sec.Content)
		relBase := filepath.Join(".ai")

		switch ClassifySection(heading) {
		case classAgent:
			content = mergeAgentSections(sections, i, content)
			i = skipMergedSections(sections, i)
			repo.Agents = append(repo.Agents, arslib.Agent{
				ID:      id,
				Path:    filepath.ToSlash(filepath.Join(relBase, "agents", id, "AGENT.md")),
				Content: content,
			})
		case classSkill:
			content = mergeSkillSections(sections, i, content)
			i = skipMergedSections(sections, i)
			repo.Skills = append(repo.Skills, arslib.Skill{
				ID:      id,
				Path:    filepath.ToSlash(filepath.Join(relBase, "skills", id, "SKILL.md")),
				Content: content,
			})
		default:
			repo.Instructions = append(repo.Instructions, arslib.Instruction{
				ID:      id,
				Path:    filepath.ToSlash(filepath.Join(relBase, "instructions", id+".md")),
				Content: content,
			})
		}
	}

	return repo
}

func mergeAgentSections(sections []markdown.Section, start int, base string) string {
	return mergeChildSections(sections, start, base, classAgent)
}

func mergeSkillSections(sections []markdown.Section, start int, base string) string {
	return mergeChildSections(sections, start, base, classSkill)
}

func mergeChildSections(sections []markdown.Section, start int, base string, kind classification) string {
	var b strings.Builder
	b.WriteString(base)
	for j := start + 1; j < len(sections); j++ {
		next := sections[j]
		heading := strings.TrimSpace(next.Heading)
		if heading == "" {
			continue
		}
		switch ClassifySection(heading) {
		case classAgent, classSkill:
			return strings.TrimSpace(b.String())
		}
		if isTopLevelInstruction(heading) {
			return strings.TrimSpace(b.String())
		}
		b.WriteString("\n\n## ")
		b.WriteString(heading)
		b.WriteString("\n\n")
		b.WriteString(strings.TrimSpace(next.Content))
	}
	return strings.TrimSpace(b.String())
}

func skipMergedSections(sections []markdown.Section, start int) int {
	for j := start + 1; j < len(sections); j++ {
		heading := strings.TrimSpace(sections[j].Heading)
		if heading == "" {
			continue
		}
		switch ClassifySection(heading) {
		case classAgent, classSkill:
			return j - 1
		}
		if isTopLevelInstruction(heading) {
			return j - 1
		}
	}
	return len(sections) - 1
}

func isTopLevelInstruction(heading string) bool {
	return strings.EqualFold(strings.TrimSpace(heading), "Repository Instructions")
}
