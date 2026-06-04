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
