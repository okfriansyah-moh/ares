package validator

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

func validateSkills(root string) []arslib.Finding {
	entries, err := safepath.ReadDir(root, ".ai/skills")
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
		exists, err := safepath.Exists(root, relPath)
		if err != nil || !exists {
			msg := "SKILL.md does not exist"
			if err != nil {
				msg = err.Error()
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
