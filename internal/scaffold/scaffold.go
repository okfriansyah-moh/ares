package scaffold

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/okfriansyah-moh/ares/internal/safepath"
)

//go:embed templates
var templates embed.FS

// ErrAlreadyInitialised is returned when .ai/ exists and Force is false.
var ErrAlreadyInitialised = errors.New("scaffold: .ai/ already initialised")

// Options configures ars init.
type Options struct {
	Root  string
	Force bool
}

type templateData struct {
	ProjectName string
}

// Run writes a valid .ai/ skeleton under opts.Root.
func Run(opts Options) error {
	root := filepath.Clean(opts.Root)
	if root == "" || root == "." {
		return fmt.Errorf("scaffold: root path is required")
	}

	if err := safepath.IsInsideRoot(root, ".ai"); err != nil {
		return fmt.Errorf("scaffold: %w", err)
	}

	exists, err := safepath.Exists(root, ".ai")
	if err != nil {
		return fmt.Errorf("scaffold: %w", err)
	}
	if exists {
		if !opts.Force {
			return ErrAlreadyInitialised
		}
		if err := safepath.RemoveAll(root, ".ai"); err != nil {
			return fmt.Errorf("scaffold: %w", err)
		}
	}

	data := templateData{ProjectName: filepath.Base(root)}

	err = fs.WalkDir(templates, "templates", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		rel, ok := strings.CutPrefix(path, "templates/")
		if !ok {
			return fmt.Errorf("scaffold: unexpected template path %q", path)
		}
		rel = strings.TrimSuffix(rel, ".tmpl")

		content, err := renderTemplate(path, data)
		if err != nil {
			return err
		}

		targetRel := filepath.Join(".ai", filepath.FromSlash(rel))
		if err := safepath.WriteFile(root, targetRel, content, 0o644); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Initialised .ai/ in %s\n", root)
	return nil
}

func renderTemplate(path string, data templateData) ([]byte, error) {
	raw, err := templates.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("scaffold: read template %q: %w", path, err)
	}

	tmpl, err := template.New(filepath.Base(path)).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("scaffold: parse template %q: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("scaffold: execute template %q: %w", path, err)
	}
	return buf.Bytes(), nil
}
