# PLAN.md — ARES Implementation Plan

> **Version:** 1.0
> **Date:** 2026-06-04
> **Author:** Core Team
> **Status:** Ready for Implementation
> **Source of Truth:** `docs/architecture.md`

---

## 1. Goal

Build **ARES** (`ars`), a single statically compiled Go binary that solves repository convention fragmentation across AI-assisted development tools. The binary exposes four commands:

- `ars init` — scaffold a valid `.ai/` skeleton in a repository
- `ars validate` — check `.ai/` structure, cross-references, and required sections
- `ars compose --target <cursor|copilot|claude|codex>` — translate `.ai/` into a provider-specific artifact
- `ars import <github|cursor|claude>` — read a provider artifact and produce an equivalent `.ai/`

The binary is distributed as a standalone executable (`go install`), a distroless container image, and eventually a Homebrew formula. No runtime dependencies. No database. No network calls in v1.

**Why:** Repository knowledge is fragmented across `.github/`, `.cursor/`, `CLAUDE.md`, and `AGENTS.md`. When teams switch tools or use multiple tools, knowledge must be rewritten or duplicated. ARES introduces `.ai/` as a single provider-agnostic source of truth that generates all provider artifacts.

---

## 2. Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                        User Repository                               │
│                                                                      │
│   ┌──────────────────────────────────────────────────────────┐      │
│   │                       .ai/                               │      │
│   │   manifest.yaml   instructions/   agents/                │      │
│   │   skills/         prompts/                               │      │
│   └─────────────────────────┬────────────────────────────────┘      │
│                              │                                        │
│         ┌──────────┬─────────┼─────────┬──────────┐                 │
│         │          │         │         │          │                 │
│     ars init   ars validate  │     ars import                        │
│         │          │     ars compose   │                             │
│         ▼          ▼         ▼         ▼                             │
│      .ai/      findings  provider    .ai/                            │
│     skeleton            artifacts   from                             │
│                         .cursor/    provider                         │
│                         .github/                                     │
│                         CLAUDE.md                                    │
│                         AGENTS.md                                    │
└──────────────────────────────────────────────────────────────────────┘
```

**Key architectural decisions (non-negotiable):**

| Decision | Rationale |
|---|---|
| Single static binary (`CGO_ENABLED=0`) | No runtime install; `go install` works out of the box; cross-platform |
| File-based only (no DB, no network in v1) | ARES operates on the repository filesystem; no state synchronisation problem |
| `Composer` and `Importer` interfaces | Open/Closed principle; add a provider without modifying caller code |
| Narrow `safepath` package for all I/O | Centralises path-traversal and symlink guards; every package imports it |
| `embed` stdlib for scaffold templates | Templates ship inside the binary; no external file deps at runtime |
| Deterministic output (idempotent compose) | Same `.ai/` input always produces byte-identical output; safe for CI diff checks |
| Distroless final image (`nonroot`) | Zero OS packages; no shell; eliminates the largest class of container CVEs |

---

## 3. Tech Stack

| Layer | Technology | Version | Rationale |
|---|---|---|---|
| Language | Go | 1.26 (latest stable) | Static binary; strong stdlib for file I/O and YAML; `go install` distribution |
| CLI framework | `github.com/spf13/cobra` | latest | De-facto standard; subcommands, help generation, shell completion |
| YAML parsing | `gopkg.in/yaml.v3` | latest | Strict struct mapping; supports depth limit for DoS prevention |
| Markdown parsing | `github.com/yuin/goldmark` | latest | AST-level parse for section extraction during import and validate |
| Embedded templates | `embed` (stdlib) | Go 1.16+ | `ars init` scaffold ships inside binary |
| Testing assertion | `github.com/stretchr/testify` | latest | Table-driven tests; `assert` / `require` helpers |
| Vulnerability scan | `govulncheck` | latest | Go-native vuln scanner against the Go vuln DB |
| Static analysis | `staticcheck` | latest | Additional linting beyond `go vet` |
| Container base | `gcr.io/distroless/static-debian12:nonroot` | latest | No shell; no package manager; runs as UID 65532 |
| Container builder | `golang:1.26-alpine` | latest | Multi-stage build; discarded in final image |
| CI | GitHub Actions | — | govulncheck + tests + container build on every PR |
| Build automation | `make` | system | `build`, `test`, `lint`, `vuln`, `docker-build` targets |

**Not used (and why):**

| Technology | Reason |
|---|---|
| PostgreSQL / any DB | No persistent state; all operations are file-based; no migrations needed |
| SvelteKit / any frontend | CLI is the interface; no GUI in v1 |
| Docker Compose | Single binary; no multi-service topology |
| gRPC / HTTP server | No network communication in v1 |
| Any ORM | No database |
| CGO | Static compilation requires `CGO_ENABLED=0`; CGO disabled everywhere |

---

## 4. Project Structure

```
ares/
├── cmd/
│   └── ars/
│       └── main.go                     ← cobra root + subcommand registration
├── internal/
│   ├── config/
│   │   ├── manifest.go                 ← parse and validate manifest.yaml
│   │   └── types.go                    ← Manifest, Project, Defaults structs
│   ├── safepath/
│   │   └── safepath.go                 ← path-traversal + symlink guards (shared)
│   ├── markdown/
│   │   └── markdown.go                 ← goldmark AST section extractor (shared)
│   ├── scaffold/
│   │   ├── scaffold.go                 ← ars init: write .ai/ from embedded FS
│   │   └── templates/                  ← embedded .ai/ skeleton files
│   │       ├── manifest.yaml.tmpl
│   │       └── instructions/README.md.tmpl
│   ├── validator/
│   │   ├── validator.go                ← Validator interface + runner
│   │   ├── manifest.go                 ← manifest.yaml validation rules
│   │   ├── agents.go                   ← AGENT.md required-section checks
│   │   ├── skills.go                   ← skill reference resolution
│   │   └── prompts.go                  ← prompt file existence checks
│   ├── compose/
│   │   ├── composer.go                 ← Composer interface + registry
│   │   ├── cursor.go                   ← .cursor/ output
│   │   ├── copilot.go                  ← .github/copilot-instructions.md output
│   │   ├── claude.go                   ← CLAUDE.md output
│   │   └── codex.go                    ← AGENTS.md output
│   └── importer/
│       ├── importer.go                 ← Importer interface + registry
│       ├── github.go                   ← import from .github/copilot-instructions.md
│       ├── cursor.go                   ← import from .cursor/rules/
│       └── claude.go                   ← import from CLAUDE.md
├── pkg/
│   └── arslib/
│       └── types.go                    ← public types: Repository, Agent, Skill, Prompt, Instruction
├── docs/
│   ├── architecture.md
│   ├── PLAN.md                         ← this file
│   └── decisions/
│       ├── ADR-0001-go-cli.md
│       ├── ADR-0002-markdown-as-source-format.md
│       └── ADR-0003-distroless-container.md
├── .ai/                                ← ARES's own .ai/ (reference implementation)
├── examples/
│   └── a2a-brainstormer/               ← example .ai/ for a real project
├── SPEC.md                             ← ARS v1 specification
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

---

## 5. Implementation Tasks

### Dependency Graph

```
Task 1 (Project Scaffold) ─────────────────────────────────────────────────┐
    │                                                                        │
    ▼                                                                        │
Task 2 (Domain Types — pkg/arslib) ──────────────────────────────────────── │
    │                                                                        │
    ├──────────────────────┐                                                 │
    ▼                      ▼                                                 │
Task 3 (Config)      Task 4 (Markdown Utility)                              │
    │                      │                                                 │
    └────────┬─────────────┘                                                 │
             ▼                                                               │
     ┌───────┴───────────────────────────────┐                              │
     │               │                       │                              │
     ▼               ▼                       ▼                              │
Task 5 (Scaffold) Task 6 (Validator)  Task 7 (Compose Infra + Cursor)       │
                                             │                              │
                                      ┌──────┴──────┐                      │
                                      ▼             ▼                      │
                               Task 8 (Copilot) Task 9 (Claude + Codex)   │
     │                                                                      │
     │           Task 10 (Importer Infra + GitHub)                         │
     │                  │                                                   │
     │           Task 11 (Cursor + Claude Importers)                       │
     │                                                                      │
     └───── All Tasks 5–11 ─────────────────────────────────────────────── │
                                  │                                         │
                                  ▼                                         │
                         Task 12 (CLI Wire-up) ─────────────────────────────
                                  │
                                  ▼
                         Task 13 (Security Hardening)
                                  │
                                  ▼
                         Task 14 (Container + Release)
                                  │
                         All Tasks 12–14
                                  │
                                  ▼
                         Task 15 (Integration Tests + SPEC.md + README.md)
                                  │
                                  ▼
                         Task 16 (Installation Script + GitHub Release)
```

---

### Task 1 — Project Scaffold ✅

**Goal:** Initialise the Go module, establish the full directory structure, write the Makefile, `.gitignore`, and verify the workspace compiles with no source files present.

**Files to create:**

- `go.mod` — module `github.com/ars-standard/ars`, Go 1.26; declare all deps (cobra, yaml.v3, goldmark, testify)
- `go.sum` — generated by `go mod tidy`
- `Makefile` — targets:
  - `build`: `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/ars ./cmd/ars`
  - `test`: `go test -race -count=1 ./...`
  - `lint`: `go vet ./... && staticcheck ./...`
  - `vuln`: `govulncheck ./...`
  - `docker-build`: `docker build -t ars:dev .`
  - `clean`: removes `bin/`
- `.gitignore` — `bin/`, `*.test`, `.env`, `coverage.out`
- Directory stubs (`.gitkeep`): `cmd/ars/`, `internal/config/`, `internal/safepath/`, `internal/markdown/`, `internal/scaffold/templates/`, `internal/validator/`, `internal/compose/`, `internal/importer/`, `pkg/arslib/`, `docs/decisions/`, `examples/`, `.ai/`
- `cmd/ars/main.go` — minimal `main()` that prints version and exits; no cobra yet

**Coding standards:**

- No global mutable state; `main()` is a thin entry point only
- `go.mod` pins Go 1.26 minimum — no older syntax permitted
- YAGNI: no feature flags, no build tags, no conditional imports

