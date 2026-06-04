package safepath

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrPathEscape is returned when a resolved path leaves the repository root.
var ErrPathEscape = errors.New("path escapes root")

// Join joins path parts under root and rejects paths that escape root.
func Join(root string, parts ...string) (string, error) {
	root = filepath.Clean(root)
	joined := filepath.Join(append([]string{root}, parts...)...)
	cleaned := filepath.Clean(joined)
	rootClean := root
	if !strings.HasSuffix(rootClean, string(os.PathSeparator)) {
		rootClean += string(os.PathSeparator)
	}
	if !strings.HasPrefix(cleaned+string(os.PathSeparator), rootClean) {
		return "", fmt.Errorf("safepath: %w: %q escapes root %q", ErrPathEscape, joined, root)
	}
	return cleaned, nil
}

// IsInsideRoot reports whether rel resolves inside root (stub; full checks in Task 13).
func IsInsideRoot(root, rel string) error {
	_, err := Join(root, rel)
	return err
}

// WriteFile writes data to rel under root atomically (stub; hardened in Task 13).
func WriteFile(root, rel string, data []byte, perm os.FileMode) error {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	path, err := Join(root, parts...)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
