package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/okfriansyah-moh/ares/internal/compose"
	"github.com/okfriansyah-moh/ares/internal/config"
	"github.com/okfriansyah-moh/ares/internal/importer"
	"github.com/okfriansyah-moh/ares/internal/markdown"
	"github.com/okfriansyah-moh/ares/internal/safepath"
	"github.com/okfriansyah-moh/ares/internal/scaffold"
	"github.com/okfriansyah-moh/ares/internal/validator"
	"github.com/okfriansyah-moh/ares/internal/version"
	"github.com/okfriansyah-moh/ares/pkg/arslib"
	"github.com/spf13/cobra"
)

var errValidationFailed = errors.New("validation failed")

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ars",
		Short:         "AI Repository Standard CLI",
		Version:       version.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newComposeCmd())
	cmd.AddCommand(newImportCmd())
	return cmd
}

func newInitCmd() *cobra.Command {
	var rootFlag string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold .ai/",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot(rootFlag)
			if err != nil {
				return err
			}
			return scaffold.Run(scaffold.Options{Root: root, Force: force})
		},
	}
	cmd.Flags().StringVar(&rootFlag, "root", "", "repository root")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing .ai/")
	return cmd
}

func newValidateCmd() *cobra.Command {
	var rootFlag string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate .ai/",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot(rootFlag)
			if err != nil {
				return err
			}

			findings, err := validator.Run(root)
			if err != nil {
				return err
			}

			if jsonOutput {
				data, err := json.MarshalIndent(findings, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
			} else {
				for _, finding := range findings {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", finding.Level.String(), finding.Path, finding.Message)
				}
			}

			if hasErrorFinding(findings) {
				return errValidationFailed
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&rootFlag, "root", "", "repository root")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print findings as JSON")
	return cmd
}

func newComposeCmd() *cobra.Command {
	var rootFlag string
	var target string

	cmd := &cobra.Command{
		Use:   "compose --target <target>",
		Short: "Compose provider artifacts from .ai/",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot(rootFlag)
			if err != nil {
				return err
			}
			repo, err := loadRepository(root)
			if err != nil {
				return err
			}
			if err := compose.Compose(root, target, repo); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Composed %s\n", target)
			return nil
		},
	}
	cmd.Flags().StringVar(&rootFlag, "root", "", "repository root")
	cmd.Flags().StringVar(&target, "target", "", "compose target")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

func newImportCmd() *cobra.Command {
	var rootFlag string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "import <source>",
		Short: "Import provider artifacts into .ai/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot(rootFlag)
			if err != nil {
				return err
			}
			repo, err := importer.Import(root, args[0])
			if err != nil {
				return err
			}
			created, conflicts, err := importer.WriteRepository(root, repo, overwrite)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %d files, %d conflicts skipped\n", created, conflicts)
			return nil
		},
	}
	cmd.Flags().StringVar(&rootFlag, "root", "", "repository root")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing .ai/ files")
	return cmd
}

func resolveRoot(flagValue string) (string, error) {
	if strings.TrimSpace(flagValue) == "" {
		return os.Getwd()
	}
	root, err := filepath.Abs(flagValue)
	if err != nil {
		return "", err
	}
	return root, nil
}

func hasErrorFinding(findings []arslib.Finding) bool {
	for _, finding := range findings {
		if finding.Level == arslib.Error {
			return true
		}
	}
	return false
}

func loadRepository(root string) (*arslib.Repository, error) {
	manifest, err := config.Load(root)
	if err != nil {
		return nil, err
	}

	repo := &arslib.Repository{Manifest: *manifest}
	if repo.Instructions, err = loadInstructions(root); err != nil {
		return nil, err
	}
	if repo.Agents, err = loadAgents(root); err != nil {
		return nil, err
	}
	if repo.Skills, err = loadSkills(root); err != nil {
		return nil, err
	}
	if repo.Prompts, err = loadPrompts(root); err != nil {
		return nil, err
	}
	return repo, nil
}

