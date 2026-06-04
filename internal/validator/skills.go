package validator

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
)

func validateSkills(root string) []arslib.Finding {
	skillsDir, err := safepath.Join(root, ".ai", "skills")
	if err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    ".ai/skills",
			Message: err.Error(),
		}}
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    ".ai/skills",
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
		relPath := filepath.ToSlash(filepath.Join(".ai", "skills", name, "SKILL.md"))
		absPath, err := safepath.Join(root, ".ai", "skills", name, "SKILL.md")
		if err != nil {
			findings = append(findings, arslib.Finding{
				Level:   arslib.Error,
				Path:    relPath,
				Message: err.Error(),
			})
			continue
		}

		if _, err := os.Stat(absPath); err != nil {
			msg := err.Error()
			if os.IsNotExist(err) {
				msg = "SKILL.md does not exist"
			}
			findings = append(findings, arslib.Finding{
				Level:   arslib.Error,
				Path:    relPath,
				Message: msg,
			})
		}
	}

	return findings
}
