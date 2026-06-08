<!-- ars:source .ai/ -->
# ares

## Repository Instructions

<!-- ars:source .ai/instructions/architecture.md -->
# Repository Knowledge Architecture

## Purpose
Define the structure and ownership model of `.ai/`.
See `mission.md` for the problem statement, product scope, and portability rationale.

## Principles
- Markdown is the source of truth. `manifest.yaml` is the only required configuration file.
- Prefer convention over configuration. Choose obvious file names and locations over extra metadata.
- Keep file count low and information density high.
- A single opened file should explain its topic without requiring a chain of lookups.

## Structure
- `manifest.yaml` stores minimal repository identity and defaults.
- `instructions/` holds repository-wide guidance.
- `agents/<name>/AGENT.md` defines a thin role and points to the skills it uses.
- `skills/<name>/SKILL.md` is the authoritative source for reusable knowledge.
- `prompts/` contains reusable prompts that invoke skills without duplicating them.

## Ownership
- `instructions/` own repository-level rules and ARES-specific guidance.
- `agents/` own role boundaries and skill references.
- `skills/` own reusable methods, checklists, and output expectations.
- `prompts/` own reusable request templates.

## Change Rules
- Put durable knowledge in the narrowest file that owns it.
- Prefer updating an existing file over creating a new one.
- Add `references/` under a skill only when it removes real duplication.
- Keep agents thin, skills universal, and prompts reusable.

<!-- ars:source .ai/instructions/contribution-guidelines.md -->
# Contribution Guidelines

## How To Change `.ai/`
1. Find the file that already owns the knowledge.
2. Update that file directly.
3. Prefer refining existing guidance over adding new files.
4. Keep the result concise enough to scan quickly.

## Ownership Rules
- `instructions/mission.md` owns ARES-specific purpose, scope, and portability rationale.
- `instructions/architecture.md` owns `.ai/` structure and file ownership.
- `instructions/` owns repository-wide rules.
- `skills/` own reusable methods, checklists, and output expectations.
- `agents/` define roles, boundaries, and skill references.
- `prompts/` provide reusable request templates and should stay lean.

## Quality Bar
- No placeholders, TODOs, or filler text.
- No provider-specific wording unless a file is explicitly provider-specific.
- No duplicated guidance across instructions, agents, skills, and prompts.
- No metadata files unless they provide clear value that markdown cannot.

## Review Checklist
- Is this the smallest useful change?
- Can a human understand the topic from this one file?
- Does the guidance stay tool-agnostic?
- Is any repeated text better replaced by a skill reference?

<!-- ars:source .ai/instructions/mission.md -->
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

## Codex Skills

- .agents/skills/architecture-management/SKILL.md
- .agents/skills/plan-management/SKILL.md
- .agents/skills/task-implementation/SKILL.md
- .agents/skills/task-review/SKILL.md

## architect

<!-- ars:source .ai/agents/architect/AGENT.md -->
# Architect Agent

## Role
Own repository and system architecture decisions.

## Responsibilities
- Create or update architecture documents.
- Review structure, boundaries, dependencies, and drift.
- Write or update ADRs for material decisions.
- Surface tradeoffs before recommending complexity.

## Uses
- `.ai/skills/architecture-management/SKILL.md`

## Boundaries
- Do not make detailed implementation plans unless architecture is already settled.
- Do not duplicate skill guidance in this file.

### Skills
- .agents/skills/architecture-management/SKILL.md

### Subagent
- .codex/agents/architect.toml

## planner

<!-- ars:source .ai/agents/planner/AGENT.md -->
# Planner Agent

## Role
Turn approved requirements into sequenced implementation plans.

## Responsibilities
- Create task-based plans from specs or requests.
- Review plans for scope, order, and executability.
- Update plans without breaking existing numbering or flow.
- Flag architectural gaps that need escalation.

## Uses
- `.ai/skills/plan-management/SKILL.md`

## Boundaries
- Do not make architectural decisions that belong to the Architect.
- Do not duplicate skill guidance in this file.

### Skills
- .agents/skills/plan-management/SKILL.md

### Subagent
- .codex/agents/planner.toml

