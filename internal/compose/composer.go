package compose

import (
	"errors"
	"fmt"
	"sort"

	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
)

// ErrUnknownTarget is returned when the compose target is not registered.
var ErrUnknownTarget = errors.New("compose: unknown target")

// Registry maps compose target names to Composer implementations.
type Registry struct {
	composers map[string]arslib.Composer
}

// DefaultRegistry holds built-in compose targets.
var DefaultRegistry = NewRegistry()

func init() {
	DefaultRegistry.Register(&CursorComposer{})
	DefaultRegistry.Register(&CopilotComposer{})
	DefaultRegistry.Register(&ClaudeComposer{})
	DefaultRegistry.Register(&CodexComposer{})
}

// NewRegistry returns an empty compose registry.
func NewRegistry() *Registry {
	return &Registry{composers: make(map[string]arslib.Composer)}
}

// Register adds a composer for its Target() name.
func (r *Registry) Register(c arslib.Composer) {
	r.composers[c.Target()] = c
}

// Get returns the composer for target.
func (r *Registry) Get(target string) (arslib.Composer, bool) {
	c, ok := r.composers[target]
	return c, ok
}

// Targets returns registered target names in sorted order.
func (r *Registry) Targets() []string {
	out := make([]string, 0, len(r.composers))
	for name := range r.composers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Compose runs the named target against repo under root.
func Compose(root, target string, repo *arslib.Repository) error {
	c, ok := DefaultRegistry.Get(target)
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownTarget, target)
	}
	if err := c.Compose(root, repo); err != nil {
		return err
	}
	if _, err := safepath.Join(root, outputPathForTarget(target)); err != nil {
		return fmt.Errorf("compose: %w", err)
	}
	return nil
}

func outputPathForTarget(target string) string {
	switch target {
	case "cursor":
		return ".cursor"
	case "copilot":
		return ".github"
	case "claude":
		return "CLAUDE.md"
	case "codex":
		return "AGENTS.md"
	default:
		return target
	}
}
