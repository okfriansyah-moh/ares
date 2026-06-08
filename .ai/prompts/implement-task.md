# Implement Task

## Use

Implement Task {{TASK_NUMBER}} from `docs/PLAN.md`.

## Inputs

- `AGENTS.md`
- `docs/architecture.md`
- `docs/PLAN.md`
- Task number: `{{TASK_NUMBER}}`

## Instructions

1. Read `AGENTS.md`.
2. Read `docs/architecture.md`.
3. Read `docs/PLAN.md`, focusing on Task {{TASK_NUMBER}}, its dependencies, validation, and referenced context.
4. Implement only Task {{TASK_NUMBER}}. Do not implement future tasks or redesign architecture.
5. Run the task validation and the repo validation commands required by `AGENTS.md` and `docs/PLAN.md`.
6. Report files changed, validation results, skipped commands, and any blocker.

## Check

- Only Task {{TASK_NUMBER}} changed.
- Required task files are complete and non-stub.
- Validation was run or skipped with a clear reason.
- No future task scope was added.
