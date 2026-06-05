# Purpose

Review a completed implementation task for compliance with the plan, architecture, tests, security expectations, complexity budget, and release readiness.

# Inputs

- `docs/architecture.md` for system boundaries, provider mappings, security requirements, and release rules.
- `docs/PLAN.md` for the selected task, dependencies, files, validation commands, and definition of done.
- The implementation diff, relevant code, tests, generated artifacts, and command output.
- Current `git status` to separate task changes from unrelated work.

# Review Checklist

- PLAN compliance: verify only the selected task was implemented, required files are complete, future tasks were not pulled forward, and task-specific validation was run.
- Architecture compliance: verify package boundaries, interfaces, local file-based behavior, deterministic compose/import output, and `.ai/` source-of-truth rules are preserved.
- Test coverage: verify focused unit or integration tests cover new behavior, edge cases, errors, and regressions; confirm `go test -race -count=1 ./...` passes.
- Security: verify path handling, symlink behavior, atomic writes, dependency vulnerability checks, and absence of unnecessary network, runtime, or privilege-expanding behavior.
- Complexity: verify KISS, YAGNI, DRY, narrow interfaces, no speculative abstractions, no global mutable state, and complexity within the plan budget.
- Release readiness: verify build, vet, staticcheck, tests, `govulncheck`, CLI behavior, container expectations, and versioning impact as relevant to the task.

# Findings Format

- Lead with findings ordered by severity.
- Include file and line references where possible.
- State the violated plan, architecture, test, security, complexity, or release requirement.
- Describe the user-visible or maintenance risk.
- Keep summaries brief and secondary to findings.
- If no issues are found, say so clearly and list any residual risk or commands not run.

# Anti-Patterns

- Reviewing only for style while missing plan or architecture drift.
- Treating future-task implementation as a bonus.
- Accepting untested behavior because the change is small.
- Ignoring security checks around filesystem access and generated artifacts.
- Recommending broad refactors unrelated to the selected task.
- Hiding failed or skipped validation commands.
- Mixing unrelated dirty worktree changes into the review result.
