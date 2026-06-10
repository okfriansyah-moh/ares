package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	contentutil "github.com/okfriansyah-moh/ares/internal/content"
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
	hasCopilotInstructions := err == nil
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("import github: %w", err)
	}

	sectionsRepo := emptyImportedRepository("imported-project")
	if hasCopilotInstructions {
		sections, err := markdown.ExtractSections(data)
		if err != nil {
			return nil, fmt.Errorf("import github: %w", err)
		}
		projectName := inferProjectName(data, sections)
		sectionsRepo = sectionsToRepository(sections, projectName)
	}

	repo := emptyImportedRepository(sectionsRepo.Manifest.Project.Name)
	mergeRepository(repo, sectionsRepo)

	if err := ingestGitHubInstructions(root, repo); err != nil {
		return nil, err
	}
	if err := ingestGitHubSkills(root, repo); err != nil {
		return nil, err
	}
	if err := ingestGitHubPrompts(root, repo); err != nil {
		return nil, err
	}
	if err := ingestGitHubAgents(root, repo); err != nil {
		return nil, err
	}

	if !hasCopilotInstructions && len(repo.Instructions) == 0 && len(repo.Agents) == 0 && len(repo.Skills) == 0 && len(repo.Prompts) == 0 {
		return nil, fmt.Errorf("import github: copilot artifacts not found at %s", path)
	}

	sort.Slice(repo.Instructions, func(i, j int) bool { return repo.Instructions[i].ID < repo.Instructions[j].ID })
	sort.Slice(repo.Agents, func(i, j int) bool { return repo.Agents[i].ID < repo.Agents[j].ID })
	sort.Slice(repo.Skills, func(i, j int) bool { return repo.Skills[i].ID < repo.Skills[j].ID })
	sort.Slice(repo.Prompts, func(i, j int) bool { return repo.Prompts[i].ID < repo.Prompts[j].ID })

	return repo, nil
}

func ingestGitHubInstructions(root string, repo *arslib.Repository) error {
	entries, err := safepath.ReadDir(root, ".github/instructions")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("import github: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".instructions.md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".instructions.md")
		rel := filepath.ToSlash(filepath.Join(".github", "instructions", entry.Name()))
		content, err := readGitHubMarkdownBody(root, rel)
		if err != nil {
			return fmt.Errorf("import github: %w", err)
		}
		if !contentutil.HasBody(content) {
			continue
		}
		upsertInstruction(repo, arslib.Instruction{
			ID:      id,
			Path:    filepath.ToSlash(filepath.Join(".ai", "instructions", id+".md")),
			Content: content,
		})
	}

	return nil
}

func ingestGitHubSkills(root string, repo *arslib.Repository) error {
	entries, err := safepath.ReadDir(root, ".github/skills")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("import github: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		rel := filepath.ToSlash(filepath.Join(".github", "skills", id, "SKILL.md"))
		exists, err := safepath.Exists(root, rel)
		if err != nil {
			return fmt.Errorf("import github: %w", err)
		}
		if !exists {
			continue
		}
		content, err := readGitHubMarkdownBody(root, rel)
		if err != nil {
			return fmt.Errorf("import github: %w", err)
		}
		if !contentutil.HasBody(content) {
			continue
		}
		extras, err := collectGitHubSkillExtraFiles(root, id)
		if err != nil {
			return fmt.Errorf("import github: %w", err)
		}
		upsertSkill(repo, arslib.Skill{
			ID:         id,
			Path:       filepath.ToSlash(filepath.Join(".ai", "skills", id, "SKILL.md")),
			Content:    content,
			ExtraFiles: extras,
		})
	}

	return nil
}

func collectGitHubSkillExtraFiles(root, skillID string) ([]arslib.ExtraFile, error) {
	baseRel := filepath.ToSlash(filepath.Join(".github", "skills", skillID))
	return walkGitHubExtraFiles(root, baseRel, "")
}

func walkGitHubExtraFiles(root, baseRel, subRel string) ([]arslib.ExtraFile, error) {
	dirRel := baseRel
	if subRel != "" {
		dirRel = filepath.ToSlash(filepath.Join(baseRel, subRel))
	}

	entries, err := safepath.ReadDir(root, dirRel)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var extras []arslib.ExtraFile
	for _, entry := range entries {
		entryRel := entry.Name()
		if subRel != "" {
			entryRel = filepath.ToSlash(filepath.Join(subRel, entry.Name()))
		}
		if entry.IsDir() {
			subs, err := walkGitHubExtraFiles(root, baseRel, entryRel)
			if err != nil {
				return nil, err
			}
			extras = append(extras, subs...)
		} else {
			if subRel == "" && entry.Name() == "SKILL.md" {
				continue
			}
			data, err := safepath.ReadFile(root, filepath.ToSlash(filepath.Join(baseRel, entryRel)))
			if err != nil {
				return nil, err
			}
			extras = append(extras, arslib.ExtraFile{Rel: entryRel, Content: data})
		}
	}
	sort.Slice(extras, func(i, j int) bool { return extras[i].Rel < extras[j].Rel })
	return extras, nil
}

