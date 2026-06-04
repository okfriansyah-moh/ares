package importer

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type classification int

const (
	classInstruction classification = iota
	classAgent
	classSkill
)

var (
	agentHeadingRe = regexp.MustCompile(`(?i)^agent[:\s]`)
	skillHeadingRe = regexp.MustCompile(`(?i)^skill[:\s]`)
)

// ClassifySection maps a heading to agent, skill, or instruction content.
func ClassifySection(heading string) classification {
	switch {
	case agentHeadingRe.MatchString(heading):
		return classAgent
	case skillHeadingRe.MatchString(heading):
		return classSkill
	default:
		return classInstruction
	}
}

// slugify converts a heading into a stable directory or file stem.
func slugify(heading string) string {
	h := strings.TrimSpace(heading)
	h = agentHeadingRe.ReplaceAllString(h, "")
	h = skillHeadingRe.ReplaceAllString(h, "")
	h = strings.TrimSpace(h)

	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(h) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if (r == ' ' || r == '-' || r == '_') && !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}

	s := strings.Trim(b.String(), "-")
	if len(s) > 50 {
		s = strings.TrimRight(s[:50], "-")
	}
	if s == "" {
		return "section"
	}
	return s
}

func uniqueSlug(heading string, used map[string]bool) string {
	base := slugify(heading)
	if !used[base] {
		used[base] = true
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}
