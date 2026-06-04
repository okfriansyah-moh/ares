# ARES Architecture

> **Version:** 1.0
> **Date:** 2026-06-04
> **Status:** Accepted
> **Author:** Core Team
> **Source of Truth:** `ares-brainstorming-result.md`, `ares-mission.md`, `cli-ares-plan.md`

---

## 1. Problem Statement

AI-assisted development tools each define their own repository convention format:

| Tool | Convention |
|---|---|
| GitHub Copilot | `.github/copilot-instructions.md` |
| Cursor | `.cursor/rules/` |
| Claude Code | `CLAUDE.md` |
| OpenAI Codex | `AGENTS.md` |
| Roo / Cline / Kimi | vendor-specific directories |

When a team changes tools or uses multiple tools simultaneously, repository knowledge must be rewritten, duplicated, or manually kept in sync. Every new tool that enters the market adds a new migration burden. There is no shared layer.

---

## 2. Solution: ARES

ARES (AI Repository Standard) is the reference implementation of ARS. It introduces `.ai/` as the canonical repository knowledge layer. All provider-specific files are generated from it.

```
Human edits:

  .ai/
    manifest.yaml
    instructions/
    agents/
    skills/
    prompts/

        │
        ▼
    ars compose

        │
        ├──▶ .cursor/          (Cursor)
        ├──▶ .github/          (GitHub Copilot)
        ├──▶ CLAUDE.md         (Claude Code)
        └──▶ AGENTS.md         (OpenAI Codex)
```

**Golden rule:** delete all generated files → `ars compose` → everything comes back.

The reverse direction is also supported:

```
Provider artifact:

  .github/   |   .cursor/   |   CLAUDE.md

        │
        ▼
    ars import

        │
        ▼
  .ai/   ← canonical source going forward
```

ARES is closer to OpenAPI or Terraform than to an agent runtime. The standard is the product. Tooling exists to support the standard, not replace it.

---

## 3. Scope

### In Scope — v1

| Area | Detail |
|---|---|
| `.ai/` standard | `manifest.yaml`, `instructions/`, `agents/`, `skills/`, `prompts/` |
| CLI binary | `ars init`, `ars validate`, `ars compose`, `ars import` |
| Compose targets | `cursor`, `copilot`, `claude`, `codex` |
| Import sources | `github` (Copilot), `cursor`, `claude` |
| Container image | Distroless, non-root, statically compiled |

### Out of Scope — v1

| Area | Why |
|---|---|
| `ars run` / execution runtime | Out of scope until the standard is adopted |
| Provider API integration | ARES is not a routing or inference layer |
| Agent orchestration | Not a workflow engine |
| Global skill registry | Not a marketplace |
| Memory or state system | Not an agent memory layer |
| Web or TUI frontend | CLI is sufficient for a repository tool |
| Database | All operations are local and file-based |

---

## 4. System Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                           Repository Root                            │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                         .ai/                                │    │
│  │   manifest.yaml                                             │    │
│  │   instructions/       agents/       skills/      prompts/   │    │
│  └─────────────────────────────┬───────────────────────────────┘    │
│                                │                                     │
│          ┌──────────┬──────────┼──────────┬──────────┐              │
│          │          │          │          │          │              │
│      ars init   ars validate  ars compose          ars import       │
│          │          │          │                    │               │
│          ▼          ▼          ▼                    ▼               │
│       scaffold   report   provider artifacts    .ai/ skeleton       │
│       .ai/                 .cursor/             from provider       │
│       skeleton             .github/             convention          │
│                            CLAUDE.md                                │
│                            AGENTS.md                                │
└──────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Single Responsibility |
|---|---|
| `ars init` | Scaffold a valid `.ai/` skeleton in the repository root |
| `ars validate` | Check `.ai/` structure, cross-references, and required sections |
| `ars compose` | Translate `.ai/` into a provider-specific artifact |
| `ars import` | Read a provider artifact and produce an equivalent `.ai/` |
| `config` | Parse and validate `manifest.yaml` |
| `scaffold` | Write the initial `.ai/` directory structure |
| `validator` | Report structural errors, missing files, and broken skill references |
| `compose/*` | One composer per provider target |
| `importer/*` | One importer per provider source |

