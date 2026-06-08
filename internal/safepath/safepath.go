package safepath

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrPathEscape is returned when a resolved path leaves the repository root.
var ErrPathEscape = errors.New("path escapes root")

// ErrSymlink is returned when a path resolves to a symlink.
var ErrSymlink = errors.New("symlink rejected")

// Join joins path parts under root and rejects paths that escape root.
func Join(root string, parts ...string) (string, error) {
	root = filepath.Clean(root)
	if root == "" || root == "." {
		return "", fmt.Errorf("safepath: root path is required")
	}
	for _, part := range parts {
		if filepath.IsAbs(part) {
			return "", fmt.Errorf("safepath: %w: absolute path %q escapes root %q", ErrPathEscape, part, root)
		}
	}

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

// IsInsideRoot reports whether rel resolves inside root.
func IsInsideRoot(root, rel string) error {
	_, err := Join(root, splitRel(rel)...)
	return err
}

// ReadFile reads rel under root after rejecting path escapes and symlinks.
func ReadFile(root, rel string) ([]byte, error) {
	path, err := Join(root, splitRel(rel)...)
	if err != nil {
		return nil, err
	}
	if err := rejectSymlinkPath(root, path, true); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// WriteFile writes data to rel under root atomically.
func WriteFile(root, rel string, data []byte, perm os.FileMode) error {
	path, err := Join(root, splitRel(rel)...)
	if err != nil {
		return err
	}

	parent := filepath.Dir(path)
	if err := rejectSymlinkPath(root, parent, false); err != nil {
		return err
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	if err := rejectSymlinkPath(root, parent, true); err != nil {
		return err
	}

	if exists, err := Exists(root, rel); err != nil {
		return err
	} else if exists {
		if err := rejectSymlinkPath(root, path, true); err != nil {
			return err
		}
	}

	tmp, err := os.CreateTemp(parent, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		// Windows does not replace an existing destination on rename.
		if runtime.GOOS == "windows" {
			if rmErr := os.Remove(path); rmErr == nil {
				if err2 := os.Rename(tmpName, path); err2 == nil {
					cleanup = false
					return nil
				}
			}
		}
		return err
	}
	cleanup = false
	return nil
}

// MkdirAll creates rel under root after rejecting path escapes.
func MkdirAll(root, rel string, perm os.FileMode) error {
	path, err := Join(root, splitRel(rel)...)
	if err != nil {
		return err
	}
	if err := rejectSymlinkPath(root, path, false); err != nil {
		return err
	}
	return os.MkdirAll(path, perm)
}

// WalkDir walks subpath under root and skips symlinks.
func WalkDir(root, subpath string, fn fs.WalkDirFunc) error {
	path, err := Join(root, splitRel(subpath)...)
	if err != nil {
		return err
	}
	if err := rejectSymlinkPath(root, path, true); err != nil {
		return err
	}
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fn(p, d, walkErr)
		}
		if d.Type()&os.ModeSymlink != 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		return fn(p, d, nil)
	})
}

// ReadDir reads a directory under root after rejecting path escapes and symlinks.
func ReadDir(root, rel string) ([]os.DirEntry, error) {
	path, err := Join(root, splitRel(rel)...)
	if err != nil {
		return nil, err
	}
	if err := rejectSymlinkPath(root, path, true); err != nil {
		return nil, err
	}
	return os.ReadDir(path)
}

// Exists reports whether rel exists under root after rejecting path escapes.
func Exists(root, rel string) (bool, error) {
	path, err := Join(root, splitRel(rel)...)
	if err != nil {
		return false, err
	}
	if err := rejectSymlinkPath(root, filepath.Dir(path), false); err != nil {
		return false, err
	}
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else if err := rejectSymlinkPath(root, path, true); err != nil {
		return false, err
	}
	return true, nil
}

// RemoveAll removes rel under root after rejecting path escapes.
func RemoveAll(root, rel string) error {
	path, err := Join(root, splitRel(rel)...)
	if err != nil {
		return err
	}
	if path == filepath.Clean(root) {
		return fmt.Errorf("safepath: refusing to remove repository root")
	}
	if err := rejectSymlinkPath(root, filepath.Dir(path), false); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func rejectSymlinkPath(root, path string, requireFinal bool) error {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	if !strings.HasPrefix(path+string(os.PathSeparator), root+string(os.PathSeparator)) {
		return fmt.Errorf("safepath: %w: %q escapes root %q", ErrPathEscape, path, root)
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	if rel == "." {
		return rejectSymlink(root, requireFinal)
	}

	current := root
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) && !requireFinal {
				return nil
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("safepath: %w: %s", ErrSymlink, current)
		}
	}
	return nil
}

func rejectSymlink(path string, mustExist bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) && !mustExist {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("safepath: %w: %s", ErrSymlink, path)
	}
	return nil
}

func splitRel(rel string) []string {
	if filepath.IsAbs(rel) {
		return []string{rel}
	}
	rel = filepath.ToSlash(rel)
	if strings.TrimSpace(rel) == "" {
		return nil
	}
	return strings.Split(rel, "/")
}
