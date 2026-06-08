# Purpose

Implement one planned task at a time from `docs/PLAN.md` while preserving architecture, scope, and validation discipline.

# Inputs

- `docs/architecture.md` for system boundaries and non-negotiable design decisions.
- `docs/PLAN.md` for task order, dependencies, required files, coding standards, and validation steps.
- Relevant `.ai/instructions/`, agents, skills, and prompts for repository-specific guidance.
- Current code, tests, and git status for implementation context.

# Process

- Read the full target task in `docs/PLAN.md`, including its goal, files, coding standards, validation, and prompt context.
- Check the dependency graph and task summary before starting; only work on the selected task and its direct prerequisites if they are missing.
- Treat the task boundary as the listed goal plus listed files and validation criteria. Do not add future task behavior, new architecture, or speculative abstractions.
- Inspect existing package patterns before editing. Keep changes small, idiomatic, and aligned with current boundaries.
- Implement production code and focused tests together. Avoid placeholders, TODOs, or stubs unless the plan explicitly calls for temporary scaffolding.
- If the task reveals a plan gap, make the smallest implementation-compatible choice and document the issue for review rather than expanding scope.

# Validation

- Run the task-specific validation commands from `docs/PLAN.md`.
- Before finishing a task, run:

```sh
go build ./...
go vet ./...
staticcheck ./...
go test -race -count=1 ./...
govulncheck ./...
```

- Confirm all required files were created or updated with non-stub implementation.
- Confirm no future task files or behaviors were introduced.
- Confirm failures are fixed before continuing, or clearly report any unavailable tool or external blocker.

# Anti-Patterns

- Implementing multiple plan tasks in one pass.
- Adding runtime, network, database, frontend, plugin, or marketplace behavior in v1.
- Redesigning architecture to make the current task easier.
- Creating abstractions before repeated complexity exists.
- Skipping race tests, vet, staticcheck, or vulnerability checks after implementation.
- Preparing a commit with unrelated changes, generated noise, failing tests, or unreviewed scope drift.
- Committing without checking `git status`, reviewing the diff, and using a task-scoped commit message.
