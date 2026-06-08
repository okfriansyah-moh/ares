package validator

import (
	"sort"

	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

// Run validates .ai/ under root and returns all findings sorted by path then level.
func Run(root string) ([]arslib.Finding, error) {
	findings := make([]arslib.Finding, 0)
	findings = append(findings, validateManifest(root)...)
	findings = append(findings, validateAgents(root)...)
	findings = append(findings, validateSkills(root)...)
	findings = append(findings, validatePrompts(root)...)

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Level < findings[j].Level
	})

	return findings, nil
}

func levelString(l arslib.FindingLevel) string {
	return l.String()
}
