# ARES

ARES is the reference implementation of ARS, the AI Repository Standard.

![ARES Infographic](assets/ares-infographic.png)

It lets a repository define durable AI coding knowledge once in `.ai/`, then generate provider-specific files for Cursor, GitHub Copilot, Claude Code, and OpenAI Codex.

The golden rule: delete generated provider files, run `ars compose`, and everything comes back from `.ai/`.

## Installation

### One-line installer

macOS and Linux, no Go required:

```sh
curl -fsSL https://raw.githubusercontent.com/okfriansyah-moh/ares/main/install.sh | bash
```

Then add to PATH if prompted:

```sh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

### Go Install

```sh
go install github.com/okfriansyah-moh/ares/cmd/ars@latest
```

### Docker

```sh
docker run --rm -v "$(pwd):/repo" ghcr.io/okfriansyah-moh/ares:latest compose --target cursor --root /repo
```

### Homebrew

Coming soon:

```sh
brew install ars-standard/tap/ars
```

## Quick Start

```sh
ars init

# Edit canonical repository knowledge:
# .ai/manifest.yaml
# .ai/instructions/
# .ai/agents/
# .ai/skills/
# .ai/prompts/

ars validate
ars compose --target cursor
```

## Command Reference

| Command                                             | Description                               |
| --------------------------------------------------- | ----------------------------------------- |
| `ars init [--root <path>] [--force]`                | Scaffold `.ai/`.                          |
| `ars validate [--root <path>] [--json]`             | Validate `.ai/` structure and references. |
| `ars compose --target <target> [--root <path>]`     | Generate provider artifacts from `.ai/`.  |
| `ars import <source> [--root <path>] [--overwrite]` | Import provider artifacts into `.ai/`.    |

## Provider Support

| Provider       | Compose target | Output                               |
| -------------- | -------------- | ------------------------------------ |
| Cursor         | `cursor`       | `.cursor/rules/`, `.cursor/prompts/` |
| GitHub Copilot | `copilot`      | `.github/copilot-instructions.md`    |
| Claude Code    | `claude`       | `CLAUDE.md`                          |
| OpenAI Codex   | `codex`        | `AGENTS.md`                          |

| Provider artifact                 | Import source |
| --------------------------------- | ------------- |
| `.github/copilot-instructions.md` | `github`      |
| `.cursor/rules/*.mdc`             | `cursor`      |
| `CLAUDE.md`                       | `claude`      |

## Architecture

```text
.ai/
  manifest.yaml
  instructions/
  agents/
  skills/
  prompts/
      |
      v
   ars compose
      |
      +--> .cursor/
      +--> .github/copilot-instructions.md
      +--> CLAUDE.md
      +--> AGENTS.md
```

ARES is a local, file-based CLI. It is not an agent runtime, provider router, workflow engine, memory system, database-backed app, web app, or marketplace.

## Repository Format

```text
.ai/
  manifest.yaml                 project metadata
  instructions/<name>.md         repository-wide instructions
  agents/<name>/AGENT.md         agent role, responsibilities, uses, boundaries
  skills/<name>/SKILL.md         reusable knowledge
  prompts/<name>.md              reusable prompt templates
```

See [SPEC.md](SPEC.md) for the full ARS v1 specification.

## Contributing

Read [SPEC.md](SPEC.md), [docs/architecture.md](docs/architecture.md), and [docs/PLAN.md](docs/PLAN.md) before changing behavior. Keep `.ai/` as the canonical source of repository knowledge and provider files as generated artifacts.
