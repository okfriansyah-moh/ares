package validator

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ars-standard/ars/internal/markdown"
	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
)

var requiredAgentSections = []string{
	"Role",
	"Responsibilities",
	"Uses",
	"Boundaries",
}

func validateAgents(root string) []arslib.Finding {
	agentsDir, err := safepath.Join(root, ".ai", "agents")
	if err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    ".ai/agents",
			Message: err.Error(),
		}}
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    ".ai/agents",
			Message: err.Error(),
		}}
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	var findings []arslib.Finding
	for _, name := range names {
		relPath := filepath.ToSlash(filepath.Join(".ai", "agents", name, "AGENT.md"))
		absPath, err := safepath.Join(root, ".ai", "agents", name, "AGENT.md")
		if err != nil {
			findings = append(findings, arslib.Finding{
				Level:   arslib.Error,
				Path:    relPath,
				Message: err.Error(),
			})
			continue
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			findings = append(findings, arslib.Finding{
				Level:   arslib.Error,
				Path:    relPath,
				Message: "AGENT.md does not exist",
			})
			continue
		}

		sections, err := markdown.ExtractSections(data)
		if err != nil {
			findings = append(findings, arslib.Finding{
				Level:   arslib.Error,
				Path:    relPath,
				Message: err.Error(),
			})
			continue
		}

		for _, heading := range requiredAgentSections {
			if !markdown.HasSection(sections, heading) {
				findings = append(findings, arslib.Finding{
					Level:   arslib.Error,
					Path:    relPath,
					Message: "missing required section ## " + heading,
				})
			}
		}

		usesSection, ok := markdown.FindSection(sections, "Uses")
		if !ok {
			continue
		}

		for _, ref := range extractSkillRefs(usesSection.Content) {
			findings = append(findings, validateSkillRef(root, relPath, ref)...)
		}
	}

	return findings
}

func extractSkillRefs(content string) []string {
	var refs []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimLeft(line, "-*")
		line = strings.TrimSpace(line)
		line = strings.Trim(line, `"'`)
		if line == "" {
			continue
		}
		if strings.Contains(line, "skills/") || strings.Contains(line, "SKILL.md") || strings.Contains(line, "..") {
			refs = append(refs, line)
		}
	}
	return refs
}

func validateSkillRef(root, agentPath, ref string) []arslib.Finding {
	rel, err := normaliseSkillRef(ref)
	if err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    agentPath,
			Message: err.Error(),
		}}
	}

	parts := strings.Split(filepath.ToSlash(rel), "/")
	absPath, err := safepath.Join(root, parts...)
	if err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    agentPath,
			Message: "skill reference escapes repository root: " + ref,
		}}
	}

	if _, err := os.Stat(absPath); err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    agentPath,
			Message: "skill reference not found: " + ref,
		}}
	}

	return nil
}

func normaliseSkillRef(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", nil
	}

	switch {
	case strings.HasPrefix(ref, ".ai/"):
		return ref, nil
	case strings.HasPrefix(ref, "skills/"):
		return filepath.ToSlash(filepath.Join(".ai", ref)), nil
	default:
		if strings.Contains(ref, "..") {
			return ref, nil
		}
		return "", nil
	}
}
