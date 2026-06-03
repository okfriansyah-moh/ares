![ARES AI Repository Standard](./ares-infographic.png)

# ARES — AI Repository Standard

A provider-agnostic standard for AI-assisted development repositories.

Define repository knowledge once in `.ai/`, then generate provider-specific conventions for Cursor, GitHub Copilot, Claude, Codex, and future tools.

## What ARES Is

- A repository standard for AI-assisted development.
- A canonical source of truth in `.ai/`.
- A portability layer for provider conventions.

## What ARES Is Not

- Not an agent runtime.
- Not an orchestration engine.
- Not a memory system.
- Not a provider API integration.

## Core Structure

```
.ai/
  manifest.yaml
  instructions/
  agents/
  skills/
  prompts/
```

Everything else is generated from this source.

## Repository Layout

- `.ai/`
  - `manifest.yaml` — repository metadata and defaults.
  - `instructions/` — repository-wide rules and policies.
  - `agents/` — role definitions and agent instructions.
  - `skills/` — reusable knowledge units and best practices.
  - `prompts/` — reusable execution units and templates.

## Goals

1. Solve provider convention fragmentation.
2. Keep repository knowledge centralized.
3. Enable simple migration between providers.
4. Keep the standard small and practical.

## Command Concepts

- `ars init` — initialize the `.ai/` repository structure.
- `ars validate` — verify required files, metadata, and structure.
- `ars compose` — generate provider artifacts from `.ai/`.
- `ars import` — convert existing provider-specific conventions into `.ai/`.

## Compose Targets

- `cursor` → generates `.cursor/`
- `copilot` → generates `.github/`
- `claude` → generates `CLAUDE.md`
- `codex` → generates `AGENTS.md`

## Why ARES Works

- It keeps knowledge in one place.
- It protects repositories from tool-specific lock-in.
- It supports future provider additions without rewriting repository conventions.

## Example Workflow

```bash
ars import github
ars compose --target cursor
```

This moves repository conventions from GitHub Copilot into the `.ai/` standard, then generates a Cursor-compatible output.

## Success Criteria

If the repository can move from provider-specific artifacts into `.ai/` and back again without losing intent, ARES is providing value.

## Notes

The canonical repository source is `.ai/`. All other provider artifacts are derived.

## License

This repository is covered by the `LICENSE` file in the root.