**Validation:**

- `go mod tidy`: zero errors, `go.sum` present
- `go build ./...`: zero errors (empty packages compile)
- `make build`: produces `bin/ars` binary

**Prompt context needed:** §8.5 (Dockerfile reference), §8.11 (Makefile targets)

---

### Task 2 — Domain Types (`pkg/arslib`) ✅

**Goal:** Define all public domain types shared across packages — the in-memory representation of a `.ai/` directory and the three core interfaces. These types are the single source of truth; no other package defines its own copy.

**Files to create:**

- `pkg/arslib/types.go`
  - `Manifest` struct: `Version string`, `Project Project`, `Defaults Defaults`
  - `Project` struct: `Name string`, `Description string`, `Repository string`
  - `Defaults` struct: `Agent string`
  - `Repository` struct: `Manifest Manifest`, `Instructions []Instruction`, `Agents []Agent`, `Skills []Skill`, `Prompts []Prompt`
  - `Instruction` struct: `ID string` (filename stem), `Path string`, `Content string`
  - `Agent` struct: `ID string` (directory name), `Path string`, `Content string`, `SkillRefs []string`
  - `Skill` struct: `ID string`, `Path string`, `Content string`, `References []string`
  - `Prompt` struct: `ID string`, `Path string`, `Content string`
  - `FindingLevel` type: `OK`, `Warning`, `Error` (iota + Stringer)
  - `Finding` struct: `Level FindingLevel`, `Path string`, `Message string`

- `pkg/arslib/interfaces.go`
  - `Composer` interface: `Compose(root string, repo *Repository) error`, `Target() string`
  - `Importer` interface: `Import(root string) (*Repository, error)`, `Source() string`
  - `Validator` interface: `Validate(root string) ([]Finding, error)`

- `pkg/arslib/types_test.go`
  - Table-driven: zero-value `Repository` is safe to range over (no nil pointer panics)
  - `FindingLevel.String()` returns human-readable label for all three levels
  - `Finding` JSON round-trip: marshal + unmarshal produces identical struct

**Coding standards:**

- SRP: types.go owns data; interfaces.go owns contracts — no mixed concerns
- No unexported fields that would silently fail JSON/YAML unmarshalling
- All slice fields initialised to `nil` (not `[]T{}`) — callers use `len()` not nil check
- Big O: all types are plain structs; O(1) construction, O(n) iteration — no hidden allocations

**Validation:**

- `go build ./pkg/arslib/...`: zero errors
- `go test -race -count=1 ./pkg/arslib/...`: all tests pass
- `go vet ./pkg/arslib/...`: zero issues
- `govulncheck ./pkg/arslib/...`: zero findings

**Prompt context needed:** §8.1 (Core Go Interfaces), §8.2 (.ai/ Schema)

---

### Task 3 — Config Package (`internal/config`) ✅

**Goal:** Parse and validate `manifest.yaml` into the `arslib.Manifest` type. Return typed errors that callers can distinguish (missing file vs. parse error vs. validation error).

**Files to create:**

- `internal/config/manifest.go`
  - `Load(root string) (*arslib.Manifest, error)` — resolves `{root}/.ai/manifest.yaml` via `safepath.Join` (guarded); reads file; unmarshals YAML with depth limit (see §8.4); validates required fields
  - `Validate(m *arslib.Manifest) error` — checks: `version` non-empty, `project.name` non-empty, `defaults.agent` is a non-empty string or absent (optional)
  - `Write(root string, m *arslib.Manifest) error` — marshals to YAML; writes atomically (tmp → rename)

- `internal/config/manifest_test.go`
  - `TestLoad_ValidManifest`: parses a well-formed `manifest.yaml` into correct struct
  - `TestLoad_MissingFile`: returns typed `*os.PathError`
  - `TestLoad_InvalidYAML`: returns parse error (not panic)
  - `TestLoad_MissingProjectName`: `Validate()` returns descriptive error
  - `TestWrite_Roundtrip`: `Write` then `Load` produces identical struct (idempotent)
  - `TestLoad_PathTraversal`: path `../../etc/passwd` rejected by safepath guard (see Task 13 — guard is a stub in this task, enforced in Task 13)

**Coding standards:**

- DRY: validation rules defined once in `Validate()`; not duplicated in `Load()`
- OCP: `Load` calls `Validate` internally; caller can also call `Validate` on a pre-built struct
- Error wrapping: use `fmt.Errorf("config: %w", err)` so callers can `errors.Is`/`errors.As`
- Big O: O(m) where m = manifest file size; YAML depth capped at 8 levels to prevent anchor-expansion DoS

**Validation:**

- `go build ./internal/config/...`: zero errors
- `go test -race -count=1 ./internal/config/...`: all tests pass
- `go vet ./internal/config/...`: zero issues
- `govulncheck ./internal/config/...`: zero findings

**Prompt context needed:** §8.2 (manifest.yaml schema), §8.4 (path security invariants), §8.7 (validation rules)

---

### Task 4 — Markdown Section Utility (`internal/markdown`) ✅

**Goal:** Shared utility that extracts named sections from a Markdown file using goldmark's AST. Used by both `validator` (checking required headings) and `importer` (extracting content by heading). Zero domain knowledge — operates on raw text only.

**Files to create:**

- `internal/markdown/markdown.go`
  - `Section` struct: `Heading string`, `Level int`, `Content string`
  - `ExtractSections(src []byte) ([]Section, error)` — walks goldmark AST; collects heading text → body content pairs; O(n) where n = source length
  - `FindSection(sections []Section, heading string) (Section, bool)` — case-insensitive linear scan; O(k) where k = number of sections
  - `HasSection(sections []Section, heading string) bool` — wraps `FindSection`
  - `HeadingText(node ast.Node, src []byte) string` — extracts plain text from a heading node, strips Markdown formatting

- `internal/markdown/markdown_test.go`
  - `TestExtractSections_Basic`: `## Role`, `## Responsibilities`, `## Uses` extracted correctly
  - `TestExtractSections_Nested`: H3 under H2 captured with correct Level
  - `TestExtractSections_Empty`: empty input returns empty slice (not error)
  - `TestExtractSections_NoHeadings`: body-only file returns single section with empty Heading
  - `TestFindSection_CaseInsensitive`: `## role` found when searching `"Role"`
  - `TestFindSection_Missing`: returns `false`; no panic on nil sections
  - `TestExtractSections_LargeFile`: 10 000-line file completes in < 100ms (regression guard)

**Coding standards:**

- SRP: section extraction has zero knowledge of `.ai/` structure or file paths
- KISS: no regex — goldmark AST walk is more robust and readable
- YAGNI: no front-matter parser, no link resolution, no image handling
- Big O: O(n) time, O(k) space where k = number of extracted sections

**Validation:**

- `go build ./internal/markdown/...`: zero errors
- `go test -race -count=1 ./internal/markdown/...`: all tests pass, including timing assertion
- `go vet ./internal/markdown/...`: zero issues
- `govulncheck ./internal/markdown/...`: zero findings

**Prompt context needed:** §8.1 (Core interfaces), §8.2 (AGENT.md required sections), §8.8 (compose algorithm — sections used during import)

---

### Task 5 — Scaffold Package (`ars init`) ✅

**Goal:** Implement `ars init` — write a valid `.ai/` skeleton from embedded templates into the repository root. Abort if `.ai/` already exists unless `--force` is passed.

**Files to create:**

- `internal/scaffold/scaffold.go`
  - `//go:embed templates` embed directive; `var templates embed.FS`
  - `Options` struct: `Root string`, `Force bool`
  - `Run(opts Options) error` — guards:
    1. `safepath.IsInsideRoot(opts.Root, ".ai")` (stub in this task; enforced in Task 13)
    2. Check `.ai/` does not exist; return `ErrAlreadyInitialised` if present and `!opts.Force`
    3. Walk embedded `templates/` FS; write each file to the target path
    4. Write `manifest.yaml` with `project.name` inferred from `filepath.Base(opts.Root)`
    5. Print "Initialised .ai/ in {root}" to stdout
  - `ErrAlreadyInitialised` sentinel error

- `internal/scaffold/templates/manifest.yaml.tmpl` — minimal manifest template
- `internal/scaffold/templates/instructions/README.md.tmpl` — one-line placeholder
- `internal/scaffold/scaffold_test.go`
  - `TestRun_FreshDirectory`: creates correct file tree; `manifest.yaml` parseable by `config.Load`
  - `TestRun_AlreadyExists`: returns `ErrAlreadyInitialised` without modifying existing files
  - `TestRun_Force`: overwrites `.ai/` when `Force: true`
  - `TestRun_ProjectNameInferred`: `project.name` equals `filepath.Base(root)` in manifest
  - All tests use `t.TempDir()` — no writes to real filesystem

**Coding standards:**

- DRY: template rendering via `text/template` — no hand-built string concatenation
- YAGNI: no interactive prompts, no config wizard, no multi-step init
- SRP: `scaffold` writes files; `config` validates the result — two separate concerns
- Big O: O(t) where t = number of template files; currently O(1), bounded by embedded FS size

**Validation:**

- `go build ./internal/scaffold/...`: zero errors
- `go test -race -count=1 ./internal/scaffold/...`: all tests pass
- `go vet ./internal/scaffold/...`: zero issues
- Smoke: `go run ./cmd/ars init --root /tmp/test-repo` → `.ai/manifest.yaml` present and valid

**Prompt context needed:** §8.2 (.ai/ schema), §8.4 (path security), §8.11 (Definition of Done checklist)

---

### Task 6 — Validator (`ars validate`) ✅

**Goal:** Implement `ars validate` — report all structural errors in a `.ai/` directory with file path, level, and message. Exit code 0 if no Errors; exit code 1 if any Error.

**Files to create:**

- `internal/validator/validator.go`
  - `Run(root string) ([]Finding, error)` — orchestrates all sub-validators in a fixed order; collects all findings; does not abort early; returns combined slice sorted by `Path` then `Level` (O(n log n) sort for deterministic output)
  - `levelString(l FindingLevel) string` — for CLI output formatting

