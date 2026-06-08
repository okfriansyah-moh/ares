package importer

import (
	"regexp"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/markdown"
)

var (
	skillPathRefRe   = regexp.MustCompile(`(?i)(?:^|[\s\("'` + "`" + `])(?:\.ai/|\.github/|\.claude/|\.agents/|\.cursor/)?skills/([a-z0-9._-]+)/SKILL\.md`)
	contextHeadingRe = regexp.MustCompile(`(?i)^context:\s*([a-z0-9._-]+)\s*$`)
)

func extractSkillRefs(content string) []string {
	refs := make([]string, 0)
	seen := map[string]struct{}{}

	add := func(id string) {
		id = strings.ToLower(strings.TrimSpace(id))
		if id == "" {
			return
		}
		ref := "skills/" + id + "/SKILL.md"
		if _, ok := seen[ref]; ok {
			return
		}
		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}

	for _, m := range skillPathRefRe.FindAllStringSubmatch(content, -1) {
		if len(m) > 1 {
			add(m[1])
		}
	}

	sections, err := markdown.ExtractSections([]byte(content))
	if err == nil {
		for _, sec := range sections {
			h := strings.TrimSpace(sec.Heading)
			if h == "" {
				continue
			}
			if m := contextHeadingRe.FindStringSubmatch(strings.ToLower(h)); len(m) > 1 {
				add(m[1])
			}

			head := strings.ToLower(h)
			if head != "uses" && head != "skills" {
				continue
			}
			for _, line := range strings.Split(sec.Content, "\n") {
				line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
				line = strings.Trim(line, "`\"'")
				if line == "" {
					continue
				}
				for _, m := range skillPathRefRe.FindAllStringSubmatch(line, -1) {
					if len(m) > 1 {
						add(m[1])
					}
				}
			}
		}
	}

	return refs
}
