package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_DefaultDev(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "dev"
	assert.Equal(t, "ars vdev", String())
}

func TestString_ReleaseVersion(t *testing.T) {
	original := Version
	defer func() { Version = original }()

	Version = "v1.0.0"
	assert.Equal(t, "ars v1.0.0", String())
}