- `internal/validator/manifest.go`
  - `validateManifest(root string) []Finding` — calls `config.Load`; reports: missing file (Error), parse failure (Error), missing `project.name` (Error), unknown `version` (Warning), missing `defaults.agent` (Warning)

- `internal/validator/agents.go`
  - `validateAgents(root string) []Finding` — walks `agents/*/AGENT.md`; for each: file exists (Error if missing), required sections present via `markdown.HasSection` — `## Role`, `## Responsibilities`, `## Uses`, `## Boundaries` (Error per missing section); extract `## Uses` refs; check each referenced skill exists (Error if not)

- `internal/validator/skills.go`
  - `validateSkills(root string) []Finding` — walks `skills/*/SKILL.md`; for each: file exists (Error if missing)

- `internal/validator/prompts.go`
  - `validatePrompts(root string) []Finding` — walks `prompts/*.md`; for each: file exists (Error if missing); `## Use` section present (Warning if missing)

- `internal/validator/validator_test.go`
  - Table-driven: one sub-test per finding type; each sets up a `t.TempDir()` `.ai/` with the specific defect
  - `TestRun_ValidTree`: well-formed `.ai/` returns zero Error-level findings
  - `TestRun_MissingManifest`: returns Error for missing `manifest.yaml`
  - `TestRun_MissingAgentSection`: returns Error for each missing required heading
  - `TestRun_BrokenSkillRef`: agent references `skills/foo/SKILL.md`; foo does not exist → Error
  - `TestRun_DeterministicOrder`: same input always returns findings in identical order (run twice, compare)
  - `TestRun_PathTraversalInSkillRef`: skill reference containing `..` rejected as Error

**Coding standards:**

- OCP: adding a new validation rule = adding to the relevant sub-file; `validator.go` is not modified
- ISP: each sub-validator is a standalone function, not a method on a shared struct
- Big O: O(f) time and O(f) space where f = number of files in `.ai/`; finding sort is O(n log n)
- No `panic` in validation path; all errors converted to `Finding` entries

**Validation:**

- `go build ./internal/validator/...`: zero errors
- `go test -race -count=1 ./internal/validator/...`: all tests pass
- `go vet ./internal/validator/...`: zero issues
- `govulncheck ./internal/validator/...`: zero findings
- Smoke: `go run ./cmd/ars validate --root /tmp/test-repo` (after `ars init`) → exit 0, zero error findings

**Prompt context needed:** §8.2 (AGENT.md required sections), §8.7 (full validation rules), §8.4 (path security)

---

### Task 7 — Compose Infrastructure + Cursor Target ✅

**Goal:** Define the `Composer` interface, the global compose registry, and implement the first provider target (`cursor`). The registry pattern ensures `ars compose` can dispatch to any target without a switch statement in the caller.

**Files to create:**

- `internal/compose/composer.go`
  - `Registry` struct: `map[string]arslib.Composer`; methods `Register(c arslib.Composer)`, `Get(target string) (arslib.Composer, bool)`, `Targets() []string` (sorted, O(k log k))
  - `DefaultRegistry` package-level var; `init()` registers all built-in composers
  - `Compose(root, target string, repo *arslib.Repository) error` — looks up registry; returns `ErrUnknownTarget` if not found; calls `Compose`; validates output is inside root via `safepath`

- `internal/compose/cursor.go`
  - `CursorComposer` implements `arslib.Composer`
  - `Target() string` returns `"cursor"`
  - `Compose(root string, repo *arslib.Repository) error`:
    1. Resolve output dir: `safepath.Join(root, ".cursor")` (see §8.4)
    2. Clear and recreate `.cursor/rules/` and `.cursor/prompts/`
    3. `instructions/*.md` → `.cursor/rules/<name>.mdc` with `---\ntype: always\n---\n` front matter + content
    4. For each agent: inline referenced skill content into the agent's rule file; write to `.cursor/rules/<agent-id>.mdc` with `type: agent-requested`
    5. `prompts/*.md` → `.cursor/prompts/<name>.prompt` (verbatim, no transformation)
    6. Write `manifest.yaml` `project.name` as a header comment in the first rule file
    7. All writes via `safepath.WriteFile` (see Task 13)
  - `cursorRuleHeader(agentType string) string` — pure function, testable in isolation

- `internal/compose/cursor_test.go`
  - `TestCursorComposer_BasicOutput`: minimal `Repository` → correct `.cursor/rules/` file tree
  - `TestCursorComposer_SkillInlined`: agent with one skill ref → skill content appears in rule file
  - `TestCursorComposer_NoPrompts`: repository with zero prompts → `.cursor/prompts/` exists but is empty
  - `TestCursorComposer_Idempotent`: run twice → identical output (byte-for-byte, using `filepath.WalkDir` checksum)
  - `TestCursorComposer_PathTraversal`: agent ID containing `../` is sanitised, not written outside root
  - All tests use `t.TempDir()`

**Coding standards:**

- OCP: adding a new target = add one file implementing `arslib.Composer`; `composer.go` never changes
- DRY: `cursorRuleHeader` is a pure function; reused by all rule-writing paths
- KISS: flat `.cursor/rules/` directory; no sub-folders; maps directly to cursor's expected layout
- Big O: O(a × s) where a = agents, s = average skills per agent; acceptable for typical repo sizes (<100 agents, <20 skills)

**Validation:**

- `go build ./internal/compose/...`: zero errors
- `go test -race -count=1 ./internal/compose/...`: all tests pass
- `go vet ./internal/compose/...`: zero issues
- `govulncheck ./internal/compose/...`: zero findings
- Smoke: compose against `.ai/` in this repo → `.cursor/` written correctly

**Prompt context needed:** §8.3 (provider mapping table), §8.4 (path security), §8.8 (compose algorithm)

---

### Task 8 — Copilot Composer ✅

**Goal:** Implement the `copilot` compose target — produce `.github/copilot-instructions.md` from `.ai/`.

**Files to create:**

- `internal/compose/copilot.go`
  - `CopilotComposer` implements `arslib.Composer`
  - `Target() string` returns `"copilot"`
  - `Compose(root string, repo *arslib.Repository) error`:
    1. Output path: `safepath.Join(root, ".github", "copilot-instructions.md")`
    2. Build output via `strings.Builder` (single allocation; no intermediate files):
       - Header comment: `<!-- Generated by ars compose --target copilot. Source: .ai/ -->`
       - `project.name` as H1 title
       - All `instructions/*.md` content under `## Repository Instructions`
       - For each agent: `## Agent: {name}` section with responsibilities + inlined skills
    3. Ensure `.github/` exists; write atomically (tmp → rename)
  - `copilotAgentSection(agent arslib.Agent, skills map[string]arslib.Skill) string` — pure function

- `internal/compose/copilot_test.go`
  - `TestCopilotComposer_HeaderPresent`: generated file contains the `<!-- Generated by ars -->` comment
  - `TestCopilotComposer_AllAgentsIncluded`: N agents in repo → N `## Agent:` sections in output
  - `TestCopilotComposer_SkillInlined`: skill content appears under the agent section that references it
  - `TestCopilotComposer_Idempotent`: run twice → identical `.github/copilot-instructions.md`
  - `TestCopilotComposer_EmptyInstructions`: no instructions files → only header + agent sections
  - All tests use `t.TempDir()`

**Coding standards:**

- DRY: `copilotAgentSection` is a pure function shared by `Build` and potentially future diff-check logic
- KISS: single output file; no sub-directory structure; mirrors how Copilot actually reads the file
- SRP: `copilot.go` only knows how to write Copilot format; parsing/reading is handled upstream
- Big O: O(a × s + i) where a = agents, s = skills, i = instruction bytes; O(n) space for `strings.Builder`

**Validation:**

- `go build ./internal/compose/...`: zero errors
- `go test -race -count=1 ./internal/compose/...`: all tests pass (all composers)
- `go vet ./internal/compose/...`: zero issues
- `govulncheck ./internal/compose/...`: zero findings

**Prompt context needed:** §8.3 (provider mapping table — copilot row), §8.4 (path security)

---

### Task 9 — Claude + Codex Composers ✅

**Goal:** Implement the `claude` and `codex` compose targets. Both are single-file outputs similar to `copilot`; they are batched into one task because their structures differ only in output filename and minor formatting conventions.

**Files to create:**

- `internal/compose/claude.go`
  - `ClaudeComposer` implements `arslib.Composer`
  - `Target() string` returns `"claude"`
  - `Compose(root string, repo *arslib.Repository) error` — writes `CLAUDE.md` at repo root
  - Format: project name H1, `<!-- ars:source .ai/ -->` marker, instructions sections, agent sections with inlined skills
  - Claude-specific: agent sections use `## {agent-id}` (lowercase); skills under `### Context: {skill-id}`

- `internal/compose/codex.go`
  - `CodexComposer` implements `arslib.Composer`
  - `Target() string` returns `"codex"`
  - `Compose(root string, repo *arslib.Repository) error` — writes `AGENTS.md` at repo root
  - Format: matches OpenAI Codex AGENTS.md convention; agent blocks use YAML-like headers

- `internal/compose/claude_test.go`
  - `TestClaudeComposer_OutputFilename`: output is `CLAUDE.md` at root, not inside a subdirectory
  - `TestClaudeComposer_SourceMarker`: `<!-- ars:source .ai/ -->` present (needed for future `ars import claude`)
  - `TestClaudeComposer_Idempotent`
  - `TestClaudeComposer_PathTraversal`: agent ID `../evil` → error, no file written outside root

- `internal/compose/codex_test.go`
  - `TestCodexComposer_OutputFilename`: output is `AGENTS.md` at root
  - `TestCodexComposer_SourceMarker`: `<!-- ars:source .ai/ -->` present
  - `TestCodexComposer_Idempotent`

**Coding standards:**

- DRY: common pattern (header comment, title, instructions, agents) extracted into a shared `buildMarkdownOutput(format composerFormat, repo *arslib.Repository) string` pure function in `compose/shared.go` — claude and codex both call it with different `composerFormat` config
- YAGNI: no provider-specific metadata beyond what the provider actually uses
- OCP: `composerFormat` is a data struct, not a sub-interface — behaviour stays in `buildMarkdownOutput`
- Big O: O(a × s + i) same as copilot; all string building via `strings.Builder`