func ingestGitHubPrompts(root string, repo *arslib.Repository) error {
	entries, err := safepath.ReadDir(root, ".github/prompts")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("import github: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".prompt.md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".prompt.md")
		rel := filepath.ToSlash(filepath.Join(".github", "prompts", entry.Name()))
		content, err := readGitHubMarkdownBody(root, rel)
		if err != nil {
			return fmt.Errorf("import github: %w", err)
		}
		if !contentutil.HasBody(content) {
			continue
		}
		upsertPrompt(repo, arslib.Prompt{
			ID:      id,
			Path:    filepath.ToSlash(filepath.Join(".ai", "prompts", id+".md")),
			Content: content,
		})
	}

	return nil
}

func ingestGitHubAgents(root string, repo *arslib.Repository) error {
	entries, err := safepath.ReadDir(root, ".github/agents")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("import github: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".agent.md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".agent.md")
		rel := filepath.ToSlash(filepath.Join(".github", "agents", entry.Name()))
		content, err := readGitHubMarkdownBody(root, rel)
		if err != nil {
			return fmt.Errorf("import github: %w", err)
		}
		if !contentutil.HasBody(content) {
			continue
		}
		upsertAgent(repo, arslib.Agent{
			ID:        id,
			Path:      filepath.ToSlash(filepath.Join(".ai", "agents", id, "AGENT.md")),
			Content:   content,
			SkillRefs: extractSkillRefs(content),
		})
	}

	return nil
}

func readGitHubMarkdownBody(root, rel string) (string, error) {
	data, err := safepath.ReadFile(root, rel)
	if err != nil {
		return "", err
	}
	body := stripMarkdownFrontmatter(string(data))
	body = cleanImportedMarkdownBody(body)
	if body == "" {
		return "", nil
	}
	return body, nil
}

func stripMarkdownFrontmatter(s string) string {
	t := strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
	if !strings.HasPrefix(t, "---\n") {
		return s
	}
	rest := strings.TrimPrefix(t, "---\n")
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return s
	}
	return strings.TrimPrefix(rest[idx:], "\n---\n")
}

func mergeRepository(dst, src *arslib.Repository) {
	dst.Manifest = src.Manifest
	for _, inst := range src.Instructions {
		upsertInstruction(dst, inst)
	}
	for _, skill := range src.Skills {
		upsertSkill(dst, skill)
	}
	for _, prompt := range src.Prompts {
		upsertPrompt(dst, prompt)
	}
	for _, agent := range src.Agents {
		upsertAgent(dst, agent)
	}
}

func upsertInstruction(repo *arslib.Repository, inst arslib.Instruction) {
	for i := range repo.Instructions {
		if repo.Instructions[i].ID == inst.ID {
			repo.Instructions[i] = inst
			return
		}
	}
	repo.Instructions = append(repo.Instructions, inst)
}

func upsertSkill(repo *arslib.Repository, skill arslib.Skill) {
	for i := range repo.Skills {
		if repo.Skills[i].ID == skill.ID {
			repo.Skills[i] = skill
			return
		}
	}
	repo.Skills = append(repo.Skills, skill)
}

func upsertPrompt(repo *arslib.Repository, prompt arslib.Prompt) {
	for i := range repo.Prompts {
		if repo.Prompts[i].ID == prompt.ID {
			repo.Prompts[i] = prompt
			return
		}
	}
	repo.Prompts = append(repo.Prompts, prompt)
}

func upsertAgent(repo *arslib.Repository, agent arslib.Agent) {
	for i := range repo.Agents {
		if repo.Agents[i].ID == agent.ID {
			repo.Agents[i] = agent
			return
		}
	}
	repo.Agents = append(repo.Agents, agent)
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

		content := cleanImportedMarkdownBody(sec.Content)
		relBase := filepath.Join(".ai")
		kind := ClassifySection(heading)

		switch kind {
		case classAgent:
			content = mergeAgentSections(sections, i, content)
			i = skipMergedSections(sections, i)
			if !contentutil.HasBody(content) {
				continue
			}
			id := uniqueSlug(heading, used)
			repo.Agents = append(repo.Agents, arslib.Agent{
				ID:        id,
				Path:      filepath.ToSlash(filepath.Join(relBase, "agents", id, "AGENT.md")),
				Content:   content,
				SkillRefs: extractSkillRefs(content),
			})
		case classSkill:
			content = mergeSkillSections(sections, i, content)
			i = skipMergedSections(sections, i)
			if !contentutil.HasBody(content) {
				continue
			}
			id := uniqueSlug(heading, used)
			repo.Skills = append(repo.Skills, arslib.Skill{
				ID:      id,
				Path:    filepath.ToSlash(filepath.Join(relBase, "skills", id, "SKILL.md")),
				Content: content,
			})
		default:
			if !contentutil.HasBody(content) {
				continue
			}
			id := uniqueSlug(heading, used)
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
