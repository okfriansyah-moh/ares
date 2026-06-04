package arslib

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_ZeroValueSafeToRange(t *testing.T) {
	var repo Repository

	assert.NotPanics(t, func() {
		for range repo.Instructions {
		}
		for range repo.Agents {
		}
		for range repo.Skills {
		}
		for range repo.Prompts {
		}
	})
}

func TestFindingLevel_String(t *testing.T) {
	tests := []struct {
		level FindingLevel
		want  string
	}{
		{level: OK, want: "OK"},
		{level: Warning, want: "Warning"},
		{level: Error, want: "Error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.level.String())
		})
	}
}

func TestFinding_JSONRoundTrip(t *testing.T) {
	original := Finding{
		Level:   Warning,
		Path:    ".ai/agents/planner/AGENT.md",
		Message: "missing ## Role",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Finding
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}