**Validation:**

- `go build ./internal/compose/...`: zero errors
- `go test -race -count=1 ./internal/compose/...`: all composer tests pass
- `go vet ./internal/compose/...`: zero issues
- `govulncheck ./internal/compose/...`: zero findings

**Prompt context needed:** §8.3 (provider mapping table — claude and codex rows), §8.4 (path security)

---

### Task 10 — Importer Infrastructure + GitHub Source ✅

**Goal:** Define the `Importer` interface, the global import registry, and implement the first import source (`github` — reads `.github/copilot-instructions.md`). The import flow is the inverse of compose: parse a provider file, infer `.ai/` structure, write it out.

**Files to create:**

- `internal/importer/importer.go`
  - `Registry` struct: same pattern as `compose.Registry`; `Register`, `Get`, `Sources() []string`
  - `DefaultRegistry` package-level var; `init()` registers built-in importers
  - `Import(root, source string) (*arslib.Repository, error)` — looks up registry; calls `Import`; validates output paths via `safepath`
  - `WriteRepository(root string, repo *arslib.Repository, overwrite bool) (created int, conflicts int, error)` — writes `.ai/` from an in-memory `Repository`; skip existing files when `!overwrite`; report counts; all writes via `safepath.WriteFile`

- `internal/importer/github.go`
  - `GitHubImporter` implements `arslib.Importer`
  - `Source() string` returns `"github"`
  - `Import(root string) (*arslib.Repository, error)`:
    1. Guard: `safepath.Join(root, ".github", "copilot-instructions.md")`
    2. Read file; call `markdown.ExtractSections`
    3. Heuristic section classification (see §8.9):
       - Sections whose headings match `/^agent[:\s]/i` → `agents/<slug>/AGENT.md`
       - Sections whose headings match `/^skill[:\s]/i` → `skills/<slug>/SKILL.md`
       - All other top-level sections → `instructions/<slug>.md`
    4. Infer `manifest.yaml` from detected project name (look for H1 or first non-heading line)
    5. Return `*arslib.Repository`; caller writes via `WriteRepository`
  - `slugify(heading string) string` — pure: lowercase, replace spaces/special chars with `-`, max 50 chars, no leading/trailing `-`

- `internal/importer/github_test.go`
  - `TestGitHubImporter_BasicParse`: copilot-instructions.md with one agent section → `Repository.Agents` has one entry
  - `TestGitHubImporter_InstructionSections`: non-agent content → `Repository.Instructions`
  - `TestGitHubImporter_MissingFile`: returns descriptive error (not panic)
  - `TestGitHubImporter_SlugCollision`: two sections with same slug → second is `<slug>-2` (no silent overwrite)
  - `TestWriteRepository_SkipExisting`: existing file not overwritten when `overwrite=false`; `conflicts` count correct
  - `TestWriteRepository_Overwrite`: existing file replaced when `overwrite=true`

**Coding standards:**

- OCP: `importer.go` never changes when a new source is added
- KISS: heuristic classification is regex-based and transparent; no ML, no scoring
- DRY: `slugify` is a package-level pure function shared by all importers
- Big O: O(n) parse where n = file size; O(s) section classification where s = section count; O(k) write where k = output files

**Validation:**

- `go build ./internal/importer/...`: zero errors
- `go test -race -count=1 ./internal/importer/...`: all tests pass
- `go vet ./internal/importer/...`: zero issues
- `govulncheck ./internal/importer/...`: zero findings
- Smoke: run `ars import github` on a repo with a `.github/copilot-instructions.md` → `.ai/` written

**Prompt context needed:** §8.9 (import algorithm), §8.3 (provider mapping — github row), §8.4 (path security)

---

### Task 11 — Cursor + Claude Importers ✅

**Goal:** Implement the `cursor` and `claude` import sources. Cursor reads `.cursor/rules/*.mdc`; Claude reads `CLAUDE.md` (identified by the `<!-- ars:source .ai/ -->` marker inserted by Task 9).

**Files to create:**

- `internal/importer/cursor.go`
  - `CursorImporter` implements `arslib.Importer`
  - `Source() string` returns `"cursor"`
  - `Import(root string) (*arslib.Repository, error)`:
    1. Walk `safepath.Join(root, ".cursor", "rules")` for `.mdc` files
    2. For each file: parse YAML front matter (`type: agent-requested` → agent; `type: always` → instruction)
    3. Strip front matter; remaining content is the rule body
    4. `type: agent-requested` files: create `Agent` with ID from filename stem
    5. `type: always` files: create `Instruction` with ID from filename stem

- `internal/importer/claude.go`
  - `ClaudeImporter` implements `arslib.Importer`
  - `Source() string` returns `"claude"`
  - `Import(root string) (*arslib.Repository, error)`:
    1. Read `safepath.Join(root, "CLAUDE.md")`
    2. Check for `<!-- ars:source .ai/ -->` marker; if absent, warn (not error — still try to import)
    3. Extract sections via `markdown.ExtractSections`; same heuristic classification as GitHub importer
    4. Return `*arslib.Repository`

- `internal/importer/cursor_test.go`
  - `TestCursorImporter_AgentRule`: `.mdc` with `type: agent-requested` → `Repository.Agents` entry
  - `TestCursorImporter_InstructionRule`: `.mdc` with `type: always` → `Repository.Instructions` entry
  - `TestCursorImporter_EmptyRulesDir`: no `.mdc` files → empty `Repository` (not error)
  - `TestCursorImporter_FrontMatterStripped`: agent content does not contain the YAML front matter

- `internal/importer/claude_test.go`
  - `TestClaudeImporter_WithMarker`: `CLAUDE.md` with `<!-- ars:source .ai/ -->` → clean import
  - `TestClaudeImporter_WithoutMarker`: `CLAUDE.md` without marker → imports with Warning finding
  - `TestClaudeImporter_MissingFile`: returns error

**Coding standards:**

- DRY: section classification heuristic shared with GitHub importer via `internal/importer/classify.go` (pure function); do not duplicate the regex
- KISS: cursor front matter parsed with a simple `strings.SplitN` on `---` boundaries; no YAML library needed for 2-field front matter
- Big O: O(r) where r = number of rule files for cursor; O(n) for claude where n = CLAUDE.md size

**Validation:**

- `go build ./internal/importer/...`: zero errors
- `go test -race -count=1 ./internal/importer/...`: all tests pass
- `go vet ./internal/importer/...`: zero issues
- `govulncheck ./internal/importer/...`: zero findings
- Round-trip smoke: `ars compose --target cursor` then `ars import cursor` on the same repo → no information lost in agent sections

**Prompt context needed:** §8.9 (import algorithm), §8.3 (provider mapping), §8.4 (path security)

---

### Task 12 — CLI Wire-up (`cmd/ars`) ✅

**Goal:** Wire all packages into the Cobra command tree. Each subcommand is a thin adapter: parse flags → call internal package → format output → set exit code. No business logic in `cmd/`.

**Files to create / modify:**

- `cmd/ars/main.go` — full implementation replacing the Task 1 stub:
  - `rootCmd` with `--version` flag (embed at build time via `-ldflags "-X main.version=..."`)
  - `initCmd` — `ars init [--root <path>] [--force]` → calls `scaffold.Run`
  - `validateCmd` — `ars validate [--root <path>] [--json]` → calls `validator.Run`; `--json` prints findings as JSON array; exit 1 on any Error-level finding
  - `composeCmd` — `ars compose --target <target> [--root <path>]` → loads repo via `config.Load` + file walk, calls `compose.Compose`
  - `importCmd` — `ars import <source> [--root <path>] [--overwrite]` → calls `importer.Import` + `importer.WriteRepository`
  - All commands: print to `stdout`; errors to `stderr`; never `os.Exit` directly — return errors to cobra's `RunE` which sets exit code

- `cmd/ars/main_test.go`
  - Integration tests using `cobra/cmd.Execute()` with a temporary working directory
  - `TestInit_CreatesAIDir`: `ars init` creates `.ai/manifest.yaml`
  - `TestValidate_ExitZeroOnValid`: valid `.ai/` → exit 0
  - `TestValidate_ExitOneOnError`: missing agent section → exit 1
  - `TestCompose_CursorTarget`: compose produces `.cursor/rules/`
  - `TestImport_GitHubSource`: import from mock `.github/copilot-instructions.md` → `.ai/` files
  - `TestVersion_Flag`: `ars --version` prints version string (does not panic on missing ldflags)

**Coding standards:**

- SRP: `cmd/` contains only flag parsing and output formatting; zero business logic
- KISS: `RunE` returns error; cobra handles `os.Exit(1)` — no manual exit code management
- DRY: `--root` flag resolved in a shared `resolveRoot(flagValue string) (string, error)` helper; reused by all four commands
- YAGNI: no `--config` file, no environment variable overrides in v1

**Validation:**

- `go build ./cmd/ars/...`: zero errors; produces `ars` binary
- `go test -race -count=1 ./cmd/ars/...`: all integration tests pass
- `go vet ./...`: zero issues across entire module
- `govulncheck ./...`: zero findings across entire module
- `staticcheck ./...`: zero issues
- `ars --help`: shows all four subcommands

**Prompt context needed:** §8.1 (Core interfaces), §8.11 (Definition of Done checklist), §8.7 (validate exit codes)

---

### Task 13 — Security Hardening (`internal/safepath`) ✅

**Goal:** Implement the shared security package that enforces path-traversal prevention, symlink rejection, and safe file write semantics. Retrofit all existing packages to use it. Every file-read and file-write operation in the codebase must go through this package.

**Files to create / modify:**