---

## 5. CLI Architecture

The ARES CLI is a single statically compiled Go binary. All operations are local and file-based. No network calls in v1.

### Directory Structure

```
ares/
├── cmd/
│   └── ars/
│       └── main.go                  ← cobra root command + subcommand registration
├── internal/
│   ├── config/
│   │   ├── manifest.go              ← parse, validate, and marshal manifest.yaml
│   │   └── types.go                 ← Manifest, Project, Defaults structs
│   ├── scaffold/
│   │   └── scaffold.go              ← ars init: create .ai/ skeleton from embedded template
│   ├── validator/
│   │   ├── validator.go             ← Validator interface + runner
│   │   ├── manifest.go              ← manifest.yaml validation rules
│   │   ├── agents.go                ← AGENT.md required-section checks
│   │   ├── skills.go                ← skill reference resolution
│   │   └── prompts.go               ← prompt file existence checks
│   ├── compose/
│   │   ├── composer.go              ← Composer interface
│   │   ├── cursor.go                ← .cursor/ output
│   │   ├── copilot.go               ← .github/copilot-instructions.md output
│   │   ├── claude.go                ← CLAUDE.md output
│   │   └── codex.go                 ← AGENTS.md output
│   └── importer/
│       ├── importer.go              ← Importer interface
│       ├── github.go                ← import from .github/copilot-instructions.md
│       ├── cursor.go                ← import from .cursor/rules/
│       └── claude.go                ← import from CLAUDE.md
├── pkg/
│   └── arslib/
│       └── types.go                 ← public types: Repository, Agent, Skill, Prompt, Instruction
├── docs/
│   ├── architecture.md              ← this file
│   └── decisions/                   ← ADRs
├── .ai/                             ← ARES's own .ai/ (reference implementation)
├── examples/
│   └── a2a-brainstormer/            ← example .ai/ for a real project
├── SPEC.md                          ← ARS v1 specification
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

### Core Interfaces

```go
// Composer translates .ai/ into a single provider-specific artifact.
type Composer interface {
    Compose(root string, repo *arslib.Repository) error
    Target() string
}

// Importer reads a provider artifact and returns a parsed Repository.
type Importer interface {
    Import(root string) (*arslib.Repository, error)
    Source() string
}

// Validator checks the .ai/ structure and returns all findings.
type Validator interface {
    Validate(root string) ([]Finding, error)
}

// Finding is a single validation result.
type Finding struct {
    Level   Level  // OK, Warning, Error
    Path    string
    Message string
}
```

### Key Domain Types

```go
// Repository is the in-memory representation of a .ai/ directory.
type Repository struct {
    Manifest     Manifest
    Instructions []Instruction
    Agents       []Agent
    Skills       []Skill
    Prompts      []Prompt
}

type Agent struct {
    ID           string
    Path         string
    Content      string   // raw AGENT.md markdown
    SkillRefs    []string // skill IDs referenced under ## Uses
}

type Skill struct {
    ID         string
    Path       string
    Content    string   // raw SKILL.md markdown
    References []string // files under references/
}
```

---

## 6. Data Model — `.ai/` Schema

### manifest.yaml

```yaml
version: "2.0"           # ARS spec version — bump on structural changes to .ai/
project:
  name: string           # repository name (required)
  description: string    # one-line description (optional)
  repository: string     # canonical repo URL (optional)
defaults:
  agent: string          # default agent ID used by compose targets
```

### agents/\<name\>/AGENT.md

Markdown. Required sections:

| Section | Purpose |
|---|---|
| `## Role` | One sentence: what the agent owns |
| `## Responsibilities` | Bullet list of what it does |
| `## Uses` | Skill references (e.g., `.ai/skills/<name>/SKILL.md`) |
| `## Boundaries` | What it does NOT do |

### skills/\<name\>/SKILL.md

Markdown. The authoritative knowledge source for the skill. Free-form sections. May reference supplementary materials under `references/`.

### instructions/\<name\>.md