func loadInstructions(root string) ([]arslib.Instruction, error) {
	files, err := sortedFiles(root, ".ai", "instructions")
	if err != nil {
		return nil, err
	}

	var out []arslib.Instruction
	for _, name := range files {
		if filepath.Ext(name) != ".md" {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(".ai", "instructions", name))
		content, err := readRootFile(root, rel)
		if err != nil {
			return nil, err
		}
		out = append(out, arslib.Instruction{
			ID:      strings.TrimSuffix(name, filepath.Ext(name)),
			Path:    rel,
			Content: content,
		})
	}
	return out, nil
}

func loadAgents(root string) ([]arslib.Agent, error) {
	names, err := sortedDirs(root, ".ai", "agents")
	if err != nil {
		return nil, err
	}

	var out []arslib.Agent
	for _, name := range names {
		rel := filepath.ToSlash(filepath.Join(".ai", "agents", name, "AGENT.md"))
		content, err := readRootFile(root, rel)
		if err != nil {
			return nil, err
		}
		refs, err := agentSkillRefs(content)
		if err != nil {
			return nil, err
		}
		out = append(out, arslib.Agent{
			ID:        name,
			Path:      rel,
			Content:   content,
			SkillRefs: refs,
		})
	}
	return out, nil
}

func loadSkills(root string) ([]arslib.Skill, error) {
	names, err := sortedDirs(root, ".ai", "skills")
	if err != nil {
		return nil, err
	}

	var out []arslib.Skill
	for _, name := range names {
		rel := filepath.ToSlash(filepath.Join(".ai", "skills", name, "SKILL.md"))
		content, err := readRootFile(root, rel)
		if err != nil {
			return nil, err
		}
		refs, err := loadSkillReferences(root, name)
		if err != nil {
			return nil, err
		}
		out = append(out, arslib.Skill{
			ID:         name,
			Path:       rel,
			Content:    content,
			References: refs,
		})
	}
	return out, nil
}

func loadPrompts(root string) ([]arslib.Prompt, error) {
	files, err := sortedFiles(root, ".ai", "prompts")
	if err != nil {
		return nil, err
	}

	var out []arslib.Prompt
	for _, name := range files {
		if filepath.Ext(name) != ".md" {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(".ai", "prompts", name))
		content, err := readRootFile(root, rel)
		if err != nil {
			return nil, err
		}
		out = append(out, arslib.Prompt{
			ID:      strings.TrimSuffix(name, filepath.Ext(name)),
			Path:    rel,
			Content: content,
		})
	}
	return out, nil
}

func loadSkillReferences(root, skillID string) ([]string, error) {
	entries, err := safepath.ReadDir(root, filepath.ToSlash(filepath.Join(".ai", "skills", skillID, "references")))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	refs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		refs = append(refs, filepath.ToSlash(filepath.Join(".ai", "skills", skillID, "references", entry.Name())))
	}
	sort.Strings(refs)
	return refs, nil
}

func agentSkillRefs(content string) ([]string, error) {
	sections, err := markdown.ExtractSections([]byte(content))
	if err != nil {
		return nil, err
	}
	uses, ok := markdown.FindSection(sections, "Uses")
	if !ok {
		return nil, nil
	}

	var refs []string
	for _, line := range strings.Split(uses.Content, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "-*")
		line = strings.TrimSpace(strings.Trim(line, `"'`))
		if line != "" {
			refs = append(refs, line)
		}
	}
	return refs, nil
}

func sortedFiles(root string, parts ...string) ([]string, error) {
	if _, err := safepath.Join(root, parts...); err != nil {
		return nil, err
	}
	entries, err := safepath.ReadDir(root, filepath.ToSlash(filepath.Join(parts...)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func sortedDirs(root string, parts ...string) ([]string, error) {
	if _, err := safepath.Join(root, parts...); err != nil {
		return nil, err
	}
	entries, err := safepath.ReadDir(root, filepath.ToSlash(filepath.Join(parts...)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func readRootFile(root, rel string) (string, error) {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	data, err := safepath.ReadFile(root, filepath.ToSlash(filepath.Join(parts...)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
