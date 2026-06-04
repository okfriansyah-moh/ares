# ARES Mission

## Purpose
ARES is the reference implementation of ARS, the AI Repository Standard.
Its job is to make repository knowledge portable across AI coding tools by treating `.ai/` as the canonical source of truth.

## The Problem
Repository knowledge is currently scattered across provider-specific conventions such as `.github/`, `.cursor/`, `CLAUDE.md`, `AGENTS.md`, and similar tool-owned files.
That fragmentation creates repository chaos.
When a team moves from Copilot to Cursor, Cursor to Claude, or Claude to Codex, they often have to rewrite repository conventions instead of carrying intent forward.

## Why `.ai/` Exists
`.ai/` exists to give the repository one human-first, AI-readable, git-friendly, provider-agnostic knowledge layer.
The repository should express its durable knowledge once in markdown-first source files.
Provider-specific artifacts are derived from that source, not treated as independent systems of record.

## What Belongs In `.ai/`
`.ai/` contains the portable knowledge that should survive provider changes:
- `manifest.yaml` for minimal repository identity and defaults
- `instructions/` for repository-wide guidance
- `agents/` for thin role definitions
- `skills/` for reusable knowledge
- `prompts/` for reusable task templates

These categories are in scope because they already recur across Cursor, Copilot, Claude, Codex, and related tools, but today appear under inconsistent conventions.

## What ARS Owns
ARS owns repository knowledge:
- Repository instructions
- Repository agents
- Repository skills
- Repository prompts
- The mapping from canonical `.ai/` knowledge to provider conventions

## What ARS Does Not Own
ARS does not own:
- Model selection
- Provider routing
- Agent execution
- Inference
- Billing
- API keys
- Token usage
- Workflow runtime or orchestration
- Memory systems
- Shared registries or marketplaces

ARES should feel closer to OpenAPI or Terraform than to an agent runtime.
The standard is the product. Tooling exists to support the standard, not replace it.

## Portability Goal
The success case is straightforward:
repository knowledge can move into `.ai/`, remain understandable there, and be composed into provider conventions for Cursor, Copilot, Claude, Codex, and future tools without losing intent.

In practical terms, teams should be able to:
- import provider-owned repository knowledge into `.ai/`
- edit `.ai/` as the canonical layer
- compose provider-specific artifacts from the same source

## Scope Discipline
ARES should stay focused on the repository knowledge layer.
If it grows into a runtime, orchestration engine, marketplace, or execution framework, it stops solving the fragmentation problem that motivated ARS in the first place.
