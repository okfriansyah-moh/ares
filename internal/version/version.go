package version

import "strings"

// Version is overridden by release builds using -ldflags.
var Version = "dev"

// String returns the CLI version label.
func String() string {
	if strings.HasPrefix(Version, "v") {
		return "ars " + Version
	}
	return "ars v" + Version
}
