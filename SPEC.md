# ARS v1 Specification

## 1. What Is ARS

ARS, the AI Repository Standard, defines a portable repository knowledge layer for AI-assisted development tools.

ARES is the reference implementation. It stores durable knowledge in `.ai/` and composes provider-specific artifacts for Cursor, GitHub Copilot, Claude Code, and OpenAI Codex.

The problem ARS solves is convention fragmentation: each AI coding tool expects instructions in a different location and format. ARS lets teams write repository knowledge once, then generate provider artifacts from the same source.

In scope for v1:

- `.ai/` repository knowledge format
- `ars init`
- `ars validate`
- `ars compose`
- `ars import`
- Compose targets: `cursor`, `copilot`, `claude`, `codex`
- Import sources: `github`, `cursor`, `claude`

Non-goals for v1:

- Agent execution or orchestration
- Provider API integration
- Model routing
- Runtime memory
- Databases
- Web or TUI frontend
- Skill marketplace or registry

The golden rule: delete generated provider artifacts, run `ars compose`, and the provider files come back from `.ai/`.

## 2. `.ai/` File Format

The `.ai/` directory is the canonical source of repository knowledge.

```text
.ai/
  manifest.yaml
  instructions/
  agents/
  skills/
  prompts/
```

### `manifest.yaml`

```yaml
version: "2.0"
project:
  name: string
  description: string
  repository: string
defaults:
  agent: string
```

Fields:

| Field | Required | Description |
|---|---:|---|
| `version` | Yes | ARS schema version. v1 recognizes `"2.0"`. |
| `project.name` | Yes | Repository or project name. |
| `project.description` | No | One-line project description. |
| `project.repository` | No | Canonical repository URL. |
| `defaults.agent` | No | Default agent ID used by compose targets. |

### `agents/<name>/AGENT.md`

Each agent has one `AGENT.md`. The directory name is the agent ID.

Required sections:

| Section | Required | Description |
|---|---:|---|
| `## Role` | Yes | One sentence describing ownership. |
| `## Responsibilities` | Yes | Bullet list of responsibilities. |
| `## Uses` | Yes | Skill references, usually `skills/<name>/SKILL.md`. |
| `## Boundaries` | Yes | What the agent does not do. |

### `skills/<name>/SKILL.md`

Each skill has one `SKILL.md`. The directory name is the skill ID. The file is free-form Markdown and may reference supplementary files under `references/`.

### `instructions/<name>.md`

Repository-wide instructions are free-form Markdown. The filename stem is the instruction ID.

### `prompts/<name>.md`

Prompt templates are Markdown. Recommended sections:

| Section | Description |
|---|---|
| `## Use` | One sentence goal. |
| `## Inputs` | Required context. |
| `## Instructions` | Steps to perform. |
| `## Check` | Validation criteria. |

## 3. `ars` CLI

All commands are local and file-based. v1 makes no network calls.

| Command | Purpose |
|---|---|
| `ars init [--root <path>] [--force]` | Scaffold `.ai/`. |
| `ars validate [--root <path>] [--json]` | Validate `.ai/`. |
| `ars compose --target <target> [--root <path>]` | Generate provider artifacts from `.ai/`. |
| `ars import <source> [--root <path>] [--overwrite]` | Convert provider artifacts into `.ai/`. |

### Exit Codes

| Command | Exit code |
|---|---|
| `validate` | `0` when no Error-level findings exist; `1` when any Error-level finding exists. |
| Other commands | `0` on success; nonzero on command error. |

Warnings do not fail validation.

### Validation Rules

| Rule | Level |
|---|---|
| Missing `.ai/manifest.yaml` | Error |
| Unparseable `manifest.yaml` | Error |
| Empty `project.name` | Error |
| Unrecognized `version` | Warning |
| Unknown `defaults.agent` | Warning |
| Missing required agent section | Error |
| Skill reference not found | Error |
| Skill reference escapes root | Error |
| Prompt missing `## Use` | Warning |

## 4. Provider Mappings

| `.ai/` source | Cursor | Copilot | Claude | Codex |
|---|---|---|---|---|
| `instructions/*.md` | `.cursor/rules/*.mdc` with `type: always` | Top section of `.github/copilot-instructions.md` | Top section of `CLAUDE.md` | Top section of `AGENTS.md` |
| `agents/<n>/AGENT.md` | `.cursor/rules/<n>.mdc` with `type: agent-requested` | `## Agent: <n>` block | `## <n>` section | Agent entry in `AGENTS.md` |
| `skills/<n>/SKILL.md` | Inlined into referencing agent rule | Inlined under relevant agent content | Inlined under agent context | Inlined under agent context |
| `prompts/<n>.md` | `.cursor/prompts/<n>.prompt` | Not supported in v1 | Custom slash command stub | Not supported in v1 |
| `manifest.yaml project.name` | Header comment | H1/header comment | H1 title | H1 title |

Mapping requirements:

- Preserve semantic intent.
- Produce deterministic output.
- Include a source marker in generated files.
- Regenerate complete managed target artifacts rather than partially updating them.
- `ars compose --target codex` creates `AGENTS.md` only when it is missing; an existing root `AGENTS.md` is path-validated and preserved.
- Keep generated artifacts traceable to `.ai/`.

## 5. Versioning

ARS uses the `manifest.yaml` `version` field to identify the `.ai/` schema.

Versioning rules:

- Compatible validation changes may add warnings.
- Structural changes require a version bump.
- Compose and import behavior must preserve existing v1 repositories.
- New provider targets or import sources must not change existing mappings.
- Existing commands must remain stable unless the manifest version changes.

## 6. Extension Guide

### Add A Compose Target

1. Add `internal/compose/<provider>.go`.
2. Implement `arslib.Composer`.
3. Register the composer in the compose registry.
4. Document the mapping in this spec.
5. Add unit and integration tests.

### Add An Import Source

1. Add `internal/importer/<source>.go`.
2. Implement `arslib.Importer`.
3. Register the importer in the import registry.
4. Reuse shared section classification when possible.
5. Add unit and integration tests.

### Add A `.ai/` Category

1. Update `manifest.yaml` semantics.
2. Bump the schema version.
3. Update validation.
4. Update all compose targets.
5. Update import behavior where applicable.
6. Update this spec.