Markdown. Repository-wide rules. Free-form. No required sections.

### prompts/\<name\>.md

Markdown. Reusable prompt template. Recommended sections:

| Section | Purpose |
|---|---|
| `## Use` | One sentence goal |
| `## Inputs` | What to attach |
| `## Instructions` | What to do |
| `## Check` | Validation criteria |

---

## 7. Data Flows

### 7.1 Compose Flow (`ars compose --target cursor`)

```
1. Read and parse manifest.yaml
2. Walk agents/, skills/, instructions/, prompts/
3. Build in-memory Repository struct
4. Validate .ai/ structure (abort on Error-level findings)
5. Select CursorComposer
6. Map Repository to .cursor/ structure:
       instructions/*.md       → .cursor/rules/<name>.mdc (type: always)
       agents/<n>/AGENT.md     → .cursor/rules/<n>.mdc (type: agent-requested)
       skills/<n>/SKILL.md     → inlined into referencing agent rule
       prompts/<n>.md          → .cursor/prompts/<n>.prompt (if supported)
7. Write .cursor/ to repository root
8. Print summary: N files written
```

### 7.2 Import Flow (`ars import github`)

```
1. Detect .github/copilot-instructions.md
2. Parse the file:
       Role sections       → agents/<name>/AGENT.md
       Skill blocks        → skills/<name>/SKILL.md
       Global instructions → instructions/<name>.md
3. Infer manifest.yaml fields from detected content
4. Write .ai/ structure (skip existing files, report conflicts)
5. Print summary: N files created, M conflicts
```

### 7.3 Validate Flow (`ars validate`)

```
1. Check manifest.yaml exists and is parseable
2. For each agent: AGENT.md exists + has required sections
3. For each skill reference in agent ## Uses: SKILL.md exists
4. For each prompt: file exists
5. Check version field in manifest.yaml is recognized
6. Report per-finding: level, path, message
7. Exit code 0 if no Errors, 1 if any Error
```

### 7.4 Init Flow (`ars init`)

```
1. Check .ai/ does not already exist (abort if it does, with --force flag to override)
2. Write scaffolded .ai/ from embedded template:
       .ai/manifest.yaml
       .ai/instructions/README.md (placeholder)
       .ai/agents/.gitkeep
       .ai/skills/.gitkeep
       .ai/prompts/.gitkeep
3. Print next steps
```

---

## 8. Provider Mapping

The table below defines how each `.ai/` category maps to each provider's convention.

| `.ai/` Source | `--target cursor` | `--target copilot` | `--target claude` | `--target codex` |
|---|---|---|---|---|
| `instructions/*.md` | `.cursor/rules/*.mdc` (always) | `.github/copilot-instructions.md` top section | `CLAUDE.md` top section | `AGENTS.md` top section |
| `agents/<n>/AGENT.md` | `.cursor/rules/<n>.mdc` (agent-requested) | Role block in copilot instructions | Agent section in `CLAUDE.md` | Agent entry in `AGENTS.md` |
| `skills/<n>/SKILL.md` | Inlined into referencing agent rule | Inlined under relevant instructions | Inlined under agent context | Inlined under agent context |
| `prompts/<n>.md` | `.cursor/prompts/<n>.prompt` | Not natively supported — skipped | Custom slash command stub | Not natively supported — skipped |
| `manifest.yaml` `project.name` | Header comment in rules | Header comment | `CLAUDE.md` title | `AGENTS.md` title |

### Mapping Design Rules

1. **Lossless intent.** Compose must preserve the semantic intent of `.ai/` content, not just copy text.
2. **No orphaned output.** Every generated file traces to at least one `.ai/` source file.
3. **Idempotent.** Running `ars compose` twice produces the same output.
4. **Overwrite safe.** Compose always regenerates the full target; it never partially updates.

---

## 9. Tech Stack

