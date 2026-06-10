package markdown

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Section is a markdown heading and its body content until the next heading.
type Section struct {
	Heading string
	Level   int
	Content string
}

// ExtractSections parses src and returns heading-to-body sections in document order.
func ExtractSections(src []byte) ([]Section, error) {
	if len(src) == 0 {
		return nil, nil
	}

	doc := goldmark.DefaultParser().Parse(text.NewReader(src))

	type heading struct {
		text  string
		level int
		start int
		end   int
	}

	var headings []heading
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		h, ok := node.(*ast.Heading)
		if !ok {
			continue
		}
		start, end := headingRange(h, src)
		headings = append(headings, heading{
			text:  HeadingText(h, src),
			level: h.Level,
			start: start,
			end:   end,
		})
	}

	if len(headings) == 0 {
		return []Section{{
			Content: strings.TrimSpace(string(src)),
		}}, nil
	}

	var sections []Section
	if preface := strings.TrimSpace(string(src[:headings[0].start])); preface != "" {
		sections = append(sections, Section{Content: preface})
	}

	for i, h := range headings {
		contentEnd := len(src)
		if i+1 < len(headings) {
			contentEnd = headings[i+1].start
		}
		sections = append(sections, Section{
			Heading: h.text,
			Level:   h.level,
			Content: strings.TrimSpace(string(src[h.end:contentEnd])),
		})
	}

	return sections, nil
}

// FindSection returns the first section whose heading matches heading case-insensitively.
func FindSection(sections []Section, heading string) (Section, bool) {
	want := strings.TrimSpace(heading)
	for _, s := range sections {
		if strings.EqualFold(strings.TrimSpace(s.Heading), want) {
			return s, true
		}
	}
	return Section{}, false
}

// HasSection reports whether a section with the given heading exists.
func HasSection(sections []Section, heading string) bool {
	_, ok := FindSection(sections, heading)
	return ok
}

// HeadingText returns plain text from a heading node with inline formatting removed.
func HeadingText(node ast.Node, src []byte) string {
	h, ok := node.(*ast.Heading)
	if !ok {
		return ""
	}

	var b strings.Builder
	for child := h.FirstChild(); child != nil; child = child.NextSibling() {
		b.WriteString(inlineText(child, src))
	}
	return strings.TrimSpace(b.String())
}

func inlineText(node ast.Node, src []byte) string {
	switch n := node.(type) {
	case *ast.Text:
		return string(n.Segment.Value(src))
	case *ast.String:
		return string(n.Value)
	case *ast.CodeSpan:
		var b strings.Builder
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			b.WriteString(inlineText(child, src))
		}
		return b.String()
	case *ast.Emphasis, *ast.Link:
		var b strings.Builder
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			b.WriteString(inlineText(child, src))
		}
		return b.String()
	default:
		if !node.HasChildren() {
			return ""
		}
		var b strings.Builder
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			b.WriteString(inlineText(child, src))
		}
		return b.String()
	}
}

func headingRange(node *ast.Heading, src []byte) (start int, end int) {
	lines := node.Lines()
	if lines == nil || lines.Len() == 0 {
		return 0, 0
	}

	first := lines.At(0)
	last := lines.At(lines.Len() - 1)
	return lineStart(src, first.Start), lineEnd(src, last.Stop)
}

func lineStart(src []byte, pos int) int {
	if pos < 0 {
		return 0
	}
	if pos > len(src) {
		pos = len(src)
	}
	for pos > 0 && src[pos-1] != '\n' {
		pos--
	}
	return pos
}

func lineEnd(src []byte, pos int) int {
	if pos < 0 {
		return 0
	}
	if pos > len(src) {
		pos = len(src)
	}
	for pos < len(src) && src[pos] != '\n' {
		pos++
	}
	if pos < len(src) {
		return pos + 1
	}
	return pos
}