- `internal/safepath/safepath.go`
  - `Join(root string, parts ...string) (string, error)` — joins and cleans path; returns error if result escapes `root` (i.e., `!strings.HasPrefix(cleaned, filepath.Clean(root)+string(os.PathSeparator))`)
  - `ReadFile(root, path string) ([]byte, error)` — calls `Join`; calls `os.Lstat` (not `os.Stat`) to detect symlinks; rejects symlinks with `ErrSymlink`; calls `os.ReadFile` only after validation
  - `WriteFile(root, path string, data []byte, perm os.FileMode) error` — calls `Join`; calls `os.MkdirAll` on parent; writes to a temporary file in the same directory; renames atomically
  - `MkdirAll(root, path string, perm os.FileMode) error` — guarded `os.MkdirAll`
  - `WalkDir(root, subpath string, fn fs.WalkDirFunc) error` — guarded `filepath.WalkDir`; skips symlinks silently
  - `ErrPathEscape` sentinel error
  - `ErrSymlink` sentinel error

- `internal/safepath/safepath_test.go`
  - `TestJoin_ValidPath`: normal path returns cleaned result inside root
  - `TestJoin_DotDotEscape`: `../../etc/passwd` → `ErrPathEscape`
  - `TestJoin_AbsoluteEscape`: absolute path `/etc/passwd` → `ErrPathEscape`
  - `TestReadFile_Symlink`: symlink pointing outside root → `ErrSymlink` (test only on non-Windows)
  - `TestWriteFile_Atomic`: interrupted write (simulated) does not leave partial file (verify tmp cleanup)
  - `TestWriteFile_PathEscape`: attempt to write outside root → `ErrPathEscape`
  - `TestWalkDir_SkipsSymlinks`: symlink inside walk root is skipped, not followed

- **Retrofit all file I/O**: modify `internal/config/`, `internal/scaffold/`, `internal/validator/`, `internal/compose/`, `internal/importer/` to replace all `os.ReadFile`, `os.WriteFile`, `filepath.WalkDir` calls with `safepath.*` equivalents

**Security requirements covered by this task:**

| Threat | Mitigation |
|---|---|
| Path traversal via malicious `.ai/` content | `safepath.Join` rejects escaping paths at every call site |
| Symlink following (TOCTOU) | `os.Lstat` before every read; `ErrSymlink` on symlink targets |
| Partial write corruption | Atomic temp-then-rename in `safepath.WriteFile` |
| Directory traversal in agent IDs | `safepath.Join` called with agent ID as a path segment |

**Coding standards:**

- SRP: `safepath` has zero knowledge of `.ai/` semantics; it is a pure filesystem safety layer
- DRY: all file I/O goes through one package — no scattered `os.ReadFile` calls in feature code
- KISS: no capability system, no ACLs — just path-prefix checking and symlink rejection
- Big O: O(p) per call where p = path string length; negligible overhead

**Validation:**

- `go build ./...`: zero errors (all packages compile after retrofit)
- `go test -race -count=1 ./internal/safepath/...`: all tests pass, including symlink test
- `go test -race -count=1 ./...`: all existing tests still pass after retrofit
- `go vet ./...`: zero issues
- `govulncheck ./...`: zero findings
- Security audit: `grep -rn "os\.ReadFile\|os\.WriteFile\|filepath\.WalkDir" internal/ cmd/` → zero hits (all replaced by safepath)

**Prompt context needed:** §8.4 (path security invariants), §8.10 (security considerations table from architecture.md)

---

### Task 14 — Container + Release Hardening ✅

**Goal:** Write the production Dockerfile (distroless, non-root, zero CVEs), the GitHub Actions CI workflow, and verify the full security posture with a `govulncheck` clean run.

**Files to create:**

- `Dockerfile`
  ```dockerfile
  # Stage 1: Build
  FROM golang:1.26-alpine AS builder
  WORKDIR /src
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build -trimpath -ldflags="-s -w" \
      -o /ars ./cmd/ars

  # Stage 2: Final — no shell, no package manager, nonroot user
  FROM gcr.io/distroless/static-debian12:nonroot
  COPY --from=builder /ars /ars
  USER nonroot:nonroot
  ENTRYPOINT ["/ars"]
  ```

- `.github/workflows/ci.yml`
  - Triggers: `push` to `main`, `pull_request`
  - Jobs:
    1. `test`: `go test -race -count=1 ./...`
    2. `lint`: `go vet ./... && staticcheck ./...`
    3. `vuln`: `govulncheck ./...`
    4. `build`: `make build` → upload `bin/ars` as artifact
    5. `docker`: `docker build -t ars:${{ github.sha }} .` → `docker run --rm ars:${{ github.sha }} --version`
  - Go version matrix: `[1.26]` — pin to known good version
  - Cache: `actions/cache` on Go module cache and build cache

