package markdown

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractSections_Basic(t *testing.T) {
	src := []byte(`## Role
Owns planning.

## Responsibilities
- Plan work
- Track tasks

## Uses
- skills/plan-management/SKILL.md
`)

	sections, err := ExtractSections(src)
	require.NoError(t, err)
	require.Len(t, sections, 3)

	assert.Equal(t, "Role", sections[0].Heading)
	assert.Equal(t, 2, sections[0].Level)
	assert.Contains(t, sections[0].Content, "Owns planning.")

	assert.Equal(t, "Responsibilities", sections[1].Heading)
	assert.Equal(t, 2, sections[1].Level)
	assert.Contains(t, sections[1].Content, "Plan work")

	assert.Equal(t, "Uses", sections[2].Heading)
	assert.Equal(t, 2, sections[2].Level)
	assert.Contains(t, sections[2].Content, "plan-management")
}

func TestExtractSections_Nested(t *testing.T) {
	src := []byte(`## Parent
parent body

### Child
child body
`)

	sections, err := ExtractSections(src)
	require.NoError(t, err)
	require.Len(t, sections, 2)

	assert.Equal(t, "Parent", sections[0].Heading)
	assert.Equal(t, 2, sections[0].Level)
	assert.Contains(t, sections[0].Content, "parent body")

	assert.Equal(t, "Child", sections[1].Heading)
	assert.Equal(t, 3, sections[1].Level)
	assert.Contains(t, sections[1].Content, "child body")
}

func TestExtractSections_PreservesReadableBoundaries(t *testing.T) {
	src := []byte(`## Repository Instructions
Design before code - brainstorming skill
Vertical slice per module - .github/skills/vertical-slice/SKILL.md

- Modules export service interfaces only - no cross-module internal imports
- No map[string]any across module boundaries - typed structs
- backend/internal/platform/ = shared infra

Commands and todo requirements: AGENTS.md Validation Requirement.
Use rtk for verbose command output.
`)

	sections, err := ExtractSections(src)
	require.NoError(t, err)
	require.Len(t, sections, 1)

	content := sections[0].Content
	assert.Contains(t, content, "skill\nVertical slice")
	assert.Contains(t, content, "imports\n- No map")
	assert.Contains(t, content, "structs\n- backend/internal/platform")
	assert.Contains(t, content, "Requirement.\nUse rtk")
	assert.NotContains(t, content, "skillVertical")
	assert.NotContains(t, content, "imports- No")
	assert.NotContains(t, content, "Requirement.Use")
}

func TestExtractSections_Empty(t *testing.T) {
	sections, err := ExtractSections(nil)
	require.NoError(t, err)
	assert.Empty(t, sections)

	sections, err = ExtractSections([]byte{})
	require.NoError(t, err)
	assert.Empty(t, sections)
}

func TestExtractSections_NoHeadings(t *testing.T) {
	src := []byte("Plain repository guidance without headings.\n")

	sections, err := ExtractSections(src)
	require.NoError(t, err)
	require.Len(t, sections, 1)
	assert.Empty(t, sections[0].Heading)
	assert.Contains(t, sections[0].Content, "Plain repository guidance")
}

func TestFindSection_CaseInsensitive(t *testing.T) {
	src := []byte(`## role
Role content
`)

	sections, err := ExtractSections(src)
	require.NoError(t, err)
	section, ok := FindSection(sections, "Role")
	require.True(t, ok)
	assert.Equal(t, "role", section.Heading)
	assert.Contains(t, section.Content, "Role content")
}

func TestFindSection_Missing(t *testing.T) {
	section, ok := FindSection(nil, "Role")
	assert.False(t, ok)
	assert.Empty(t, section)

	sections, err := ExtractSections([]byte("## Other\nbody\n"))
	require.NoError(t, err)
	_, ok = FindSection(sections, "Role")
	assert.False(t, ok)
}

func TestExtractSections_LargeFile(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 10_000; i++ {
		b.WriteString("line of content\n")
	}
	b.WriteString("## Tail\nend\n")

	start := time.Now()
	sections, err := ExtractSections([]byte(b.String()))
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.NotEmpty(t, sections)
	assert.Less(t, elapsed, 100*time.Millisecond, "extract took %s", elapsed)
}
