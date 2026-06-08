package safepath

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoin_ValidPath(t *testing.T) {
	root := t.TempDir()

	got, err := Join(root, ".ai", "manifest.yaml")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, ".ai", "manifest.yaml"), got)
}

func TestJoin_DotDotEscape(t *testing.T) {
	root := t.TempDir()

	_, err := Join(root, "..", "..", "etc", "passwd")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPathEscape))
}

func TestJoin_AbsoluteEscape(t *testing.T) {
	root := t.TempDir()

	_, err := Join(root, string(filepath.Separator)+"etc", "passwd")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPathEscape))
}

func TestReadFile_Symlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(outside, []byte("secret"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "link.txt")))

	_, err := ReadFile(root, "link.txt")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSymlink))
}

func TestReadFile_ParentSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}

	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, WriteFile(outside, "secret.txt", []byte("secret"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, ".ai")))

	_, err := ReadFile(root, ".ai/secret.txt")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSymlink))
}

func TestWriteFile_Atomic(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, WriteFile(root, "dir/file.txt", []byte("first"), 0o644))
	require.NoError(t, WriteFile(root, "dir/file.txt", []byte("second"), 0o644))

	data, err := ReadFile(root, "dir/file.txt")
	require.NoError(t, err)
	assert.Equal(t, "second", string(data))

	entries, err := os.ReadDir(filepath.Join(root, "dir"))
	require.NoError(t, err)
	for _, entry := range entries {
		assert.NotContains(t, entry.Name(), ".tmp")
	}
}

func TestWriteFile_PathEscape(t *testing.T) {
	root := t.TempDir()

	err := WriteFile(root, "../../evil.txt", []byte("evil"), 0o644)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrPathEscape))
}

func TestWalkDir_SkipsSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}

	root := t.TempDir()
	require.NoError(t, WriteFile(root, "real/file.txt", []byte("real"), 0o644))
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "linkdir")))

	var seen []string
	err := WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		require.NoError(t, err)
		rel, relErr := filepath.Rel(root, path)
		require.NoError(t, relErr)
		seen = append(seen, filepath.ToSlash(rel))
		return nil
	})
	require.NoError(t, err)
	assert.Contains(t, seen, "real/file.txt")
	assert.NotContains(t, seen, "linkdir/secret.txt")
}
