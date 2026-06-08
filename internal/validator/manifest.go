package validator

import (
	"path/filepath"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/config"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

var knownManifestVersions = map[string]struct{}{
	"2.0": {},
}

func validateManifest(root string) []arslib.Finding {
	manifestPath := ".ai/manifest.yaml"
	exists, err := safepath.Exists(root, manifestPath)
	if err != nil {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    manifestPath,
			Message: err.Error(),
		}}
	}

	if !exists {
		return []arslib.Finding{{
			Level:   arslib.Error,
			Path:    manifestPath,
			Message: "manifest.yaml does not exist",
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
	exists, err := safepath.Exists(root, filepath.ToSlash(filepath.Join(".ai", "agents", agentID, "AGENT.md")))
	if err != nil {
		return false
	}
	return exists
}
