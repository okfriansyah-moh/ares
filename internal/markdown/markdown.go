package markdown

import (
	"bytes"
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

	var sections []Section
	var current *Section
	var content bytes.Buffer

	flush := func() {
		if current == nil {
			return
		}
		current.Content = strings.TrimSpace(content.String())
		sections = append(sections, *current)
		content.Reset()
		current = nil
	}

	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		switch n := node.(type) {
		case *ast.Heading:
			flush()
			current = &Section{
				Heading: HeadingText(n, src),
				Level:   n.Level,
			}
		default:
			if current == nil {
				current = &Section{}
			}
			content.Write(nodeSource(node, src))
		}
	}

	flush()
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

func nodeSource(node ast.Node, src []byte) []byte {
	switch node.Kind() {
	case ast.KindText, ast.KindString, ast.KindCodeSpan, ast.KindEmphasis, ast.KindLink:
		return nil
	}

	lines := node.Lines()
	if lines != nil && lines.Len() > 0 {
		var buf bytes.Buffer
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			buf.Write(src[seg.Start:seg.Stop])
		}
		return buf.Bytes()
	}
	if !node.HasChildren() {
		return nil
	}
	var buf bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		buf.Write(nodeSource(child, src))
	}
	return buf.Bytes()
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