| Layer | Technology | Version | Rationale |
|---|---|---|---|
| CLI language | Go | 1.26 (latest stable) | Single static binary; no runtime deps; cross-platform; strong stdlib for file I/O and YAML |
| CLI framework | `github.com/spf13/cobra` | latest | De-facto standard for Go CLIs; subcommands, help generation, shell completion |
| YAML parsing | `gopkg.in/yaml.v3` | latest | Standard Go YAML library; strict struct mapping |
| Markdown parsing | `github.com/yuin/goldmark` | latest | AST-level parse for section extraction during import |
| Embedded templates | `embed` (stdlib) | Go 1.16+ | Embed `.ai/` scaffold templates in the binary; no external file deps |
| Testing | `testing` (stdlib) + `github.com/stretchr/testify` | latest | Table-driven unit tests; assertion helpers |
| Build | `make` | system | Simple, portable build automation |
| Container base | `gcr.io/distroless/static-debian12:nonroot` | latest | Zero OS packages; no shell; minimal attack surface |
| Container builder | `golang:1.26-alpine` | latest | Multi-stage build — compiles binary, discarded in final image |
| Vulnerability scan | `govulncheck` | latest | Go-native vuln scanner against the Go vuln DB |
| CI | GitHub Actions | — | Standard for open source Go projects |

### Not Used (and Why)

| Technology | Reason |
|---|---|
| PostgreSQL / any database | No persistent state; all operations are file-based |
| SvelteKit / any frontend | CLI is the interface; no GUI in v1 |
| Docker Compose | Single binary; no multi-service topology |
| gRPC / HTTP server | No network communication in v1 |
| CGO | Static compilation requires `CGO_ENABLED=0` |
| Any ORM | No database |

---

## 10. Container Strategy

### Dockerfile

```dockerfile
# ── Stage 1: Build ─────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Download deps first — cached layer unless go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary: no CGO, stripped symbols, no debug info, no local paths
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /ars \
      ./cmd/ars

# ── Stage 2: Final ─────────────────────────────────────────────────────────
# distroless/static: no shell, no package manager, no libc, no /tmp writeable
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /ars /ars

USER nonroot:nonroot

ENTRYPOINT ["/ars"]
```

### Security Requirements

| Requirement | Implementation |
|---|---|
| No shell in final image | `distroless/static-debian12` — no `/bin/sh`, no `/bin/bash` |
| No root process | `USER nonroot:nonroot` — UID 65532 |
| No OS package CVEs | Distroless has no OS packages; nothing to CVE-scan |
| No Go dependency CVEs | `govulncheck ./...` runs in CI on every PR and release |
| Minimal binary size | `-ldflags="-s -w"` strips symbol table and debug sections |
| Reproducible build | `-trimpath` removes local build paths from the binary |
| No dynamic linking | `CGO_ENABLED=0` — fully static; no glibc dependency |
| Read-only filesystem | Binary has no write path inside the container; mounts are the write surface |
| No capability escalation | Distroless base and `nonroot` user prevent privilege escalation |

### Image Tagging Convention

```
ghcr.io/ars-standard/ars:v1.0.0        ← immutable release tag
ghcr.io/ars-standard/ars:latest        ← mutable latest (CI use)
```

Releases are signed with `cosign` and published with a Software Bill of Materials (SBOM).

---

## 11. Key Architectural Decisions

