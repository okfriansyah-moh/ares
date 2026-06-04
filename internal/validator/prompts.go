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

func validatePrompts(root string) []arslib.Finding {
	promptsDir, err := safepath.Join(root, ".ai", "prompts")
	if err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    ".ai/prompts",
			Message: err.Error(),
		}}
	}

	entries, err := os.ReadDir(promptsDir)
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
		absPath, err := safepath.Join(root, ".ai", "prompts", name)
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
				Message: "prompt file does not exist",
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
