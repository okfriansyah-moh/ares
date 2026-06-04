![ARES AI Repository Standard](assets/ares-infographic.png)

# ARES

**AI Repository Standard** — define repository knowledge once in `.ai/`, generate provider-specific conventions for Cursor, GitHub Copilot, Claude, Codex, and future tools.

---

## Installation

### One-line installer — macOS and Linux (no Go required)

```bash
curl -fsSL \
  https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh \
  | bash
```

After installation, add `ars` to your PATH:

**zsh**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

**bash**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

**fish**
```bash
fish_add_path $HOME/.local/bin
```

Verify the installation:

```bash
ars --version
```

---

### Go install (requires Go 1.26+)

```bash
go install github.com/ars-standard/ars/cmd/ars@latest
```

---

### Docker

```bash
docker run --rm \
  -v "$(pwd):/repo" \
  ghcr.io/ars-standard/ars:latest \
  compose --target cursor --root /repo
```

---

### Manual download

Download a pre-built binary from the [Releases page](https://github.com/okfriansyah-moh/ares/releases/latest):

| Platform | Binary |
|---|---|
| macOS (Apple Silicon) | `ars-darwin-arm64` |
| macOS (Intel) | `ars-darwin-amd64` |
| Linux x86-64 | `ars-linux-amd64` |
| Linux ARM64 | `ars-linux-arm64` |
| Windows x86-64 | `ars-windows-amd64.exe` |

Place the binary in any directory on your `PATH` and mark it executable:

```bash
chmod +x ars-darwin-arm64
mv ars-darwin-arm64 ~/.local/bin/ars
```

---

### Pin a specific version

```bash
ARS_VERSION=v1.0.0 curl -fsSL \
  https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh \
  | bash
```

---

## What ARES Is

- A repository standard for AI-assisted development.
- A canonical source of truth in `.ai/`.
- A portability layer for provider conventions.

## What ARES Is Not

- Not an agent runtime.
- Not an orchestration engine.
- Not a memory system.
- Not a provider API integration.

---

## Quick Start

```bash
# 1. Initialise .ai/ in your repository
ars init

# 2. Edit your knowledge files
#    .ai/manifest.yaml        — project identity
#    .ai/instructions/        — repository-wide rules
#    .ai/agents/              — role definitions
#    .ai/skills/              — reusable knowledge
#    .ai/prompts/             — task templates

# 3. Validate structure
ars validate

# 4. Generate provider artifacts
ars compose --target cursor    # → .cursor/
ars compose --target copilot   # → .github/copilot-instructions.md
ars compose --target claude    # → CLAUDE.md
ars compose --target codex     # → AGENTS.md

# 5. Or import from an existing provider setup
ars import github              # from .github/copilot-instructions.md
ars import cursor              # from .cursor/rules/
ars import claude              # from CLAUDE.md
```

The golden rule: `delete all generated files → ars compose → everything comes back`.

---

## Core Structure

```
.ai/
  manifest.yaml        ← project identity and defaults
  instructions/        ← repository-wide rules
  agents/              ← role definitions (AGENT.md per role)
  skills/              ← reusable knowledge (SKILL.md per skill)
  prompts/             ← reusable task templates
```

Everything else is generated from this source.

---

## Commands

| Command | Description |
|---|---|
| `ars init` | Scaffold a valid `.ai/` skeleton in the current repository |
| `ars validate` | Check `.ai/` structure, cross-references, and required sections |
| `ars compose --target <T>` | Generate provider artifacts from `.ai/` |
| `ars import <S>` | Convert a provider artifact into `.ai/` |

### Compose targets

| `--target` | Output |
|---|---|
| `cursor` | `.cursor/rules/` and `.cursor/prompts/` |
| `copilot` | `.github/copilot-instructions.md` |
| `claude` | `CLAUDE.md` |
| `codex` | `AGENTS.md` |

### Import sources

| `<source>` | Reads from |
|---|---|
| `github` | `.github/copilot-instructions.md` |
| `cursor` | `.cursor/rules/*.mdc` |
| `claude` | `CLAUDE.md` |

---

## Example Workflow

**Migrate from GitHub Copilot to Cursor:**

```bash
ars import github              # converts .github/ → .ai/
ars compose --target cursor    # generates .cursor/ from .ai/
```

**Multi-tool team (Copilot + Cursor + Claude):**

```bash
# Each developer runs compose for their own tool
ars compose --target cursor    # Developer A
ars compose --target copilot   # Developer B
ars compose --target claude    # Developer C
```

Same `.ai/` source. Different tooling. Shared knowledge.

---

## Why ARES Works

Today, repository knowledge is fragmented:

```
.github/       ← GitHub Copilot
.cursor/       ← Cursor
CLAUDE.md      ← Claude Code
AGENTS.md      ← OpenAI Codex
```

When you change tools or add a new team member using a different AI assistant, you rewrite everything. ARES solves this by treating `.ai/` as the source and all provider files as derived artifacts.

---

## Success Criteria

Repository knowledge can move into `.ai/`, remain understandable there, and be composed into provider conventions for any tool — without losing intent.

---

## Documentation

- [Architecture](docs/architecture.md) — system design, component responsibilities, security model
- [Implementation Plan](docs/PLAN.md) — task breakdown, dependency graph, coding standards
- [Specification](SPEC.md) — `.ai/` file format, CLI reference, provider mappings
- [ADR-0001](docs/decisions/ADR-0001-go-cli.md) — why Go
- [ADR-0002](docs/decisions/ADR-0002-markdown-as-source-format.md) — why Markdown
- [ADR-0003](docs/decisions/ADR-0003-distroless-container.md) — why Distroless

---

## License

This repository is covered by the `LICENSE` file in the root.
