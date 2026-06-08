# Implement And Review Task

## Use

Implement, self-review, fix, and report Task {{TASK_NUMBER}} from `docs/PLAN.md` in one session.

## Inputs

- `AGENTS.md`
- `docs/architecture.md`
- `docs/PLAN.md`
- `.ai/skills/task-implementation/SKILL.md`
- `.ai/skills/task-review/SKILL.md`
- `.ai/prompts/pr-remediation.md`
- Task number: `{{TASK_NUMBER}}`

## Instructions

1. Read the inputs. Focus only on Task {{TASK_NUMBER}}.
2. Implement Task {{TASK_NUMBER}} using `.ai/skills/task-implementation/SKILL.md`.
3. Self-review using `.ai/skills/task-review/SKILL.md`: PLAN compliance, architecture, tests, security, complexity, release readiness.
4. For each finding, apply `.ai/prompts/pr-remediation.md`: terse classification, exact fix, no filler.
5. Fix findings immediately when they are in Task {{TASK_NUMBER}} scope. Do not implement future tasks.
6. Run task validation and required repo validation.
7. Report changed files, fixes made, validation results, skipped commands, and blockers.
8. Mark Task {{TASK_NUMBER}} Completed with checkmark in `docs/PLAN.md`

## Check

- Single task only.
- Findings fixed or explicitly classified out of scope.
- Validation run after fixes.
- Output is concise and action-focused.
