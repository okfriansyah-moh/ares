# ADR-0002 — Markdown as the Primary Source Format for `.ai/`

**Status:** Accepted
**Date:** 2026-06-04

## Context

`.ai/` files must be readable and editable by humans without tooling, renderable by every AI coding tool, and composable into provider-specific artifacts. The format choice directly affects adoption: teams must be willing to hand-author `.ai/` files.

Three formats were evaluated: Markdown, YAML/JSON, and a hybrid (YAML front matter + Markdown body).

## Decision

Use pure Markdown as the source format for `instructions/`, `agents/`, `skills/`, and `prompts/`. Use YAML only for `manifest.yaml`, where machine-readable structured data is required.

## Alternatives

| Option | Why not chosen |
|---|---|
| Pure YAML / JSON | Machine-readable but not human-pleasant for prose content; adds schema validation burden; providers render it poorly |
| YAML front matter + Markdown body | Couples the format to specific parsers; the front matter is useful only if tooling must parse per-file metadata, which v1 does not need |
| Single YAML schema for all files | Maximally structured but makes hand-authoring verbose and error-prone |

## Tradeoffs

**Gained:**
- Every AI coding tool can read and render Markdown natively
- Any text editor can author `.ai/` files without tooling
- Git diffs are readable
- No schema enforcement means no schema maintenance burden in v1
- Provider composers can extract sections by heading without a formal schema

**Given up:**
- Structural validation must be done by convention (required headings) rather than by schema
- Cross-file references (skill links in agents) must be parsed from free-form text rather than structured fields
- No machine-generated IDs — names are inferred from directory names

## Consequences

- `agents/<name>/AGENT.md` — `<name>` is the agent ID (directory name)
- `skills/<name>/SKILL.md` — `<name>` is the skill ID (directory name)
- Skill references in `AGENT.md` are written as path strings under `## Uses`
- The validator extracts required sections by heading; missing headings are reported as errors
- The composer uses goldmark AST to extract section content by heading
