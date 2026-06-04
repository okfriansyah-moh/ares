package validator

import (
	"os"
	"strings"

	"github.com/ars-standard/ars/internal/config"
	"github.com/ars-standard/ars/internal/safepath"
	"github.com/ars-standard/ars/pkg/arslib"
)

var knownManifestVersions = map[string]struct{}{
	"2.0": {},
}

func validateManifest(root string) []arslib.Finding {
	manifestPath := ".ai/manifest.yaml"
	absPath, err := safepath.Join(root, ".ai", "manifest.yaml")
	if err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    manifestPath,
			Message: err.Error(),
		}}
	}

	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return []arslib.Finding{{
				Level:   arslib.Error,
				Path:    manifestPath,
				Message: "manifest.yaml does not exist",
			}}
		}
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    manifestPath,
			Message: err.Error(),
		}}
	}

	m, err := config.Load(root)
	if err != nil {
		level := arslib.Error
		msg := err.Error()
		if strings.Contains(msg, "project.name is required") {
			msg = "project.name is required"
		} else if strings.Contains(msg, "version is required") {
			msg = "version is required"
		} else {
			msg = "manifest.yaml is unparseable"
		}
		return []arslib.Finding{{
			Level:   level,
			Path:    manifestPath,
			Message: msg,
		}}
	}

	var findings []arslib.Finding
	if _, ok := knownManifestVersions[m.Version]; !ok {
		findings = append(findings, arslib.Finding{
			Level:   arslib.Warning,
			Path:    manifestPath,
			Message: "unrecognised manifest version",
		})
	}

	if strings.TrimSpace(m.Defaults.Agent) == "" {
		findings = append(findings, arslib.Finding{
			Level:   arslib.Warning,
			Path:    manifestPath,
			Message: "defaults.agent is not set",
		})
	} else if !agentExists(root, m.Defaults.Agent) {
		findings = append(findings, arslib.Finding{
			Level:   arslib.Warning,
			Path:    manifestPath,
			Message: "defaults.agent references unknown agent",
		})
	}

	return findings
}

func agentExists(root, agentID string) bool {
	path, err := safepath.Join(root, ".ai", "agents", agentID, "AGENT.md")
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
