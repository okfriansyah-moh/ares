package importer

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/config"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

// ErrUnknownSource is returned when the import source is not registered.
var ErrUnknownSource = errors.New("importer: unknown source")

// Registry maps import source names to Importer implementations.
type Registry struct {
	importers map[string]arslib.Importer
}

// DefaultRegistry holds built-in import sources.
var DefaultRegistry = NewRegistry()

func init() {
	DefaultRegistry.Register(&GitHubImporter{})
	DefaultRegistry.Register(&CursorImporter{})
	DefaultRegistry.Register(&ClaudeImporter{})
	DefaultRegistry.Register(&CodexImporter{})
}

// NewRegistry returns an empty import registry.
func NewRegistry() *Registry {
	return &Registry{importers: make(map[string]arslib.Importer)}
}

// Register adds an importer for its Source() name.
func (r *Registry) Register(im arslib.Importer) {
	r.importers[im.Source()] = im
}

// Get returns the importer for source.
func (r *Registry) Get(source string) (arslib.Importer, bool) {
	im, ok := r.importers[source]
	return im, ok
}

// Sources returns registered source names in sorted order.
func (r *Registry) Sources() []string {
	out := make([]string, 0, len(r.importers))
	for name := range r.importers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Import parses a provider artifact into a Repository.
func Import(root, source string) (*arslib.Repository, error) {
	im, ok := DefaultRegistry.Get(source)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownSource, source)
	}

	repo, err := im.Import(root)
	if err != nil {
		return nil, err
	}

	if err := validateRepositoryPaths(root, repo); err != nil {
		return nil, fmt.Errorf("importer: %w", err)
	}

	return repo, nil
}

// WriteRepository writes repo to .ai/ under root.
func WriteRepository(root string, repo *arslib.Repository, overwrite bool) (created int, conflicts int, err error) {
	if repo == nil {
		return 0, 0, fmt.Errorf("importer: repository is nil")
	}

	instructions := append([]arslib.Instruction(nil), repo.Instructions...)
	sort.Slice(instructions, func(i, j int) bool { return instructions[i].ID < instructions[j].ID })

	for _, inst := range instructions {
		rel := filepath.ToSlash(filepath.Join(".ai", "instructions", inst.ID+".md"))
		ok, err := writeIfAllowed(root, rel, []byte(inst.Content), overwrite)
		if err != nil {
			return created, conflicts, err
		}
		if ok {
			created++
		} else {
			conflicts++
		}
	}

	agents := append([]arslib.Agent(nil), repo.Agents...)
	sort.Slice(agents, func(i, j int) bool { return agents[i].ID < agents[j].ID })

	for _, agent := range agents {
		rel := filepath.ToSlash(filepath.Join(".ai", "agents", agent.ID, "AGENT.md"))
		ok, err := writeIfAllowed(root, rel, []byte(agent.Content), overwrite)
		if err != nil {
			return created, conflicts, err
		}
		if ok {
			created++
		} else {
			conflicts++
		}
	}

	skills := append([]arslib.Skill(nil), repo.Skills...)
	sort.Slice(skills, func(i, j int) bool { return skills[i].ID < skills[j].ID })

	for _, skill := range skills {
		rel := filepath.ToSlash(filepath.Join(".ai", "skills", skill.ID, "SKILL.md"))
		ok, err := writeIfAllowed(root, rel, []byte(skill.Content), overwrite)
		if err != nil {
			return created, conflicts, err
		}
		if ok {
			created++
		} else {
			conflicts++
		}
		for _, ef := range skill.ExtraFiles {
			efRel := filepath.ToSlash(filepath.Join(".ai", "skills", skill.ID, ef.Rel))
			ok, err := writeIfAllowed(root, efRel, ef.Content, overwrite)
			if err != nil {
				return created, conflicts, err
			}
			if ok {
				created++
			} else {
				conflicts++
			}
		}
	}

	prompts := append([]arslib.Prompt(nil), repo.Prompts...)
	sort.Slice(prompts, func(i, j int) bool { return prompts[i].ID < prompts[j].ID })

	for _, prompt := range prompts {
		rel := filepath.ToSlash(filepath.Join(".ai", "prompts", prompt.ID+".md"))
		ok, err := writeIfAllowed(root, rel, []byte(prompt.Content), overwrite)
		if err != nil {
			return created, conflicts, err
		}
		if ok {
			created++
		} else {
			conflicts++
		}
	}

	exists, err := safepath.Exists(root, ".ai/manifest.yaml")
	if err != nil {
		return created, conflicts, fmt.Errorf("importer: %w", err)
	}
	if exists && !overwrite {
		conflicts++
	} else {
		if err := config.Write(root, &repo.Manifest); err != nil {
			return created, conflicts, fmt.Errorf("importer: %w", err)
		}
		created++
	}

	return created, conflicts, nil
}

func writeIfAllowed(root, rel string, data []byte, overwrite bool) (bool, error) {
	if _, err := safepath.Join(root, strings.Split(rel, "/")...); err != nil {
		return false, err
	}
	exists, err := safepath.Exists(root, strings.TrimPrefix(filepath.ToSlash(rel), "/"))
	if err != nil {
		return false, err
	}
	if exists && !overwrite {
		return false, nil
	}
	if err := safepath.WriteFile(root, rel, data, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func validateRepositoryPaths(root string, repo *arslib.Repository) error {
	if repo == nil {
		return fmt.Errorf("repository is nil")
	}
	for _, agent := range repo.Agents {
		if _, err := safepath.Join(root, ".ai", "agents", agent.ID, "AGENT.md"); err != nil {
			return err
		}
	}
	for _, skill := range repo.Skills {
		if _, err := safepath.Join(root, ".ai", "skills", skill.ID, "SKILL.md"); err != nil {
			return err
		}
	}
	for _, inst := range repo.Instructions {
		if _, err := safepath.Join(root, ".ai", "instructions", inst.ID+".md"); err != nil {
			return err
		}
	}
	for _, prompt := range repo.Prompts {
		if _, err := safepath.Join(root, ".ai", "prompts", prompt.ID+".md"); err != nil {
			return err
		}
	}
	return nil
}