- `Makefile` — add:
  - `vuln`: `govulncheck ./...`
  - `docker-scan`: `docker run --rm aquasec/trivy:latest image ars:dev --exit-code 1 --severity HIGH,CRITICAL` (optional, comment if Trivy not available in CI)
  - `release`: build cross-platform binaries (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`)

- `internal/version/version.go`
  - `var Version = "dev"` — overridden by `ldflags -X github.com/ars-standard/ars/internal/version.Version=v1.0.0`
  - `func String() string` — returns `"ars v{Version}"`

**Security requirements verified by this task:**

| Requirement | Verification |
|---|---|
| No shell in final image | `docker run ars:dev sh` → `exec: "sh": executable file not found` |
| No root process | `docker inspect ars:dev --format '{{.Config.User}}'` → `nonroot:nonroot` |
| No OS package CVEs | Distroless has no OS packages; confirmed by `trivy image --list-all-pkgs` |
| No Go dep CVEs | `govulncheck ./...` → zero findings in CI |
| Static binary | `file bin/ars` → `statically linked` |
| Stripped binary | `ls -lh bin/ars` < 8MB (debug info removed) |

**Coding standards:**

- YAGNI: no `EXPOSE` in Dockerfile (binary reads/writes files; no network ports)
- KISS: multi-stage build has exactly two stages — builder and final; no intermediate stages

**Validation:**

- `make build`: produces `bin/ars` binary
- `make vuln`: zero vulnerability findings
- `docker build -t ars:dev .`: succeeds; final image size < 10 MB
- `docker run --rm ars:dev --version`: prints `ars v{version}` and exits 0
- `docker run --rm ars:dev sh`: exits with error (no shell)
- `docker inspect ars:dev --format '{{.Config.User}}'`: returns `nonroot:nonroot`

**Prompt context needed:** §8.5 (Dockerfile spec from architecture.md), §8.6 (govulncheck + security CI), §8.10 (security considerations)

---

### Task 15 — Integration Tests + SPEC.md + README.md ✅

**Goal:** End-to-end round-trip tests covering all four commands and all four compose targets, write the ARS v1 specification, and update the README with quick-start instructions and an architecture overview.

**Files to create / modify:**

- `test/integration/roundtrip_test.go`
  - `TestRoundTrip_CursorComposeThenImport`: `compose --target cursor` then `import cursor` → all agent and instruction content preserved (semantic equality, not byte equality)
  - `TestRoundTrip_CopilotComposeThenImport`: same for `copilot` → `github`
  - `TestRoundTrip_ClaudeComposeThenImport`: same for `claude` → `claude`
  - `TestRoundTrip_AllTargets`: compose all four targets from the same `.ai/` → no file conflicts; all output files trace to a `.ai/` source
  - `TestRoundTrip_EmptyRepo`: `ars init` then `ars validate` → zero errors
  - `TestRoundTrip_AddAgentComposeValidate`: add agent file → `ars compose --target cursor` → new rule file appears; `ars validate` → zero errors
  - All tests use `t.TempDir()` and real `exec.Command("go", "run", "./cmd/ars")` against the binary built in `TestMain`

- `SPEC.md`
  - Section 1: What is ARS — problem statement, scope, non-goals
  - Section 2: `.ai/` file format — `manifest.yaml` schema, `AGENT.md` required sections, `SKILL.md` format, `instructions/` format, `prompts/` format
  - Section 3: `ars` CLI — `init`, `validate`, `compose`, `import` commands, all flags, exit codes
  - Section 4: Provider mappings — full table of how each `.ai/` category maps to each target
  - Section 5: Versioning — how to evolve the standard without breaking existing repos
  - Section 6: Extension guide — how to add a new compose target or import source

- `README.md` — update with:
  - Project description and the "golden rule"
  - Installation: `go install github.com/ars-standard/ars/cmd/ars@latest` and Docker
  - Quick start: `ars init → edit .ai/ → ars compose --target cursor`
  - Command reference table
  - Provider support matrix
  - Contributing note pointing to SPEC.md

**Final Validation Checklist:**

- [ ] `go build ./...` — zero errors
- [ ] `go vet ./...` — zero issues
- [ ] `staticcheck ./...` — zero issues
- [ ] `go test -race -count=1 ./...` — all tests pass including integration
- [ ] `govulncheck ./...` — zero findings
- [ ] `make docker-build` — image builds successfully
- [ ] `docker run --rm ars:dev --version` — exits 0
- [ ] `ars init` in a temp dir → `.ai/manifest.yaml` valid per `ars validate`
- [ ] `ars compose --target cursor` → `.cursor/rules/` present; no files outside repo root
- [ ] `ars compose --target copilot` → `.github/copilot-instructions.md` present
- [ ] `ars compose --target claude` → `CLAUDE.md` present
- [ ] `ars compose --target codex` → `AGENTS.md` present
- [ ] `ars import github` on a real Copilot repo → `.ai/` written without conflicts
- [ ] `ars import cursor` → `.ai/` written
- [ ] `ars import claude` → `.ai/` written
- [ ] Round-trip: `import then compose` preserves all agent and instruction content
- [ ] Security: `grep -rn "os\.ReadFile\|os\.WriteFile\|filepath\.WalkDir" internal/ cmd/` → zero hits
- [ ] SPEC.md complete; describes all four commands and all four targets

**Prompt context needed:** All architecture.md sections; attach full `docs/architecture.md` and all three ADR files

---

### Task 16 — Installation Script + GitHub Release ✅

**Goal:** Ship ARES like Claude Code, `gh`, and `bun` — one `curl | bash` command downloads a pre-built binary, places it in `~/.local/bin/`, and prints shell-specific PATH setup instructions. No Go, no compiler, no SDK required on the end user's machine.

```
curl -fsSL \
  https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh \
  | bash
```

After installation the user sees:

```
✓ Installed ars to /home/user/.local/bin/ars

→ Add ars to your PATH:

  echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc

  Then verify: ars --version
```

**Files to create:**

- `install.sh` — root of repository; see §8.13 for the full script

  Key behaviours:
  - OS/arch detection: `uname -s` → linux | darwin; `uname -m` → amd64 | arm64
  - Download URL: `https://github.com/okfriansyah-moh/ares/releases/latest/download/ars-{os}-{arch}`
  - Checksum: downloads `ars-{os}-{arch}.sha256`; verifies with `sha256sum` or `shasum -a 256`; skips gracefully if neither is present
  - Install dir: `${ARS_INSTALL_DIR:-$HOME/.local/bin}` — overridable via env var; never requires root
  - Atomic install: download to `mktemp`; chmod +x; `mv` to final path — never leaves a partial binary
  - Shell detection: reads `$SHELL`; prints zsh, bash, or fish instructions accordingly
  - PATH check: if `ars` is already on `$PATH`, skip the PATH message
  - Coloured output: uses ANSI codes; gracefully degrades on non-TTY (pipe to file stays readable)
  - Windows: prints a link to the releases page and exits with a clear error (no silent failure)
  - `ARS_VERSION` env var: pin a specific version (`ARS_VERSION=v1.2.0 curl … | bash`)

- `.github/workflows/release.yml` — triggered on `git push --tags v*.*.*`; see §8.13 for full YAML

  Build matrix (all five targets in parallel):
  - `linux/amd64` → `ars-linux-amd64`
  - `linux/arm64` → `ars-linux-arm64`
  - `darwin/amd64` → `ars-darwin-amd64`
  - `darwin/arm64` → `ars-darwin-arm64`
  - `windows/amd64` → `ars-windows-amd64.exe`

  Each target: runs `go test -race -count=1 ./...` + `govulncheck ./...` before build; produces binary + `.sha256` sidecar; uploads both as artifacts; final job creates GitHub Release via `softprops/action-gh-release@v2` with `generate_release_notes: true`

- `Makefile` — add three targets:

  ```makefile
  ## release-dry: build all 5 platform binaries locally; print sizes; no upload
  release-dry:
  	@mkdir -p dist
  	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/ars-linux-amd64   ./cmd/ars
  	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/ars-linux-arm64   ./cmd/ars
  	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/ars-darwin-amd64  ./cmd/ars
  	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/ars-darwin-arm64  ./cmd/ars
  	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/ars-windows-amd64.exe ./cmd/ars
  	@ls -lh dist/

  ## release-checksums: generate .sha256 for every binary in dist/
  release-checksums:
  	@cd dist && for f in ars-*; do sha256sum "$$f" > "$$f.sha256"; done
  	@cat dist/*.sha256

  ## release-tag: create and push a semver tag (usage: make release-tag VERSION=v1.0.0)
  release-tag:
  	@[ -n "$(VERSION)" ] || (echo "Usage: make release-tag VERSION=v1.2.3"; exit 1)
  	git tag -a $(VERSION) -m "Release $(VERSION)"
  	git push origin $(VERSION)
  ```

- `README.md` — add "Installation" section at the top (before Quick Start):

  ```markdown
  ## Installation

  ### One-line installer (macOS + Linux — no Go required)

  curl -fsSL \
    https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh \
    | bash

  Then add to PATH (zsh):

      echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc

  ### Go install (requires Go 1.26+)

      go install github.com/ars-standard/ars/cmd/ars@latest

  ### Docker

      docker run --rm -v "$(pwd):/repo" \
        ghcr.io/ars-standard/ars:latest \
        compose --target cursor --root /repo

  ### Homebrew (coming soon)

      brew install ars-standard/tap/ars
  ```

- `install.sh` — mark executable in git: `git update-index --chmod=+x install.sh`

**Coding standards:**

- KISS: `install.sh` is pure POSIX sh (`#!/usr/bin/env bash`) — no `jq`, no `python`, no `node` required
- Security: `set -euo pipefail`; no `eval`; temp files cleaned with `trap EXIT`; checksum verified before `mv`
- YAGNI: no auto-update mechanism, no package manager integration in v1
- DRY: the download URL pattern is defined once at the top of `install.sh`; never duplicated

**Security requirements for the installer:**

| Risk | Mitigation |
|---|---|
| MITM binary substitution | SHA-256 checksum downloaded from the same release (HTTPS only); verified before install |
| Path traversal via `ARS_INSTALL_DIR` | `mkdir -p` + `mv` — no eval, no shell expansion of untrusted input |
| Partial download corruption | Download to `mktemp`; `mv` only after checksum passes; `trap EXIT` cleans up temp |
| Privilege escalation | Never writes to `/usr/local/bin` without explicit opt-in; defaults to `~/.local/bin` |
| Script injection via `curl | bash` | `set -euo pipefail` — any error aborts immediately; no dynamic code generation |

**Validation:**

- `bash -n install.sh`: no syntax errors
- `shellcheck install.sh`: zero issues (install `shellcheck` via `brew install shellcheck`)
- `bash install.sh` in a fresh temp `$HOME`: creates `~/.local/bin/ars`; prints PATH message
- `ARS_INSTALL_DIR=/tmp/test-install bash install.sh`: installs to `/tmp/test-install/ars`
- `ARS_VERSION=v0.0.1-test bash install.sh`: uses version-pinned URL
- `make release-dry`: builds all 5 platform binaries; each < 10 MB
- `make release-checksums`: produces `.sha256` for all 5 binaries
- `make release-tag VERSION=v1.0.0` (dry-run check): creates annotated tag
- GitHub Actions: on `git push origin v1.0.0` → release workflow creates GitHub Release with 10 assets (5 binaries + 5 checksums)
- `docker run --rm ars:dev --version`: exits 0 (container path still works)

**Prompt context needed:** §8.5 (Dockerfile), §8.6 (CI pipeline), §8.13 (install.sh full source + release.yml full source)

---

## 6. Task Summary

| Task | Name | Key Files | Depends On | Est. Complexity |
|---|---|---|---|---|
| 1 ✅ | Project Scaffold | `go.mod`, `Makefile`, `.gitignore`, `cmd/ars/main.go` stub | — | Low |
| 2 ✅ | Domain Types | `pkg/arslib/types.go`, `interfaces.go`, `types_test.go` | Task 1 | Low |
| 3 ✅ | Config Package | `internal/config/manifest.go`, `types.go`, `manifest_test.go` | Task 2 | Low |
| 4 ✅ | Markdown Utility | `internal/markdown/markdown.go`, `markdown_test.go` | Task 1 | Medium |
| 5 ✅ | Scaffold (`ars init`) | `internal/scaffold/scaffold.go`, `templates/`, `scaffold_test.go` | Tasks 2, 3 | Medium |
| 6 ✅ | Validator (`ars validate`) | `internal/validator/*.go`, `validator_test.go` | Tasks 3, 4 | Medium |
| 7 ✅ | Compose Infra + Cursor | `internal/compose/composer.go`, `cursor.go`, `cursor_test.go` | Tasks 2, 3, 4 | Medium |
| 8 ✅ | Copilot Composer | `internal/compose/copilot.go`, `copilot_test.go` | Task 7 | Medium |
| 9 ✅ | Claude + Codex Composers | `internal/compose/claude.go`, `codex.go`, `shared.go`, tests | Task 7 | Medium |
| 10 ✅ | Importer Infra + GitHub | `internal/importer/importer.go`, `github.go`, `classify.go`, tests | Tasks 2, 3, 4 | High |
| 11 ✅ | Cursor + Claude Importers | `internal/importer/cursor.go`, `claude.go`, tests | Task 10 | Medium |
| 12 ✅ | CLI Wire-up | `cmd/ars/main.go`, `internal/version/version.go`, `main_test.go` | Tasks 5–11 | Medium |
| 13 ✅ | Security Hardening | `internal/safepath/safepath.go`, `safepath_test.go`, retrofit all I/O | Task 12 | High |
| 14 ✅ | Container + Release | `Dockerfile`, `.github/workflows/ci.yml`, Makefile additions | Task 13 | Low |
| 15 ✅ | Integration Tests + Docs | `test/integration/roundtrip_test.go`, `SPEC.md`, `README.md` | Tasks 12–14 | Medium |
| 16 ✅ | Installation Script + GitHub Release | `install.sh`, `.github/workflows/release.yml`, Makefile additions, `README.md` update | Task 15 | Medium |

---

## 7. How to Use This Plan

1. **Start each task in a fresh chat session** — share this `PLAN.md` + `docs/architecture.md` + the relevant §8.X sections listed under "Prompt context needed"
2. **Validate after each task** — run `go build ./...` + `go vet ./...` + `go test -race -count=1 ./...` before moving to the next task
3. **Security check after every task** — run `govulncheck ./...`; zero findings required before proceeding
4. **Retrofit safepath last (Task 13)** — all other tasks use `os.ReadFile` stubs initially; Task 13 retrofits them all at once to avoid blocking Task 5-11 development
5. **Update this plan** as you learn new information during implementation
6. **One task at a time** — do not attempt multiple tasks in a single session to avoid context overflow
7. **Source of truth** — always refer to `docs/architecture.md` for exact design decisions. This `PLAN.md` is the breakdown strategy; the architecture document is the specification.

---

## 8. Deep Knowledge Reference

This section contains complete schemas, algorithms, rules, and coding standards from `docs/architecture.md`. Include the relevant sub-sections in each task session.

---

### 8.1 Core Go Interfaces

```go
// Defined in pkg/arslib/interfaces.go — the canonical contracts.
// No other package defines its own copy of these types.

// Composer translates .ai/ into a single provider-specific artifact.
// Implementations: CursorComposer, CopilotComposer, ClaudeComposer, CodexComposer
type Composer interface {
    Compose(root string, repo *Repository) error
    Target() string // "cursor" | "copilot" | "claude" | "codex"
}

// Importer reads a provider artifact and returns an in-memory Repository.
// Implementations: GitHubImporter, CursorImporter, ClaudeImporter
type Importer interface {
    Import(root string) (*Repository, error)
    Source() string // "github" | "cursor" | "claude"
}

// Validator checks .ai/ structure and returns all findings.
type Validator interface {
    Validate(root string) ([]Finding, error)
}

// FindingLevel — severity of a validation finding.
type FindingLevel int
const (
    OK      FindingLevel = iota
    Warning              // reported, does not fail exit code
    Error                // reported, causes exit code 1
)

// Finding — a single validation result.
type Finding struct {
    Level   FindingLevel `json:"level"`
    Path    string       `json:"path"`
    Message string       `json:"message"`
}
```

---

### 8.2 `.ai/` Schema

**`manifest.yaml` — parsed into `arslib.Manifest`:**

```yaml
version: "2.0"           # ARS spec version; bump on structural changes to .ai/
project:
  name: string           # required; inferred from filepath.Base(root) by ars init
  description: string    # optional one-line description
  repository: string     # optional canonical repo URL
defaults:
  agent: string          # optional default agent ID for compose targets
```

**`agents/<name>/AGENT.md` — required sections (checked by validator):**

| Section | Level if missing | Purpose |
|---|---|---|
| `## Role` | Error | One sentence: what the agent owns |
| `## Responsibilities` | Error | Bullet list of what it does |
| `## Uses` | Error | Skill references (paths to SKILL.md) |
| `## Boundaries` | Error | What it does NOT do |

**`skills/<name>/SKILL.md`** — free-form markdown; no required sections; may have `references/` subdirectory

**`instructions/<name>.md`** — free-form repository-wide rules; no required sections

**`prompts/<name>.md`** — recommended sections:

| Section | Level if missing | Purpose |
|---|---|---|
| `## Use` | Warning | One sentence goal |
| `## Inputs` | — | What to attach |
| `## Instructions` | — | What to do |
| `## Check` | — | Validation criteria |

---

### 8.3 Provider Mapping

| `.ai/` Source | `--target cursor` | `--target copilot` | `--target claude` | `--target codex` |
|---|---|---|---|---|
| `instructions/*.md` | `.cursor/rules/<name>.mdc` (type: always) | `.github/copilot-instructions.md` top section | `CLAUDE.md` top section | `AGENTS.md` top section |
| `agents/<n>/AGENT.md` | `.cursor/rules/<n>.mdc` (type: agent-requested) | Role block in copilot instructions | `## {n}` section in `CLAUDE.md` | Agent entry in `AGENTS.md` |
| `skills/<n>/SKILL.md` | Inlined into referencing agent rule | Inlined under relevant instructions | Inlined under agent context | Inlined under agent context |
| `prompts/<n>.md` | `.cursor/prompts/<n>.prompt` | Not natively supported — skipped | Custom slash command stub | Not natively supported — skipped |
| `manifest.yaml project.name` | Header comment in first rule | Header comment | `CLAUDE.md` H1 title | `AGENTS.md` H1 title |

**Mapping design rules (non-negotiable):**

1. **Lossless intent.** Compose preserves the semantic intent of `.ai/` content; it does not silently truncate.
2. **No orphaned output.** Every generated file traces to at least one `.ai/` source file.
3. **Idempotent.** Running `ars compose` twice produces the same output (byte-identical).
4. **Overwrite safe.** Compose always regenerates the full target; it never partially updates.
5. **Source marker.** Every generated file includes `<!-- ars:source .ai/ -->` (or equivalent) so `ars import` can detect ars-managed files.

---

### 8.4 Path Security Invariants

All file I/O in the codebase must go through `internal/safepath`. Direct `os.ReadFile`, `os.WriteFile`, `filepath.WalkDir` are forbidden outside `safepath` itself.

```go
// Path escape check — the heart of safepath.Join
func Join(root string, parts ...string) (string, error) {
    joined := filepath.Join(append([]string{root}, parts...)...)
    cleaned := filepath.Clean(joined)
    rootClean := filepath.Clean(root) + string(os.PathSeparator)
    if !strings.HasPrefix(cleaned+string(os.PathSeparator), rootClean) {
        return "", fmt.Errorf("safepath: %w: %q escapes root %q", ErrPathEscape, joined, root)
    }
    return cleaned, nil
}
```

**Symlink rule:** `os.Lstat` must be called before `os.ReadFile`. If the result is a symlink, return `ErrSymlink` without following it.

**Atomic write rule:** `safepath.WriteFile` writes to `<target>.tmp` in the same directory, then calls `os.Rename`. This prevents partial writes from leaving corrupt files.

**Verification:** `grep -rn "os\.ReadFile\|os\.WriteFile\|filepath\.WalkDir" internal/ cmd/` must return zero results outside `internal/safepath/safepath.go`.

---

### 8.5 Dockerfile (Multi-stage, Distroless)

```dockerfile
# Stage 1: Build — golang:1.26-alpine; discarded in final image
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=${VERSION:-dev}" \
      -o /ars \
      ./cmd/ars

# Stage 2: Final — distroless/static-debian12:nonroot
# No shell (/bin/sh), no package manager, no libc, runs as UID 65532
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ars /ars
USER nonroot:nonroot
ENTRYPOINT ["/ars"]
```

**Build flags:**
- `-trimpath`: removes local build paths from binary (reproducible builds)
- `-ldflags="-s -w"`: strips symbol table and DWARF debug info; reduces binary size
- `CGO_ENABLED=0`: fully static binary; no glibc dependency
- `GOOS=linux GOARCH=amd64`: deterministic cross-compilation

---

### 8.6 Security CI Pipeline

```yaml
# .github/workflows/ci.yml — jobs in parallel
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.26" }
      - run: go test -race -count=1 -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out   # informational

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.26" }
      - run: go vet ./...
      - run: go install honnef.co/go/tools/cmd/staticcheck@latest
      - run: staticcheck ./...

  vuln:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.26" }
      - run: go install golang.org/x/vuln/cmd/govulncheck@latest
      - run: govulncheck ./...    # exit 1 on any finding

  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker build -t ars:${{ github.sha }} .
      - run: docker run --rm ars:${{ github.sha }} --version
```

---

### 8.7 Validation Rules

**Exit codes:**
- `0` — no Error-level findings (Warnings allowed)
- `1` — at least one Error-level finding

**Complete rule table:**

| Rule | Level | Condition |
|---|---|---|
| `manifest.yaml` missing | Error | File does not exist at `.ai/manifest.yaml` |
| `manifest.yaml` unparseable | Error | YAML parse failure |
| `project.name` empty | Error | `manifest.project.name == ""` |
| `version` unrecognised | Warning | Version string not in known set |
| `defaults.agent` references unknown agent | Warning | Agent ID in defaults not found in `agents/` |
| `agents/*/AGENT.md` missing `## Role` | Error | `markdown.HasSection` returns false |
| `agents/*/AGENT.md` missing `## Responsibilities` | Error | Same |
| `agents/*/AGENT.md` missing `## Uses` | Error | Same |
| `agents/*/AGENT.md` missing `## Boundaries` | Error | Same |
| Skill reference not found | Error | Path listed under `## Uses` resolves to non-existent file |
| Skill reference escapes root | Error | `safepath.Join` returns `ErrPathEscape` |
| `prompts/*.md` missing `## Use` | Warning | Optional but recommended |

---

### 8.8 Compose Algorithm

```
ars compose --target <T>:
  1. Load manifest.yaml → config.Load(root)
  2. Walk .ai/ → build arslib.Repository:
       a. Walk agents/*/AGENT.md → []Agent (sorted by ID)
       b. Walk skills/*/SKILL.md → []Skill (sorted by ID; load references/)
       c. Walk instructions/*.md → []Instruction (sorted by filename)
       d. Walk prompts/*.md → []Prompt (sorted by filename)
       e. Build skill map: id → Skill
       f. For each Agent: resolve SkillRefs from ## Uses section content
  3. Resolve Composer from registry: compose.DefaultRegistry.Get(T)
  4. Call Composer.Compose(root, &repo)
  5. Inside Composer.Compose:
       a. Compute all output paths via safepath.Join
       b. Build output content (pure functions — no side effects until write)
       c. Write atomically via safepath.WriteFile
       d. Remove stale output files not regenerated (prevents orphaned artifacts)
  6. Print summary: "Composed N files to <target path>"
```

**Sorting requirement:** All file walks must produce sorted output. Use `sort.Strings` on the results of `filepath.WalkDir` before processing. This ensures byte-identical output across OS and filesystem implementations.

---

### 8.9 Import Algorithm

```
ars import <S>:
  1. Resolve Importer from registry: importer.DefaultRegistry.Get(S)
  2. Call Importer.Import(root) → *arslib.Repository
  3. Inside Importer.Import:
       a. Locate source file(s) via safepath.Join
       b. Read via safepath.ReadFile
       c. Parse with markdown.ExtractSections or YAML front matter
       d. Classify sections by heuristic (see §8.9.1)
       e. Infer manifest.yaml fields from detected content
       f. Return *Repository (caller handles writing)
  4. Call importer.WriteRepository(root, repo, overwrite flag)
  5. WriteRepository:
       a. For each Agent: write agents/<id>/AGENT.md via safepath.WriteFile
       b. For each Skill: write skills/<id>/SKILL.md via safepath.WriteFile
       c. For each Instruction: write instructions/<id>.md via safepath.WriteFile
       d. Write manifest.yaml via config.Write
       e. Skip existing files when overwrite=false; count conflicts
  6. Print summary: "Created N files, M conflicts skipped"
```

**§8.9.1 Section classification heuristic:**

```go
// In internal/importer/classify.go — shared by all importers
var (
    agentHeadingRe = regexp.MustCompile(`(?i)^agent[:\s]`)
    skillHeadingRe = regexp.MustCompile(`(?i)^skill[:\s]`)
)

func ClassifySection(heading string) classification {
    switch {
    case agentHeadingRe.MatchString(heading): return classAgent
    case skillHeadingRe.MatchString(heading): return classSkill
    default:                                   return classInstruction
    }
}
```

---

### 8.10 Coding Standards (SOLID, DRY, YAGNI, KISS)

**Applied per package:**

| Principle | Application in ARES |
|---|---|
| **S**ingle Responsibility | `safepath` owns I/O safety only; `markdown` owns section extraction only; `config` owns manifest parsing only |
| **O**pen/Closed | Add a new compose target → create one file implementing `arslib.Composer`; `composer.go` is never modified |
| **L**iskov Substitution | All `Composer` and `Importer` implementations are interchangeable via their interface |
| **I**nterface Segregation | Three narrow interfaces (`Composer`, `Importer`, `Validator`); no fat interfaces with unused methods |
| **D**ependency Inversion | `cmd/ars/main.go` depends on `arslib.Composer` interface; never on `*CursorComposer` directly |
| **D**on't Repeat Yourself | Section classification in one place (`classify.go`); shared builder function in `compose/shared.go`; `slugify` in `importer.go` |
| **Y**ou Aren't Gonna Need It | No plugin system, no remote registry, no config file, no `ars run` runtime in v1 |
| **K**eep It Simple, Stupid | Heuristic import uses two regexes; compose uses `strings.Builder`; no AST transformation of output |

---

### 8.11 Complexity Budget

All hot paths must meet these Big O targets:

| Operation | Time | Space | Notes |
|---|---|---|---|
| `config.Load` | O(m) | O(m) | m = manifest file bytes; YAML depth capped at 8 |
| `markdown.ExtractSections` | O(n) | O(k) | n = source bytes; k = section count |
| `validator.Run` | O(f) sort O(r log r) | O(r) | f = total file count; r = finding count (sorted for determinism) |
| `compose.Compose` | O(a × s + i) | O(output size) | a = agents; s = avg skills per agent; i = instruction bytes |
| `importer.Import` | O(n) | O(section count) | n = source file bytes |
| `importer.WriteRepository` | O(k) | O(1) per file | k = output file count; streaming writes |
| `safepath.Join` | O(p) | O(p) | p = path string length; called per file operation |
| File walk (compose/validate) | O(f) | O(f) | f = files in `.ai/`; results sorted for determinism |

**Sorting for determinism:** Any file walk result must be sorted before processing. `sort.Strings` is O(f log f) where f = file count; this is acceptable because f < 1000 in any realistic `.ai/` directory.

---

### 8.13 Installation Script + Release Workflow

#### `install.sh` (full source)

```bash
#!/usr/bin/env bash
# install.sh — ARES one-line installer
# Usage: curl -fsSL https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh | bash
# Override install dir:   ARS_INSTALL_DIR=/usr/local/bin bash install.sh
# Pin a version:          ARS_VERSION=v1.2.0 bash install.sh

set -euo pipefail

REPO="okfriansyah-moh/ares"
INSTALL_DIR="${ARS_INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="ars"
VERSION="${ARS_VERSION:-latest}"

# ── Colour helpers (degrade gracefully on non-TTY) ───────────────────────
if [ -t 1 ]; then
  GREEN='\033[32m'; BLUE='\033[34m'; YELLOW='\033[33m'; RED='\033[31m'; RESET='\033[0m'; BOLD='\033[1m'
else
  GREEN=''; BLUE=''; YELLOW=''; RED=''; RESET=''; BOLD=''
fi

info()  { printf "${BLUE}→${RESET} %s\n"  "$*"; }
ok()    { printf "${GREEN}✓${RESET} %s\n" "$*"; }
warn()  { printf "${YELLOW}!${RESET} %s\n" "$*" >&2; }
fatal() { printf "${RED}Error:${RESET} %s\n" "$*" >&2; exit 1; }

# ── OS + arch detection ──────────────────────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "${ARCH}" in
  x86_64)            ARCH="amd64" ;;
  aarch64 | arm64)   ARCH="arm64" ;;
  *) fatal "unsupported architecture: ${ARCH}" ;;
esac

case "${OS}" in
  linux | darwin) ;;
  msys* | cygwin* | mingw*)
    fatal "Windows detected. Download the binary from: https://github.com/${REPO}/releases"
    ;;
  *) fatal "unsupported OS: ${OS}. Download from: https://github.com/${REPO}/releases" ;;
esac

ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"

# ── Resolve download URLs ────────────────────────────────────────────────
if [ "${VERSION}" = "latest" ]; then
  BASE_URL="https://github.com/${REPO}/releases/latest/download"
else
  BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
fi

DOWNLOAD_URL="${BASE_URL}/${ASSET_NAME}"
CHECKSUM_URL="${BASE_URL}/${ASSET_NAME}.sha256"

# ── Download to temp files ───────────────────────────────────────────────
info "Downloading ars (${OS}/${ARCH})..."
mkdir -p "${INSTALL_DIR}"

TMP_BIN="$(mktemp)"
TMP_SUM="$(mktemp)"
trap 'rm -f "${TMP_BIN}" "${TMP_SUM}"' EXIT

curl -fsSL --progress-bar "${DOWNLOAD_URL}" -o "${TMP_BIN}" \
  || fatal "download failed: ${DOWNLOAD_URL}"

# ── Checksum verification (optional until v1 release ships .sha256) ──────
curl -fsSL "${CHECKSUM_URL}" -o "${TMP_SUM}" 2>/dev/null || true

if [ -s "${TMP_SUM}" ]; then
  info "Verifying checksum..."
  EXPECTED="$(awk '{print $1}' "${TMP_SUM}")"
  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL="$(sha256sum "${TMP_BIN}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    ACTUAL="$(shasum -a 256 "${TMP_BIN}" | awk '{print $1}')"
  else
    warn "no sha256 tool found; skipping checksum verification"
    ACTUAL="${EXPECTED}"
  fi
  [ "${ACTUAL}" = "${EXPECTED}" ] \
    || fatal "checksum mismatch\n  expected: ${EXPECTED}\n  got:      ${ACTUAL}"
  ok "Checksum verified"
fi

# ── Install ───────────────────────────────────────────────────────────────
chmod +x "${TMP_BIN}"
mv "${TMP_BIN}" "${INSTALL_DIR}/${BINARY_NAME}"
ok "Installed ars to ${INSTALL_DIR}/${BINARY_NAME}"

# ── PATH setup instructions ───────────────────────────────────────────────
if command -v ars >/dev/null 2>&1; then
  printf '\n'
  ok "ars is already on your PATH."
  printf '\n  Verify: %sars --version%s\n' "${BOLD}" "${RESET}"
else
  printf '\n'
  case "${SHELL:-bash}" in
    */zsh)
      PATH_CMD="echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
      ;;
    */fish)
      PATH_CMD="fish_add_path \$HOME/.local/bin"
      ;;
    *)
      PATH_CMD="echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
      ;;
  esac
  printf "${YELLOW}→${RESET} Add ars to your PATH:\n\n"
  printf "  %s\n\n" "${PATH_CMD}"
  printf "  Then verify: %sars --version%s\n" "${BOLD}" "${RESET}"
fi
```

#### `.github/workflows/release.yml` (full source)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write

jobs:
  build:
    name: Build ${{ matrix.goos }}/${{ matrix.goarch }}
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        include:
          - goos: linux
            goarch: amd64
            asset: ars-linux-amd64
          - goos: linux
            goarch: arm64
            asset: ars-linux-arm64
          - goos: darwin
            goarch: amd64
            asset: ars-darwin-amd64
          - goos: darwin
            goarch: arm64
            asset: ars-darwin-arm64
          - goos: windows
            goarch: amd64
            asset: ars-windows-amd64.exe

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
          cache: true

      - name: Run tests
        run: go test -race -count=1 ./...

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Build binary
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: "0"
        run: |
          go build \
            -trimpath \
            -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=${{ github.ref_name }}" \
            -o "${{ matrix.asset }}" \
            ./cmd/ars

      - name: Generate checksum
        run: sha256sum "${{ matrix.asset }}" > "${{ matrix.asset }}.sha256"

      - name: Verify binary runs
        if: matrix.goos == 'linux'
        run: ./${{ matrix.asset }} --version

      - uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.asset }}
          path: |
            ${{ matrix.asset }}
            ${{ matrix.asset }}.sha256

  release:
    name: Create GitHub Release
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/download-artifact@v4
        with:
          path: dist/
          merge-multiple: true

      - name: List release assets
        run: ls -lh dist/

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
          generate_release_notes: true
          fail_on_unmatched_files: true
