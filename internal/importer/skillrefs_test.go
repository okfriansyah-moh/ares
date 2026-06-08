package importer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSkillRefs_FromPaths(t *testing.T) {
	content := `## Uses
- .ai/skills/plan-management/SKILL.md
- .github/skills/task-review/SKILL.md
`
	assert.Equal(t,
		[]string{"skills/plan-management/SKILL.md", "skills/task-review/SKILL.md"},
		extractSkillRefs(content),
	)
}

func TestExtractSkillRefs_FromContextAndBareIDs(t *testing.T) {
	content := `### Context: architecture-management

## Skills
- skills/task-implementation/SKILL.md
`
	assert.ElementsMatch(t,
		[]string{"skills/architecture-management/SKILL.md", "skills/task-implementation/SKILL.md"},
		extractSkillRefs(content),
	)
}