| # | Decision | Choice | Alternatives | Tradeoff |
|---|---|---|---|---|
| 1 | Primary file format for `.ai/` | Markdown | YAML, JSON, TOML | Markdown is human-readable, editable in any editor, and rendered by every provider. Sacrifices machine-strictness for portability and legibility. |
| 2 | Single config file | `manifest.yaml` only | Per-file metadata YAML | One config entry reduces cognitive load and file count. Per-file metadata was rejected; markdown front matter would couple `.ai/` to specific parsers. |
| 3 | CLI language | Go | Node.js, Python, Rust | Go produces a single static binary with no runtime install requirement. `go install` works out of the box. Cross-platform without a build matrix. |
| 4 | CLI framework | Cobra | urfave/cli, manual flags | Cobra is the de-facto standard for Go CLIs. Subcommands, help generation, and shell completion are built in. |
| 5 | No database | File-based only | SQLite for caching/indexing | ARES operates on the repository filesystem. Introducing a database creates a state-synchronization problem without demonstrated need. |
| 6 | No frontend | CLI only | TUI (bubbletea), Web UI | A CLI is the natural interface for a repository tool. A TUI or web UI adds distribution complexity before the standard has any adoption. |
| 7 | Container base | `distroless/static-debian12:nonroot` | Alpine, scratch | Distroless has no shell and no package manager, eliminating the largest class of container CVEs. Smaller than Alpine for static binaries. `scratch` lacks CA certificates needed for future HTTPS. |
| 8 | Provider mapping in code | Typed Go structs + interfaces | Template files, plugin system | Compiled composers and importers are type-safe and testable. A plugin architecture adds significant complexity before a single provider mapping is proven stable. |
| 9 | Scope lock | Repository knowledge layer only | Include `ars run` runtime | The scope lock prevents ARES from becoming another orchestration framework. The value proposition is the standard, not the execution engine. |
| 10 | Agent files named `AGENT.md` | Single `AGENT.md` per agent | `metadata.yaml` + `instructions.md` | One file per agent keeps the structure minimal and reduces file count. The YAML/markdown split adds value only when tooling must parse metadata separately — no such tooling exists yet. |

For full ADR detail, see `docs/decisions/`.

---

## 12. Extension Points

### Adding a New Compose Target

1. Create `internal/compose/<provider>.go`
2. Implement the `Composer` interface
3. Register in `internal/compose/registry.go`
4. Add `--target <provider>` to the compose command
5. Document the mapping in `SPEC.md`

### Adding a New Import Source

1. Create `internal/importer/<source>.go`
2. Implement the `Importer` interface
3. Register in `internal/importer/registry.go`
4. Add `ars import <source>` subcommand

### Adding a New `.ai/` Category (Future)

1. Update `manifest.yaml` schema — add the new category
2. Bump `version` field in `manifest.yaml`
3. Update `validator` to check the new structure
4. Update all composers to handle the new category
5. Document in `SPEC.md`

Extension must not break existing `ars compose` or `ars import` calls.

---

## 13. Security Considerations

| Surface | Risk | Mitigation |
|---|---|---|
| File writes during compose | Path traversal via malicious `.ai/` content | Resolve all paths against repository root; reject any path that escapes root |
| File reads during import | Symlink following into sensitive paths | Use `os.Lstat` before `os.Open`; reject symlinks pointing outside root |
| YAML parsing | Billion laughs / anchor expansion DoS | Use `gopkg.in/yaml.v3` with depth limit; avoid `interface{}` unmarshalling |
| Binary distribution | Supply chain tampering | Sign releases with `cosign`; publish SBOM; pin builder image by digest in Dockerfile |
| Container runtime | Privilege escalation | Nonroot user; read-only root filesystem via `--read-only` in `docker run`; no `CAP_*` grants |
| Dependency CVEs | Known vulnerabilities in Go modules | `govulncheck ./...` in CI; Dependabot for automated dep updates |

---

## 14. Definition of Done — v1

- [ ] `ars init` creates a valid, runnable `.ai/` skeleton
- [ ] `ars validate` reports all structural errors with file path and message
- [ ] `ars compose --target cursor` produces `.cursor/` from `.ai/`
- [ ] `ars compose --target copilot` produces `.github/copilot-instructions.md` from `.ai/`
- [ ] `ars compose --target claude` produces `CLAUDE.md` from `.ai/`
- [ ] `ars compose --target codex` produces `AGENTS.md` from `.ai/`
- [ ] `ars import github` converts `.github/` to `.ai/` without losing intent
- [ ] `ars import cursor` converts `.cursor/` to `.ai/` without losing intent
- [ ] `ars import claude` converts `CLAUDE.md` to `.ai/` without losing intent
- [ ] Round-trip: `ars import <source>` → `ars compose --target <same>` produces semantically equivalent output
- [ ] `go test ./...` passes; compose and import packages have ≥ 80% coverage
- [ ] `govulncheck ./...` reports zero findings
- [ ] Container image is based on distroless and runs as UID 65532 (nonroot)
- [ ] `SPEC.md` documents the full `.ai/` format, all compose targets, and all import sources
