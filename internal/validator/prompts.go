package validator

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/markdown"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

func validatePrompts(root string) []arslib.Finding {
	entries, err := safepath.ReadDir(root, ".ai/prompts")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    ".ai/prompts",
			Message: err.Error(),
		}}
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".md") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	var findings []arslib.Finding
	for _, name := range names {
		relPath := filepath.ToSlash(filepath.Join(".ai", "prompts", name))
		data, err := safepath.ReadFile(root, relPath)
		if err != nil {
			msg := err.Error()
			if os.IsNotExist(err) {
				msg = "prompt file does not exist"
			}
			findings = append(findings, arslib.Finding{
				Level:   arslib.Error,
				Path:    relPath,
				Message: msg,
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

		if !markdown.HasSection(sections, "Use") {
			findings = append(findings, arslib.Finding{
				Level:   arslib.Warning,
				Path:    relPath,
				Message: "missing recommended section ## Use",
			})
		}
	}

	return findings
}
