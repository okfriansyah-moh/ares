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
