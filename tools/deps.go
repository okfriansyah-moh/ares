//go:build tools

package tools

import (
	_ "github.com/spf13/cobra"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/yuin/goldmark"
	_ "gopkg.in/yaml.v3"
)