```

#### Release trigger workflow

```bash
# Tag and push to trigger the release workflow
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Or use the Makefile target
make release-tag VERSION=v1.0.0
```

#### Asset naming convention

| Platform | Asset filename | Checksum |
|---|---|---|
| Linux x86-64 | `ars-linux-amd64` | `ars-linux-amd64.sha256` |
| Linux ARM64 | `ars-linux-arm64` | `ars-linux-arm64.sha256` |
| macOS Intel | `ars-darwin-amd64` | `ars-darwin-amd64.sha256` |
| macOS Apple Silicon | `ars-darwin-arm64` | `ars-darwin-arm64.sha256` |
| Windows x86-64 | `ars-windows-amd64.exe` | `ars-windows-amd64.exe.sha256` |

#### User experience by method

| Method | Requires | Command |
|---|---|---|
| One-line install (recommended) | `curl`, `bash` | `curl -fsSL https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh \| bash` |
| Go install | Go 1.26+ | `go install github.com/ars-standard/ars/cmd/ars@latest` |
| Docker | Docker | `docker run --rm -v "$(pwd):/repo" ghcr.io/ars-standard/ars:latest compose --target cursor --root /repo` |
| Manual download | `curl`/browser | `https://github.com/okfriansyah-moh/ares/releases/latest` |

---

### 8.12 Definition of Done per Task

A task session is complete when:

- [ ] All listed files are created with non-stub production implementation
- [ ] `go build ./...` passes with zero errors
- [ ] `go vet ./...` reports zero issues
- [ ] `staticcheck ./...` reports zero issues
- [ ] `go test -race -count=1 ./...` passes with no failures or data races
- [ ] `govulncheck ./...` reports zero vulnerability findings
- [ ] All new file I/O goes through `safepath.*` (or uses the stub until Task 13)
- [ ] No `os.Exit` called directly in non-`main` packages
- [ ] No global mutable state introduced
- [ ] No `fmt.Println` in non-`cmd/` packages (use `slog` or return errors)
- [ ] Complexity budget met (see §8.11)
