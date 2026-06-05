# AGENTS.md - Codex Instructions for ARES

## Repository Purpose

ARES is the reference implementation of ARS, the AI Repository Standard. It provides the `ars` CLI, a single static Go binary that makes repository knowledge portable across AI coding tools by treating `.ai/` as the canonical knowledge layer and composing provider-specific artifacts such as `.cursor/`, `.github/copilot-instructions.md`, `CLAUDE.md`, and `AGENTS.md`.

ARES is a repository knowledge standard and CLI tool. It is not an agent runtime, provider router, workflow engine, memory system, marketplace, API layer, database-backed app, web app, or TUI.

## Knowledge Hierarchy

When instructions conflict, follow this order:

1. `docs/architecture.md` - architectural source of truth and scope boundaries.
2. `docs/PLAN.md` - implementation task order, acceptance criteria, and validation commands.
3. `.ai/instructions/` - repository-wide durable guidance.
4. `.ai/agents/` - role boundaries and skill references.
5. `.ai/skills/` - reusable methods, checklists, and expectations.
6. `.ai/prompts/` - reusable task templates.
7. Existing code and tests - current implementation patterns.

Before architecture, planning, implementation, review, or documentation decisions, consult the relevant files above. Do not redesign architecture or expand scope unless explicitly requested.

## Implementation Rules

- Follow `docs/PLAN.md` one task at a time; do not implement future tasks opportunistically.
- Prefer the existing Go package boundaries: `cmd/` for CLI adapters, `internal/` for implementation packages, `pkg/arslib` for public domain types.
- Keep `cmd/ars` thin: parse flags, call internal packages, format output, return errors.
- Keep operations local and file-based. No database, frontend, network calls, provider APIs, runtime orchestration, or plugin system in v1.
- Preserve deterministic output: sort file walks and make compose/import output idempotent.
- Prefer simple, explicit implementations. Apply KISS, YAGNI, DRY, and narrow interfaces.
- Avoid speculative abstractions and global mutable state.
- Do not add placeholders, TODOs, filler text, or duplicated guidance.
- After Task 13, all file I/O must go through `internal/safepath`; direct `os.ReadFile`, `os.WriteFile`, and `filepath.WalkDir` are forbidden outside that package.
- Do not call `os.Exit` outside `main`/Cobra process handling.
- Do not use `fmt.Println` in non-`cmd/` packages; return errors or use appropriate logging.
- Treat generated provider artifacts as derived from `.ai/`; durable repository knowledge belongs in `.ai/`.

## Testing Rules

After each implementation task, run:

```sh
go build ./...
go vet ./...
staticcheck ./...
go test -race -count=1 ./...
govulncheck ./...
```

Use focused tests for the changed package and broaden to `./...` before finishing. For CLI behavior, include integration-style tests around Cobra execution and temporary repositories. For security-sensitive file work, test path traversal, symlink rejection, atomic writes, and deterministic ordering.

Validation must finish with zero build errors, vet issues, staticcheck issues, test failures, data races, and vulnerability findings unless the user explicitly narrows the task and the limitation is reported.

## Review Rules

- Review against `docs/architecture.md` first, then `docs/PLAN.md`.
- Lead with correctness, security, behavioral regressions, missing tests, and scope drift.
- Confirm changes are the smallest useful change and remain tool-agnostic unless the target file is provider-specific.
- Check `.ai/` ownership: durable knowledge in the narrowest owning file, agents thin, skills reusable, prompts lean.
- Check generated artifacts are traceable to `.ai/`, deterministic, overwrite-safe, and free of orphaned output.
- For imports, check classification heuristics are transparent and shared instead of duplicated.
- For path handling, check root escape prevention, symlink handling, and atomic writes.

## Release Rules

- Build the release binary with `CGO_ENABLED=0`, `-trimpath`, and stripped symbols.
- Container releases use a two-stage build and `gcr.io/distroless/static-debian12:nonroot`; the final image must have no shell, no package manager, and run as `nonroot:nonroot`.
- Run `govulncheck ./...` before every release; zero findings are required.
- Release artifacts should include cross-platform binaries, immutable version tags, a mutable `latest` container tag, checksums, signatures, and an SBOM.
- Release workflows must verify `ars --version`, container startup, static linking, and non-root execution.
