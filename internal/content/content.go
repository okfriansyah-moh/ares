package content

import "strings"

// HasBody reports whether a generated artifact has meaningful body content.
func HasBody(s string) bool {
	return strings.TrimSpace(s) != ""
}
